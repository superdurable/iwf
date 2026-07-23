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
	"github.com/superdurable/iwf/integ/workflow/parallel"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestParallelWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestParallelWorkflow(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
		doTestParallelWorkflow(t, service.BackendTypeTemporal, minimumContinueAsNewConfig(true))
		smallWaitForFastTest()
	}
}

func TestParallelWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestParallelWorkflow(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
		doTestParallelWorkflow(t, service.BackendTypeCadence, minimumContinueAsNewConfig(false))
		smallWaitForFastTest()
	}
}

func doTestParallelWorkflow(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := parallel.NewHandler()
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
	wfId := parallel.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	_, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        parallel.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(parallel.State1),
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// Wait for the workflow to complete
	req2 := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	resp2, httpResp, err := req2.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, _ := wfHandler.GetTestResult()
	assertions := assert.New(t)
	assertions.Equalf(map[string]int64{
		"S1_start":  1,
		"S1_decide": 1,

		"S11_start":  1,
		"S11_decide": 1,
		"S12_start":  1,
		"S12_decide": 1,
		"S13_start":  1,
		"S13_decide": 1,

		"S111_start":  1,
		"S111_decide": 1,

		"S112_start":  1,
		"S112_decide": 1,

		"S121_start":  1,
		"S121_decide": 1,

		"S122_start":  1,
		"S122_decide": 1,
	}, history, "parallel test fail, %v", history)

	assertions.Equal(iwfidl.COMPLETED, resp2.GetWorkflowStatus())
	assertions.ElementsMatch([]iwfidl.StateCompletionOutput{
		{
			CompletedStateId:          parallel.State13,
			CompletedStateExecutionId: parallel.State13 + "-1",
			CompletedStateOutput: &iwfidl.EncodedObject{
				Encoding: iwfidl.PtrString("json"),
				Data:     iwfidl.PtrString("from " + parallel.State13),
			},
		},
		{
			CompletedStateId:          parallel.State111,
			CompletedStateExecutionId: parallel.State111 + "-1",
			CompletedStateOutput: &iwfidl.EncodedObject{
				Encoding: iwfidl.PtrString("json"),
				Data:     iwfidl.PtrString("from " + parallel.State111),
			},
		},
		{
			CompletedStateId:          parallel.State112,
			CompletedStateExecutionId: parallel.State112 + "-1",
			CompletedStateOutput: &iwfidl.EncodedObject{
				Encoding: iwfidl.PtrString("json"),
				Data:     iwfidl.PtrString("from " + parallel.State112),
			},
		},
		{
			CompletedStateId:          parallel.State121,
			CompletedStateExecutionId: parallel.State121 + "-1",
			CompletedStateOutput: &iwfidl.EncodedObject{
				Encoding: iwfidl.PtrString("json"),
				Data:     iwfidl.PtrString("from " + parallel.State121),
			},
		},
		{
			CompletedStateId:          parallel.State122,
			CompletedStateExecutionId: parallel.State122 + "-1",
			CompletedStateOutput: &iwfidl.EncodedObject{
				Encoding: iwfidl.PtrString("json"),
				Data:     iwfidl.PtrString("from " + parallel.State122),
			},
		},
	}, resp2.GetResults())
}
