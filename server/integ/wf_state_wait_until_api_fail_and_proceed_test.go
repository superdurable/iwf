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
	"github.com/superdurable/iwf/integ/workflow/wf_state_api_fail_and_proceed"
	"github.com/superdurable/iwf/service/common/ptr"
	"strconv"
	"testing"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
)

func TestStateApiFailAndProceedTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestStateApiFailAndProceed(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
		doTestStateApiFailAndProceed(t, service.BackendTypeTemporal, minimumContinueAsNewConfigV0())
		smallWaitForFastTest()
	}
}

func TestStateApiFailAndProceedCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestStateApiFailAndProceed(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
		doTestStateApiFailAndProceed(t, service.BackendTypeCadence, minimumContinueAsNewConfigV0())
		smallWaitForFastTest()
	}
}

func doTestStateApiFailAndProceed(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := wf_state_api_fail_and_proceed.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	closeFunc2 := startIwfService(backendType)
	defer closeFunc2()

	// start a workflow
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})
	wfId := wf_state_api_fail_and_proceed.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))

	stateOptions := &iwfidl.WorkflowStateOptions{
		StartApiRetryPolicy: &iwfidl.RetryPolicy{
			MaximumAttempts: iwfidl.PtrInt32(1),
		},
		StartApiFailurePolicy: iwfidl.PROCEED_TO_DECIDE_ON_START_API_FAILURE.Ptr(),
	}
	if config != nil { // use this hack to test the compatibility
		stateOptions = &iwfidl.WorkflowStateOptions{
			StartApiRetryPolicy: &iwfidl.RetryPolicy{
				MaximumAttempts: iwfidl.PtrInt32(1),
			},
			WaitUntilApiFailurePolicy: iwfidl.PROCEED_ON_FAILURE.Ptr(),
		}
	}

	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	startResp, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        wf_state_api_fail_and_proceed.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(wf_state_api_fail_and_proceed.State1),
		StateOptions:           stateOptions,
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// Wait for the workflow to complete
	reqWait := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	resp, httpResp, err := reqWait.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, _ := wfHandler.GetTestResult()
	assertions := assert.New(t)
	assertions.Equalf(map[string]int64{
		"S1_start":  1,
		"S1_decide": 1,
	}, history, "wf state api fail and proceed test fail, %v", history)

	assertions.Equalf(&iwfidl.WorkflowGetResponse{
		WorkflowRunId:  startResp.GetWorkflowRunId(),
		WorkflowStatus: iwfidl.COMPLETED,
		// State completions with empty output are ignored
		Results: []iwfidl.StateCompletionOutput(nil),
	}, resp, "response not expected")
}
