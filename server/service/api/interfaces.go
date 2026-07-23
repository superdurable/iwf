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

package api

import (
	"context"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service/common/errors"
)

type ApiService interface {
	ApiV1WorkflowStartPost(
		ctx context.Context, request iwfidl.WorkflowStartRequest,
	) (*iwfidl.WorkflowStartResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowWaitForStateCompletion(
		ctx context.Context, request iwfidl.WorkflowWaitForStateCompletionRequest,
	) (*iwfidl.WorkflowWaitForStateCompletionResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowSignalPost(ctx context.Context, request iwfidl.WorkflowSignalRequest) *errors.ErrorAndStatus
	ApiV1WorkflowPublishToInternalChannelPost(ctx context.Context, request iwfidl.PublishToInternalChannelRequest) *errors.ErrorAndStatus
	ApiV1WorkflowStopPost(ctx context.Context, request iwfidl.WorkflowStopRequest) *errors.ErrorAndStatus
	ApiV1WorkflowConfigUpdate(ctx context.Context, request iwfidl.WorkflowConfigUpdateRequest) *errors.ErrorAndStatus
	ApiV1WorkflowTriggerContinueAsNew(
		ctx context.Context, req iwfidl.TriggerContinueAsNewRequest,
	) (retError *errors.ErrorAndStatus)
	ApiV1WorkflowGetQueryAttributesPost(
		ctx context.Context, request iwfidl.WorkflowGetDataObjectsRequest,
	) (*iwfidl.WorkflowGetDataObjectsResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowSetQueryAttributesPost(
		ctx context.Context, request iwfidl.WorkflowSetDataObjectsRequest) *errors.ErrorAndStatus
	ApiV1WorkflowGetSearchAttributesPost(
		ctx context.Context, request iwfidl.WorkflowGetSearchAttributesRequest,
	) (*iwfidl.WorkflowGetSearchAttributesResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowSetSearchAttributesPost(
		ctx context.Context, request iwfidl.WorkflowSetSearchAttributesRequest) *errors.ErrorAndStatus
	ApiV1WorkflowGetPost(
		ctx context.Context, request iwfidl.WorkflowGetRequest,
	) (*iwfidl.WorkflowGetResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowGetWithWaitPost(
		ctx context.Context, request iwfidl.WorkflowGetRequest,
	) (*iwfidl.WorkflowGetResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowSearchPost(
		ctx context.Context, request iwfidl.WorkflowSearchRequest,
	) (*iwfidl.WorkflowSearchResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowRpcPost(
		ctx context.Context, request iwfidl.WorkflowRpcRequest,
	) (*iwfidl.WorkflowRpcResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowResetPost(
		ctx context.Context, request iwfidl.WorkflowResetRequest,
	) (*iwfidl.WorkflowResetResponse, *errors.ErrorAndStatus)
	ApiV1WorkflowSkipTimerPost(ctx context.Context, request iwfidl.WorkflowSkipTimerRequest) *errors.ErrorAndStatus
	ApiV1WorkflowDumpPost(
		ctx context.Context, request iwfidl.WorkflowDumpRequest,
	) (*iwfidl.WorkflowDumpResponse, *errors.ErrorAndStatus)
	ApiInfoHealth(ctx context.Context) *iwfidl.HealthInfo
	Close()
}
