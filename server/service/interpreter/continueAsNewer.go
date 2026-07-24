// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package interpreter

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/superdurable/iwf/config"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type ContinueAsNewer struct {
	provider interfaces.WorkflowProvider

	StateExecutionToResumeMap map[string]service.StateExecutionResumeInfo // stateExeId to StateExecutionResumeInfo
	inflightUpdateOperations  int

	stateRequestQueue     *StateRequestQueue
	interStateChannel     *InternalChannel
	stateExecutionCounter *StateExecutionCounter
	persistenceManager    *PersistenceManager
	signalReceiver        *SignalReceiver
	outputCollector       *OutputCollector
	timerProcessor        interfaces.TimerProcessor
}

func NewContinueAsNewer(
	provider interfaces.WorkflowProvider,
	interStateChannel *InternalChannel, signalReceiver *SignalReceiver, stateExecutionCounter *StateExecutionCounter,
	persistenceManager *PersistenceManager, stateRequestQueue *StateRequestQueue, collector *OutputCollector,
	timerProcessor interfaces.TimerProcessor,
) *ContinueAsNewer {
	return &ContinueAsNewer{
		provider: provider,

		StateExecutionToResumeMap: map[string]service.StateExecutionResumeInfo{},

		stateRequestQueue:     stateRequestQueue,
		interStateChannel:     interStateChannel,
		signalReceiver:        signalReceiver,
		stateExecutionCounter: stateExecutionCounter,
		persistenceManager:    persistenceManager,
		outputCollector:       collector,
		timerProcessor:        timerProcessor,
	}
}

func LoadInternalsFromPreviousRun(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	activityCfg *config.InterpreterActivityConfig,
	previousRunId string,
	continueAsNewPageSizeInBytes int32,
) (*service.ContinueAsNewDumpResponse, error) {
	if activityCfg == nil {
		panic("LoadInternalsFromPreviousRun requires activity config")
	}
	activityOptions := interfaces.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &iwfidl.RetryPolicy{
			MaximumIntervalSeconds: iwfidl.PtrInt32(5),
		},
	}
	if activityCfg.DumpWorkflowInternalActivityConfig != nil {
		activityConfig := activityCfg.DumpWorkflowInternalActivityConfig
		activityOptions.StartToCloseTimeout = activityConfig.StartToCloseTimeout
		if activityConfig.RetryPolicy != nil {
			activityOptions.RetryPolicy = activityConfig.RetryPolicy
		}
	}

	ctx = provider.WithActivityOptions(ctx, activityOptions)
	workflowId := provider.GetWorkflowInfo(ctx).WorkflowExecution.ID
	pageSize := continueAsNewPageSizeInBytes
	if pageSize == 0 {
		pageSize = service.DefaultContinueAsNewPageSizeInBytes
	}
	var sb strings.Builder
	lastChecksum := ""
	pageNum := int32(0)
	for {
		var resp iwfidl.WorkflowDumpResponse
		err := provider.ExecuteActivity(&resp, false, ctx, DumpWorkflowInternal, provider.GetBackendType(),
			iwfidl.WorkflowDumpRequest{
				WorkflowId:      workflowId,
				WorkflowRunId:   previousRunId,
				PageNum:         pageNum,
				PageSizeInBytes: pageSize,
			})
		if err != nil {
			return nil, err
		}
		if lastChecksum != "" && lastChecksum != resp.Checksum {
			// reset to start from beginning
			pageNum = 0
			sb.Reset()
			provider.GetLogger(ctx).Error("checksum has changed during the loading", lastChecksum, resp.Checksum)
			lastChecksum = ""
			continue
		} else {
			lastChecksum = resp.Checksum
			sb.WriteString(resp.JsonData)
			pageNum++
			if pageNum >= resp.TotalPages {
				break
			}
		}
	}

	var resp service.ContinueAsNewDumpResponse
	err := json.Unmarshal([]byte(sb.String()), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *ContinueAsNewer) GetSnapshot() service.ContinueAsNewDumpResponse {
	localStateExecutionToResumeMap := map[string]service.StateExecutionResumeInfo{}
	for _, key := range DeterministicKeys(c.StateExecutionToResumeMap) {
		localStateExecutionToResumeMap[key] = c.StateExecutionToResumeMap[key]
	}
	for _, value := range c.stateRequestQueue.GetAllStateResumeRequests() {
		localStateExecutionToResumeMap[value.StateExecutionId] = value
	}
	return service.ContinueAsNewDumpResponse{
		InterStateChannelReceived:  c.interStateChannel.GetAllReceived(),
		SignalsReceived:            c.signalReceiver.GetAllReceived(),
		StateExecutionCounterInfo:  c.stateExecutionCounter.Dump(),
		DataObjects:                c.persistenceManager.GetAllDataAttributes(),
		SearchAttributes:           c.persistenceManager.GetAllSearchAttributes(),
		StatesToStartFromBeginning: c.stateRequestQueue.GetAllStateStartRequests(),
		StateExecutionsToResume:    localStateExecutionToResumeMap,
		StateOutputs:               c.outputCollector.GetAll(),
		StaleSkipTimerSignals:      c.timerProcessor.Dump(),
	}
}

