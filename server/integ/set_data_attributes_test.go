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

package integ

import (
	"context"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/integ/workflow/persistence"
	"github.com/superdurable/iwf/integ/workflow/signal"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestSetDataAttributesTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	assertions := assert.New(t)

	// start test workflow server
	wfHandler := signal.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	_, closeFunc2 := startIwfServiceWithClient(service.BackendTypeTemporal)
	defer closeFunc2()

	wfId := signal.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))

	// start a workflow
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	_, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        signal.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(signal.State1),
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	assertions.Equal(httpResp.StatusCode, http.StatusOK)

	smallDataAttributes := []iwfidl.KeyValue{
		{
			Key:   iwfidl.PtrString(persistence.TestDataAttributeKey),
			Value: &persistence.TestDataAttributeVal1,
		},
		{
			Key:   iwfidl.PtrString(persistence.TestDataAttributeKey2),
			Value: &persistence.TestDataAttributeVal2,
		},
	}

	setReq := apiClient.DefaultApi.ApiV1WorkflowDataobjectsSetPost(context.Background())
	httpResp2, err := setReq.WorkflowSetDataObjectsRequest(iwfidl.WorkflowSetDataObjectsRequest{
		WorkflowId: wfId,
		Objects:    smallDataAttributes,
	}).Execute()

	failTestAtHttpError(err, httpResp2, t)

	// Wait for state to store the data attributes
	time.Sleep(time.Second)

	getReq := apiClient.DefaultApi.ApiV1WorkflowDataobjectsGetPost(context.Background())
	getResult, httpRespGet, err := getReq.WorkflowGetDataObjectsRequest(iwfidl.WorkflowGetDataObjectsRequest{
		WorkflowId: wfId,
		Keys: []string{
			persistence.TestDataAttributeKey, persistence.TestDataAttributeKey2,
		}}).Execute()
	failTestAtHttpError(err, httpRespGet, t)

	assertions.ElementsMatch(smallDataAttributes, getResult.Objects)

	// Terminate the workflow once tests completed
	stopReq := apiClient.DefaultApi.ApiV1WorkflowStopPost(context.Background())
	_, err = stopReq.WorkflowStopRequest(iwfidl.WorkflowStopRequest{
		WorkflowId: wfId,
		StopType:   iwfidl.TERMINATE.Ptr(),
	}).Execute()
}
