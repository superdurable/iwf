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

package uclient

import (
	"context"
	"github.com/superdurable/iwf/service"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
)

type UnifiedClient interface {
	Close()
	errorHandler
	StartInterpreterWorkflow(
		ctx context.Context, options StartWorkflowOptions, args ...interface{},
	) (runId string, err error)
	StartWaitForStateCompletionWorkflow(ctx context.Context, options StartWorkflowOptions) (runId string, err error)
	StartBlobStoreCleanupWorkflow(
		ctx context.Context, taskQueue, workflowID, cronSchedule, storeId string,
	) error
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
	SignalWithStartWaitForStateCompletionWorkflow(
		ctx context.Context, options StartWorkflowOptions, stateCompletionOutput iwfidl.StateCompletionOutput,
	) error
	CancelWorkflow(ctx context.Context, workflowID string, runID string) error
	TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string) error
	ListWorkflow(ctx context.Context, request *ListWorkflowExecutionsRequest) (*ListWorkflowExecutionsResponse, error)
	QueryWorkflow(
		ctx context.Context, valuePtr interface{}, workflowID string, runID string, queryType string,
		args ...interface{},
	) error // TODO it doesn't return error correctly... the error is nil when query handler is not implemented
	DescribeWorkflowExecution(
		ctx context.Context, workflowID, runID string, requestedSearchAttributes []iwfidl.SearchAttributeKeyAndType,
	) (*DescribeWorkflowExecutionResponse, error)
	GetWorkflowResult(ctx context.Context, valuePtr interface{}, workflowID string, runID string) error
	SynchronousUpdateWorkflow(
		ctx context.Context, valuePtr interface{}, workflowID, runID, updateType string, input interface{},
	) error
	ResetWorkflow(ctx context.Context, request iwfidl.WorkflowResetRequest) (runId string, err error)
	GetBackendType() (backendType service.BackendType)
	GetApiService() interface{}
}

type errorHandler interface {
	GetApplicationErrorTypeIfIsApplicationError(err error) string
	GetApplicationErrorDetails(err error, detailsPtr interface{}) error
	GetApplicationErrorTypeAndDetails(err error) (string, string)
	IsWorkflowAlreadyStartedError(error) bool
	GetRunIdFromWorkflowAlreadyStartedError(error) (string, bool)
	IsNotFoundError(error) bool
	IsRequestTimeoutError(error) bool
	IsWorkflowTimeoutError(error) bool
}

type StartWorkflowOptions struct {
	ID                       string
	TaskQueue                string
	WorkflowExecutionTimeout time.Duration
	WorkflowIDReusePolicy    *iwfidl.WorkflowIDReusePolicy
	CronSchedule             *string
	RetryPolicy              *iwfidl.WorkflowRetryPolicy
	DataAttributes           map[string]interface{}
	SearchAttributes         map[string]interface{}
	Memo                     map[string]interface{}
	WorkflowStartDelay       *time.Duration
}

type ListWorkflowExecutionsRequest struct {
	PageSize      int32
	Query         string
	NextPageToken []byte
}

type ListWorkflowExecutionsResponse struct {
	Executions    []iwfidl.WorkflowSearchResponseEntry
	NextPageToken []byte
}

type DescribeWorkflowExecutionResponse struct {
	Status                   iwfidl.WorkflowStatus
	RunId                    string
	FirstRunId               string
	SearchAttributes         map[string]iwfidl.SearchAttribute
	Memos                    map[string]iwfidl.EncodedObject
	WorkflowStartedTimestamp int64
}
