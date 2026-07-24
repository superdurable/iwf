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
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/event"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/superdurable/iwf/service/interpreter/cont"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
	"time"
)

type WorkflowUpdater struct {
	persistenceManager   *PersistenceManager
	provider             interfaces.WorkflowProvider
	continueAsNewer      *ContinueAsNewer
	continueAsNewCounter *cont.ContinueAsNewCounter
	internalChannel      *InternalChannel
	signalReceiver       *SignalReceiver
	stateRequestQueue    *StateRequestQueue
	logger               interfaces.UnifiedLogger
	basicInfo            service.BasicInfo
}

func NewWorkflowUpdater(
	ctx interfaces.UnifiedContext, provider interfaces.WorkflowProvider, persistenceManager *PersistenceManager,
	stateRequestQueue *StateRequestQueue,
	continueAsNewer *ContinueAsNewer, continueAsNewCounter *cont.ContinueAsNewCounter,
	internalChannel *InternalChannel, signalReceiver *SignalReceiver, basicInfo service.BasicInfo,
	globalVersioner *GlobalVersioner,
) (*WorkflowUpdater, error) {
	updater := &WorkflowUpdater{
		persistenceManager:   persistenceManager,
		continueAsNewer:      continueAsNewer,
		continueAsNewCounter: continueAsNewCounter,
		internalChannel:      internalChannel,
		signalReceiver:       signalReceiver,
		stateRequestQueue:    stateRequestQueue,
		basicInfo:            basicInfo,
		provider:             provider,
		logger:               provider.GetLogger(ctx),
	}
	if globalVersioner.IsAfterVersionOfTemporal26SDK() {
		err := provider.SetRpcUpdateHandler(ctx, service.ExecuteOptimisticLockingRpcUpdateType, updater.validator, updater.handler)
		if err != nil {
			return nil, err
		}
	}
	return updater, nil
}

func (u *WorkflowUpdater) handler(
	ctx interfaces.UnifiedContext, input iwfidl.WorkflowRpcRequest,
) (output *interfaces.HandlerOutput, err error) {
	u.continueAsNewer.IncreaseInflightOperation()
	defer u.continueAsNewer.DecreaseInflightOperation()

	info := u.provider.GetWorkflowInfo(ctx)

	rpcExecutionStartTime := u.provider.Now(ctx).UnixMilli()

	defer func() {
		if !u.provider.IsReplaying(ctx) {
			event.Handle(iwfidl.IwfEvent{
				EventType:          iwfidl.RPC_EXECUTION_EVENT,
				RpcName:            &input.RpcName,
				WorkflowType:       u.basicInfo.IwfWorkflowType,
				WorkflowId:         info.WorkflowExecution.ID,
				StartTimestampInMs: ptr.Any(rpcExecutionStartTime),
				SearchAttributes:   u.persistenceManager.GetAllSearchAttributes(),
			})
		}
	}()

	rpcPrep := service.PrepareRpcQueryResponse{
		DataObjects:              u.persistenceManager.LoadDataAttributes(ctx, input.DataAttributesLoadingPolicy),
		SearchAttributes:         u.persistenceManager.LoadSearchAttributes(ctx, input.SearchAttributesLoadingPolicy),
		WorkflowRunId:            info.WorkflowExecution.RunID,
		WorkflowStartedTimestamp: info.WorkflowStartTime.Unix(),
		IwfWorkflowType:          u.basicInfo.IwfWorkflowType,
		IwfWorkerUrl:             u.basicInfo.IwfWorkerUrl,
		SignalChannelInfo:        u.signalReceiver.GetInfos(),
		InternalChannelInfo:      u.internalChannel.GetInfos(),
	}

	activityOptions := interfaces.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &iwfidl.RetryPolicy{
			MaximumAttemptsDurationSeconds: input.TimeoutSeconds,
			MaximumAttempts:                iwfidl.PtrInt32(3),
		},
	}
	ctx = u.provider.WithActivityOptions(ctx, activityOptions)
	var activityOutput interfaces.InvokeRpcActivityOutput

	err = u.provider.ExecuteLocalActivity(
		&activityOutput,
		ctx,
		InvokeWorkerRpc,
		u.provider.GetBackendType(),
		&rpcPrep,
		input,
	)

	u.persistenceManager.UnlockPersistence(input.SearchAttributesLoadingPolicy, input.DataAttributesLoadingPolicy)

	if err != nil {
		return nil, u.provider.NewApplicationError(string(iwfidl.SERVER_INTERNAL_ERROR_TYPE), "activity invocation failure:"+err.Error())
	}

	handlerOutput := &interfaces.HandlerOutput{
		StatusError: activityOutput.StatusError,
	}

	rpcOutput := activityOutput.RpcOutput
	if rpcOutput != nil {
		handlerOutput.RpcOutput = &iwfidl.WorkflowRpcResponse{
			Output: rpcOutput.Output,
		}
		u.continueAsNewCounter.IncSyncUpdateReceived()
		u.persistenceManager.ProcessUpsertDataAttribute(rpcOutput.UpsertDataAttributes)
		_ = u.persistenceManager.ProcessUpsertSearchAttribute(ctx, rpcOutput.UpsertSearchAttributes)
		u.internalChannel.ProcessPublishing(rpcOutput.PublishToInterStateChannel)
		if rpcOutput.StateDecision != nil {
			u.stateRequestQueue.AddStateStartRequests(rpcOutput.StateDecision.NextStates)
		}
	}

	return handlerOutput, nil
}

func (u *WorkflowUpdater) validator(_ interfaces.UnifiedContext, input iwfidl.WorkflowRpcRequest) error {
	var daKeys, saKeys []string
	if input.HasDataAttributesLoadingPolicy() {
		daKeys = input.DataAttributesLoadingPolicy.LockingKeys
	}
	if input.HasSearchAttributesLoadingPolicy() {
		saKeys = input.SearchAttributesLoadingPolicy.LockingKeys
	}
	keysUnlocked := u.persistenceManager.CheckDataAndSearchAttributesKeysAreUnlocked(daKeys, saKeys)
	if keysUnlocked {
		return nil
	} else {
		return u.provider.NewApplicationError(string(iwfidl.RPC_ACQUIRE_LOCK_FAILURE), "requested data or search attributes are being locked by other operations")
	}
}
