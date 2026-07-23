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

type (
	BackendType string
)

const (
	EnvNameDebugMode = "DEBUG_MODE"

	DefaultContinueAsNewPageSizeInBytes = 1024 * 1024

	TaskQueue = "Interpreter_DEFAULT"

	StateStartApi        = "/api/v1/workflowState/start"
	StateDecideApi       = "/api/v1/workflowState/decide"
	WorkflowWorkerRpcApi = "/api/v1/workflowWorker/rpc"

	GetDataAttributesWorkflowQueryType   = "GetDataAttributes"
	GetSearchAttributesWorkflowQueryType = "GetSearchAttributes"
	GetCurrentTimerInfosQueryType        = "GetCurrentTimerInfos"
	ContinueAsNewDumpByPageQueryType     = "ContinueAsNewDumpByPage"
	DebugDumpQueryType                   = "DebugNewDump"
	PrepareRpcQueryType                  = "PrepareRpcQueryType"

	ExecuteOptimisticLockingRpcUpdateType = "ExecuteOptimisticLockingRpcUpdate"

	SearchAttributeGlobalVersion     = "IwfGlobalWorkflowVersion"
	SearchAttributeExecutingStateIds = "IwfExecutingStateIds"
	SearchAttributeIwfWorkflowType   = "IwfWorkflowType"

	BackendTypeCadence  BackendType = "cadence"
	BackendTypeTemporal BackendType = "temporal"

	IwfSystemConstPrefix = "__IwfSystem_"

	SkipTimerSignalChannelName            = IwfSystemConstPrefix + "SkipTimerChannel"
	FailWorkflowSignalChannelName         = IwfSystemConstPrefix + "FailWorkflowChannel"
	UpdateConfigSignalChannelName         = IwfSystemConstPrefix + "UpdateWorkflowConfig"
	ExecuteRpcSignalChannelName           = IwfSystemConstPrefix + "ExecuteRpc"
	StateCompletionSignalChannelName      = IwfSystemConstPrefix + "StateCompletion"
	TriggerContinueAsNewSignalChannelName = IwfSystemConstPrefix + "TriggerContinueAsNew"

	WorkerUrlMemoKey            = IwfSystemConstPrefix + "WorkerUrl"
	UseMemoForDataAttributesKey = IwfSystemConstPrefix + "UseMemoForDataAttributes"
	WorkflowRequestId           = IwfSystemConstPrefix + "WorkflowRequestId"
)

var ValidIwfSystemSignalNames = map[string]bool{
	SkipTimerSignalChannelName:    true,
	FailWorkflowSignalChannelName: true,
	UpdateConfigSignalChannelName: true,
	ExecuteRpcSignalChannelName:   true,
}

const (
	GracefulCompletingWorkflowStateId = "_SYS_GRACEFUL_COMPLETING_WORKFLOW"
	ForceCompletingWorkflowStateId    = "_SYS_FORCE_COMPLETING_WORKFLOW"
	ForceFailingWorkflowStateId       = "_SYS_FORCE_FAILING_WORKFLOW"
	DeadEndWorkflowStateId            = "_SYS_DEAD_END"
)

var ValidClosingWorkflowStateId = map[string]bool{
	GracefulCompletingWorkflowStateId: true,
	ForceCompletingWorkflowStateId:    true,
	ForceFailingWorkflowStateId:       true,
	DeadEndWorkflowStateId:            true,
}
