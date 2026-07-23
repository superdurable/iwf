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
	"github.com/superdurable/iwf/integ/workflow/wf_state_options_data_attributes_loading"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestWfStateOptionsDataAttributesLoading_PARTIAL_WITHOUT_LOCK(t *testing.T) {
	for _, backendType := range getBackendTypes() {
		for i := 0; i < *repeatIntegTest; i++ {
			doTestWfStateOptionsDataAttributesLoading(t, backendType, iwfidl.PARTIAL_WITHOUT_LOCKING)
			smallWaitForFastTest()
			doTestWfStateOptionsDataAttributesLoading(t, backendType, iwfidl.PARTIAL_WITHOUT_LOCKING)
			smallWaitForFastTest()
		}
	}
}

func TestWfStateOptionsDataAttributesLoading_PARTIAL_WITH_LOCK(t *testing.T) {
	for _, backendType := range getBackendTypes() {
		for i := 0; i < *repeatIntegTest; i++ {
			doTestWfStateOptionsDataAttributesLoading(t, backendType, iwfidl.PARTIAL_WITH_EXCLUSIVE_LOCK)
			smallWaitForFastTest()
			doTestWfStateOptionsDataAttributesLoading(t, backendType, iwfidl.PARTIAL_WITHOUT_LOCKING)
			smallWaitForFastTest()
		}
	}
}

func doTestWfStateOptionsDataAttributesLoading(
	t *testing.T, backendType service.BackendType, loadingType iwfidl.PersistenceLoadingType,
) {
	assertions := assert.New(t)

	wfHandler := wf_state_options_data_attributes_loading.NewHandler()
	closeFunc1 := startWorkflowWorkerWithRpc(wfHandler, t)
	defer closeFunc1()
	closeFunc2 := startIwfService(backendType)
	defer closeFunc2()

	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})

	wfId := wf_state_options_data_attributes_loading.WorkflowType + "_" + string(loadingType) + "_" + strconv.Itoa(int(time.Now().UnixNano()))

	wfInput := &iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString(string(loadingType)),
	}

	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())

	startReq := iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        wf_state_options_data_attributes_loading.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(wf_state_options_data_attributes_loading.State1),
		StateInput:             wfInput,
	}

	_, httpResp, err := req.WorkflowStartRequest(startReq).Execute()
	failTestAtHttpError(err, httpResp, t)

	reqWait := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	_, httpResp, err = reqWait.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, _ := wfHandler.GetTestResult()

	assertions.Equalf(map[string]int64{
		"S1_start":  1,
		"S1_decide": 1,
		"S2_start":  1,
		"S2_decide": 1,
		"S3_start":  1,
		"S3_decide": 1,
		"S4_start":  1,
		"S4_decide": 1,
		"S5_start":  1,
		"S5_decide": 1,
	}, history, "state options data attributes loading, %v", history)
}
