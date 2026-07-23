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
	"github.com/superdurable/iwf/integ/workflow/skipstart"
	"github.com/superdurable/iwf/service/common/ptr"
	"strconv"
	"testing"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
)

func TestSkipStartWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestSkipStartWorkflow(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
	}
}

func TestSkipStartWorkflowTemporalContinueAsNew(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestSkipStartWorkflow(t, service.BackendTypeTemporal, minimumContinueAsNewConfigV0())
		smallWaitForFastTest()
	}
}

func TestSkipStartWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestSkipStartWorkflow(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
	}
}

func TestSkipStartWorkflowCadenceContinueAsNew(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestSkipStartWorkflow(t, service.BackendTypeCadence, minimumContinueAsNewConfigV0())
		smallWaitForFastTest()
	}
}

func doTestSkipStartWorkflow(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := skipstart.NewHandler()
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
	wfId := skipstart.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	wfInput := &iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString("test data"),
	}
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	startReq := iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        skipstart.WorkflowType,
		WorkflowTimeoutSeconds: 100,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(skipstart.State1),
		StateInput:             wfInput,
		StateOptions: &iwfidl.WorkflowStateOptions{
			SkipWaitUntil: iwfidl.PtrBool(true),
		},
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}
	_, httpResp, err := req.WorkflowStartRequest(startReq).Execute()
	failTestAtHttpError(err, httpResp, t)

	assertions := assert.New(t)

	req2 := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	resp2, httpResp, err := req2.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, _ := wfHandler.GetTestResult()
	assertions.Equalf(map[string]int64{
		"S1_decide": 1,
		"S2_decide": 1,
	}, history, "skipstart.test fail, %v", history)

	assertions.Equal(iwfidl.COMPLETED, resp2.GetWorkflowStatus())
	assertions.Equal(1, len(resp2.GetResults()))
	assertions.Equal(iwfidl.StateCompletionOutput{
		CompletedStateId:          "S2",
		CompletedStateExecutionId: "S2-1",
		CompletedStateOutput:      wfInput,
	}, resp2.GetResults()[0])
}
