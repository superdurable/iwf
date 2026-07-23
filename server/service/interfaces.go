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

package service

import (
	"github.com/superdurable/iwf/gen/iwfpb"
)

type (
	InterpreterWorkflowInput struct {
		FlowType string `json:"flowType,omitempty"`

		WorkerTarget string `json:"workerTarget,omitempty"`

		StartStepType string `json:"startStepType,omitempty"`

		WaitForCompletionStepExecutionIds []string `json:"waitForCompletionStepExecutionIds,omitempty"`
		WaitForCompletionStepTypes        []string `json:"waitForCompletionStepTypes,omitempty"`

		StepInput *iwfpb.Value `json:"stepInput,omitempty"`

		StepOptions *iwfpb.StepOptions `json:"stepOptions,omitempty"`

		InitAttributes []*iwfpb.AttributeWrite `json:"initAttributes,omitempty"`

		Config *iwfpb.FlowConfig `json:"config,omitempty"`

		// IsResumeFromContinueAsNew ignores StartStepType / StepInput / StepOptions / InitAttributes.
		IsResumeFromContinueAsNew bool `json:"isResumeFromContinueAsNew,omitempty"`

		ContinueAsNewInput *ContinueAsNewInput `json:"continueAsNewInput,omitempty"`
	}

	ContinueAsNewInput struct {
		PreviousInternalRunId string `json:"previousInternalRunId"`
	}

	InterpreterWorkflowOutput struct {
		StepCompletionOutputs []*iwfpb.StepCompletionOutput `json:"stepCompletionOutputs,omitempty"`
	}

	BasicInfo struct {
		FlowType     string `json:"flowType,omitempty"`
		WorkerTarget string `json:"workerTarget,omitempty"`
	}

	InvokeWaitForMethodActivityInput struct {
		WorkerTarget string
		Request      *iwfpb.InvokeWaitForMethodRequest
	}

	InvokeExecuteMethodActivityInput struct {
		WorkerTarget string
		Request      *iwfpb.InvokeExecuteMethodRequest
	}

	GetAttributesQueryRequest struct {
		Keys    []string
		AllKeys bool
	}

	GetAttributesQueryResponse struct {
		Attributes []*iwfpb.KV
	}

	PrepareRpcQueryRequest struct {
		LockAttributeKeys []string
	}

	PrepareRpcQueryResponse struct {
		Attributes           []*iwfpb.KV
		RunId                string
		FlowStartedTimestamp int64
		FlowType             string
		WorkerTarget         string
		ChannelInfos         map[string]*iwfpb.ChannelInfo
	}

	ExecuteRpcSignalRequest struct {
		RpcInput         *iwfpb.Value             `json:"rpcInput,omitempty"`
		RpcOutput        *iwfpb.Value             `json:"rpcOutput,omitempty"`
		UpsertAttributes []*iwfpb.AttributeWrite  `json:"upsertAttributes,omitempty"`
		StepDecision     *iwfpb.StepDecision      `json:"stepDecision,omitempty"`
		RecordEvents     []*iwfpb.KV              `json:"recordEvents,omitempty"`
		PublishToChannel []*iwfpb.ChannelMessage  `json:"publishToChannel,omitempty"`
	}

	GetCurrentTimerInfosQueryResponse struct {
		StepExecutionCurrentTimerInfos map[string][]*TimerInfo // key is stepExecutionId
	}

	GetScheduledGreedyTimerTimesQueryResponse struct {
		PendingScheduled []*TimerInfo
	}

	TimerInfo struct {
		ConditionId                *string
		FiringUnixTimestampSeconds int64
		Status                     InternalTimerStatus
	}

	SkipTimerSignalRequest struct {
		StepExecutionId     string
		TimerConditionId    string
		TimerConditionIndex int
	}

	FailFlowSignalRequest struct {
		Reason string
	}

	CompleteFlowSignalRequest struct {
		Reason string
	}

	InternalTimerStatus string

	DebugDumpResponse struct {
		Config                     *iwfpb.FlowConfig
		Snapshot                   *iwfpb.ContinueAsNewDump
		FiringTimersUnixTimestamps []int64
	}
)

type StepExecutionStatus string

const FailureStepExecutionStatus StepExecutionStatus = "Failure"
const WaitingConditionsStepExecutionStatus StepExecutionStatus = "WaitingConditions"
const CompletedStepExecutionStatus StepExecutionStatus = "Completed"
const ExecuteApiFailedAndProceed StepExecutionStatus = "ExecuteApiFailedAndProceed"

// Legacy aliases used until Phase 4 renames all interpreter call sites.
const FailureStateExecutionStatus = FailureStepExecutionStatus
const WaitingCommandsStateExecutionStatus = WaitingConditionsStepExecutionStatus
const CompletedStateExecutionStatus = CompletedStepExecutionStatus

const (
	TimerPending InternalTimerStatus = "Pending"
	TimerFired   InternalTimerStatus = "Fired"
	TimerSkipped InternalTimerStatus = "Skipped"
)

// ValidateTimerSkipRequest validates if the skip timer request is valid.
// Prefer timerConditionId when non-empty; otherwise use timerIdx.
func ValidateTimerSkipRequest(
	stepExeTimerInfos map[string][]*TimerInfo, stepExeId, timerConditionId string, timerIdx int,
) (*TimerInfo, bool) {
	timerInfos := stepExeTimerInfos[stepExeId]
	if len(timerInfos) == 0 {
		return nil, false
	}
	if timerConditionId != "" {
		for _, t := range timerInfos {
			if t.ConditionId != nil && *t.ConditionId == timerConditionId {
				return t, true
			}
		}
		return nil, false
	}
	if timerIdx >= 0 && timerIdx < len(timerInfos) {
		t := timerInfos[timerIdx]
		if t.Status == TimerPending {
			return t, true
		}
	}
	return nil, false
}