func (c *ContinueAsNewer) SetQueryHandlersForContinueAsNew(ctx interfaces.UnifiedContext) error {
	return c.provider.SetQueryHandler(ctx, service.ContinueAsNewDumpByPageQueryType,
		// return the current page of the whole snapshot
		func(request iwfidl.WorkflowDumpRequest) (*iwfidl.WorkflowDumpResponse, error) {
			wholeSnapshot := c.GetSnapshot()
			wholeData, err := json.Marshal(wholeSnapshot)
			if err != nil {
				return nil, err
			}
			checksum := md5.Sum(wholeData)
			pageSize := int32(service.DefaultContinueAsNewPageSizeInBytes)
			if request.PageSizeInBytes > 0 {
				pageSize = request.PageSizeInBytes
			}
			lenInDouble := float64(len(wholeData))
			totalPages := int32(math.Ceil(lenInDouble / float64(pageSize)))
			if request.PageNum >= totalPages {
				return nil, fmt.Errorf("wrong pageNum, request %v but max is %v , shouldn't happen", request.PageNum, totalPages-1)
			}
			start := pageSize * request.PageNum
			end := start + pageSize
			if end > int32(len(wholeData)) {
				end = int32(len(wholeData))
			}
			return &iwfidl.WorkflowDumpResponse{
				Checksum:   string(checksum[:]),
				TotalPages: totalPages,
				JsonData:   string(wholeData[start:end]),
			}, nil
		})
}

func (c *ContinueAsNewer) AddPotentialStateExecutionToResume(
	stateExecutionId string, state iwfidl.StateMovement, stateExecLocals []iwfidl.KeyValue,
	commandRequest iwfidl.CommandRequest,
	completedTimerCommands map[int]service.InternalTimerStatus,
	completedSignalCommands, completedInterStateChannelCommands map[int]*iwfidl.EncodedObject,
) {
	c.StateExecutionToResumeMap[stateExecutionId] = service.StateExecutionResumeInfo{
		StateExecutionId:     stateExecutionId,
		State:                state,
		StateExecutionLocals: stateExecLocals,
		CommandRequest:       commandRequest,
		StateExecutionCompletedCommands: service.StateExecutionCompletedCommands{
			CompletedTimerCommands:             completedTimerCommands,
			CompletedSignalCommands:            completedSignalCommands,
			CompletedInterStateChannelCommands: completedInterStateChannelCommands,
		},
	}
}

func (c *ContinueAsNewer) HasAnyStateExecutionToResume() bool {
	return len(c.StateExecutionToResumeMap) > 0
}
func (c *ContinueAsNewer) RemoveStateExecutionToResume(stateExecutionId string) {
	delete(c.StateExecutionToResumeMap, stateExecutionId)
}

func (c *ContinueAsNewer) DrainThreads(ctx interfaces.UnifiedContext) error {
	// TODO: add metric for before and after Await to monitor stuck
	// NOTE: consider using AwaitWithTimeout to get an alert when workflow stuck due to a bug in the draining logic for continueAsNew

	errWait := c.provider.Await(ctx, func() bool {
		return c.allThreadsDrained(ctx)
	})
	c.provider.GetLogger(ctx).Info("done draining threads for continueAsNew", errWait)

	return errWait
}

func (c *ContinueAsNewer) IncreaseInflightOperation() {
	c.inflightUpdateOperations++
}

func (c *ContinueAsNewer) DecreaseInflightOperation() {
	c.inflightUpdateOperations--
}

// if the DrainAllSignalsAndThreads await is being called more than a few times and cannot get through,
// there is likely something wrong in the continueAsNew logic (unless state API is stuck)
// the key is runId, the value is how many times it has been called in this worker
// Using this in memory counter sot hat we don't have to use AwaitWithTimeout which will consume a timer
// TODO add TTL support because we don't have to keep the value in memory forever(likely a few hours or a day is enough)
var inMemoryContinueAsNewMonitor = make(map[string]time.Time)

const warnThreshold = time.Second * 5
const errThreshold = time.Second * 15

func (c *ContinueAsNewer) allThreadsDrained(ctx interfaces.UnifiedContext) bool {
	runId := c.provider.GetWorkflowInfo(ctx).WorkflowExecution.RunID

	remainingThreadCount := c.provider.GetThreadCount()
	if remainingThreadCount == 0 && c.inflightUpdateOperations == 0 {
		delete(inMemoryContinueAsNewMonitor, runId)
		return true
	}

	c.provider.GetLogger(ctx).Debug("continueAsNew is in draining remainingThreadCount, attempt, threadNames, inflightUpdateOperations",
		remainingThreadCount, inMemoryContinueAsNewMonitor[runId], c.provider.GetPendingThreadNames(), c.inflightUpdateOperations)

	// TODO using a flag to control this debugging info
	initTime, ok := inMemoryContinueAsNewMonitor[runId]
	if !ok {
		inMemoryContinueAsNewMonitor[runId] = time.Now()
		return false
	}

	elapsed := time.Since(initTime)

	if elapsed >= errThreshold {
		c.provider.GetLogger(ctx).Warn(
			"continueAsNew is likely stuck (unless state API is stuck) in draining remainingThreadCount, attempt, threadNames, inflightUpdateOperations",
			remainingThreadCount, inMemoryContinueAsNewMonitor[runId], c.provider.GetPendingThreadNames(), c.inflightUpdateOperations)
		return false
	}
	if elapsed >= warnThreshold {
		c.provider.GetLogger(ctx).Warn(
			"continueAsNew may be stuck (unless state API is stuck) in draining remainingThreadCount, attempt, threadNames, inflightUpdateOperations",
			remainingThreadCount, inMemoryContinueAsNewMonitor[runId], c.provider.GetPendingThreadNames(), c.inflightUpdateOperations)
	}
	return false
}
