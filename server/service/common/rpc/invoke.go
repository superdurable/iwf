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

package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"io"
	"net/http"

	"github.com/superdurable/iwf/config"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/blobstore"
	"github.com/superdurable/iwf/service/common/errors"
	"github.com/superdurable/iwf/service/common/urlautofix"
	"github.com/superdurable/iwf/service/common/utils"
)

func InvokeWorkerRpc(
	ctx context.Context, rpcPrep *service.PrepareRpcQueryResponse, req iwfidl.WorkflowRpcRequest, apiMaxSeconds int64, blobStore blobstore.BlobStore, externalStorageConfig config.ExternalStorageConfig,
) (*iwfidl.WorkflowWorkerRpcResponse, *errors.ErrorAndStatus) {
	iwfWorkerBaseUrl := urlautofix.FixWorkerUrl(rpcPrep.IwfWorkerUrl)
	// invoke worker rpc
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: iwfWorkerBaseUrl,
			},
		},
	})

	err := blobstore.LoadDataObjectsFromExternalStorage(ctx, rpcPrep.DataObjects, blobStore)
	if err != nil {
		return nil, handleWorkerRpcResponseError(err, nil)
	}

	rpcCtx, cancel := utils.TrimContextByTimeoutWithCappedDDL(ctx, req.TimeoutSeconds, apiMaxSeconds)
	defer cancel()
	workerReq := apiClient.DefaultApi.ApiV1WorkflowWorkerRpcPost(rpcCtx)

	// creating empty maps for signalChannelInfos & internalChannelInfos instead of passing in nils
	// using nil causes problems when converting to map model defined with OpenAPI
	var signalChannelInfos map[string]iwfidl.ChannelInfo
	if rpcPrep.SignalChannelInfo == nil {
		signalChannelInfos = make(map[string]iwfidl.ChannelInfo)
	} else {
		signalChannelInfos = rpcPrep.SignalChannelInfo
	}

	var internalChannelInfos map[string]iwfidl.ChannelInfo
	if rpcPrep.InternalChannelInfo == nil {
		internalChannelInfos = make(map[string]iwfidl.ChannelInfo)
	} else {
		internalChannelInfos = rpcPrep.InternalChannelInfo
	}

	workerRequest := iwfidl.WorkflowWorkerRpcRequest{
		Context: iwfidl.Context{
			WorkflowId:               req.WorkflowId,
			WorkflowRunId:            rpcPrep.WorkflowRunId,
			WorkflowStartedTimestamp: rpcPrep.WorkflowStartedTimestamp,
		},
		WorkflowType:         rpcPrep.IwfWorkflowType,
		RpcName:              req.RpcName,
		Input:                req.Input,
		SearchAttributes:     rpcPrep.SearchAttributes,
		DataAttributes:       rpcPrep.DataObjects,
		SignalChannelInfos:   &signalChannelInfos,
		InternalChannelInfos: &internalChannelInfos,
	}
	resp, httpResp, err := workerReq.WorkflowWorkerRpcRequest(workerRequest).Execute()
	if utils.CheckHttpError(err, httpResp) {
		return nil, handleWorkerRpcResponseError(err, httpResp)
	}
	decision := resp.GetStateDecision()
	if decision.HasConditionalClose() {
		return nil, handleWorkerRpcResponseError(fmt.Errorf("closing workflow in RPC is not supported yet"), nil)
	}

	if resp.UpsertDataAttributes != nil {
		err = blobstore.WriteDataObjectsToExternalStorage(ctx, resp.UpsertDataAttributes, req.WorkflowId, externalStorageConfig.ThresholdInBytes, blobStore, externalStorageConfig.Enabled)
		if err != nil {
			return nil, handleWorkerRpcResponseError(err, nil)
		}
	}

	for _, st := range decision.GetNextStates() {
		if service.ValidClosingWorkflowStateId[st.GetStateId()] {
			// TODO this need more work in workflow to support
			return nil, handleWorkerRpcResponseError(fmt.Errorf("closing workflow in RPC is not supported yet"), nil)
		}
	}
	return resp, nil
}

func handleWorkerRpcResponseError(err error, httpResp *http.Response) *errors.ErrorAndStatus {
	detailedMessage := err.Error()
	if err != nil {
		detailedMessage = err.Error()
	}

	var originalStatusCode int
	var workerError iwfidl.WorkerErrorResponse
	if httpResp != nil {
		originalStatusCode = httpResp.StatusCode
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			detailedMessage = "cannot read body from http response"
		} else {
			err := json.Unmarshal(body, &workerError)
			if err != nil {
				detailedMessage = "unable to decode worker response body to WorkerErrorResponse: body" + string(body)
			} else {
				detailedMessage = fmt.Sprintf("worker API error, status:%v, errorType:%v", originalStatusCode, workerError.GetErrorType())
			}
		}

	}

	return errors.NewErrorAndStatusWithWorkerError(
		service.HttpStatusCodeSpecial4xxError1,
		iwfidl.WORKER_API_ERROR,
		detailedMessage,
		workerError.GetDetail(),
		workerError.GetErrorType(),
		int32(originalStatusCode),
	)
}
