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
	"github.com/superdurable/iwf/integ/workflow/interstate"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestInterStateWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestInterStateWorkflow(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
		doTestInterStateWorkflow(t, service.BackendTypeTemporal, minimumContinueAsNewConfig(true))
		smallWaitForFastTest()
	}
}

func TestInterStateWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestInterStateWorkflow(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
		doTestInterStateWorkflow(t, service.BackendTypeCadence, minimumContinueAsNewConfig(false))
		smallWaitForFastTest()
	}
}

func doTestInterStateWorkflow(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := interstate.NewHandler()
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
	wfId := interstate.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	_, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        interstate.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(interstate.State1),
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	req2 := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	resp2, httpResp, err := req2.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, data := wfHandler.GetTestResult()
	assertions := assert.New(t)
	assertions.Equalf(map[string]int64{
		"S1_start":   1,
		"S1_decide":  1,
		"S21_start":  1,
		"S21_decide": 1,
		"S22_start":  1,
		"S22_decide": 1,
		"S31_start":  1,
		"S31_decide": 1,
	}, history, "interstate test fail, %v", history)

	assertions.Equal(iwfidl.COMPLETED, resp2.GetWorkflowStatus())
	// State completions with empty output are ignored
	assertions.Equal(0, len(resp2.GetResults()))
	assertions.Equal(map[string]interface{}{
		interstate.State21 + "received": interstate.TestVal1,
		interstate.State31 + "received": interstate.TestVal2,
	}, data)
}
