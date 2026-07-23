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
	"github.com/superdurable/iwf/service/common/ptr"
	"strconv"
	"testing"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/integ/workflow/basic"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
)

func TestStartWorkflowNoOptionsTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	doTestStartWorkflowWithoutStartOptions(t, service.BackendTypeTemporal)
}

func TestStartWorkflowNoOptionsCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	doTestStartWorkflowWithoutStartOptions(t, service.BackendTypeCadence)
}

func doTestStartWorkflowWithoutStartOptions(t *testing.T, backendType service.BackendType) {
	wfHandler := basic.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	client, closeFunc2 := startIwfServiceWithClient(backendType)
	defer closeFunc2()

	// start a workflow
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})
	wfId := "TestStartWorkflowWithoutStartOptions" + strconv.Itoa(int(time.Now().UnixNano()))
	wfInput := &iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString("test data"),
	}
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	startReq := iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        basic.WorkflowType,
		WorkflowTimeoutSeconds: 100,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(basic.State1),
		StateInput:             wfInput,
	}
	_, httpResp, err := req.WorkflowStartRequest(startReq).Execute()
	failTestAtHttpError(err, httpResp, t)

	requestedSAs := []iwfidl.SearchAttributeKeyAndType{
		{
			Key:       ptr.Any(service.SearchAttributeIwfWorkflowType),
			ValueType: iwfidl.KEYWORD.Ptr(),
		},
	}
	response, err := client.DescribeWorkflowExecution(context.Background(), wfId, "", requestedSAs)
	assertions := assert.New(t)
	attribute := response.SearchAttributes[service.SearchAttributeIwfWorkflowType]
	assertions.Equal(basic.WorkflowType, attribute.GetStringValue())

	// Terminate the workflow once tests completed
	stopReq := apiClient.DefaultApi.ApiV1WorkflowStopPost(context.Background())
	_, err = stopReq.WorkflowStopRequest(iwfidl.WorkflowStopRequest{
		WorkflowId: wfId,
		StopType:   iwfidl.TERMINATE.Ptr(),
	}).Execute()
}
