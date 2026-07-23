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
	"github.com/superdurable/iwf/integ/workflow/headers"
	"github.com/superdurable/iwf/service/common/ptr"
	"strconv"
	"testing"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
)

func TestHeadersWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}

	for i := 0; i < *repeatIntegTest; i++ {
		doTestWorkflowWithHeaders(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()

		doTestWorkflowWithHeaders(t, service.BackendTypeTemporal, minimumContinueAsNewConfig(true))
		smallWaitForFastTest()

		doTestWorkflowWithHeaders(t, service.BackendTypeTemporal, &iwfidl.WorkflowConfig{
			DisableSystemSearchAttribute: iwfidl.PtrBool(true),
		})
		smallWaitForFastTest()
	}
}

func TestHeadersWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestWorkflowWithHeaders(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
		doTestWorkflowWithHeaders(t, service.BackendTypeCadence, minimumContinueAsNewConfig(false))
		smallWaitForFastTest()
	}
}

func doTestWorkflowWithHeaders(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := headers.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	_, closeFunc2 := startIwfServiceByConfig(IwfServiceTestConfig{
		BackendType: backendType,
		DefaultHeaders: map[string]string{
			headers.TestHeaderKey: headers.TestHeaderValue,
		},
	})
	defer closeFunc2()

	// start a workflow
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})
	wfId := headers.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	wfInput := &iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString("test data"),
	}
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	startReq := iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        headers.WorkflowType,
		WorkflowTimeoutSeconds: 100,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(headers.State1),
		StateInput:             wfInput,
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}
	_, httpResp, err := req.WorkflowStartRequest(startReq).Execute()
	failTestAtHttpError(err, httpResp, t)

	req2 := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	_, httpResp2, err2 := req2.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err2, httpResp2, t)

	assertions := assert.New(t)

	history, _ := wfHandler.GetTestResult()
	assertions.Equalf(map[string]int64{
		"S1_start":  1,
		"S1_decide": 1,
	}, history, "headers test fail, %v", history)
}
