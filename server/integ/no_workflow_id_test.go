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
	"github.com/superdurable/iwf/integ/workflow/signal"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
)

func TestSignalWorkflowNoWorkflowId(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	assertions := assert.New(t)
	_, closeFunc2 := startIwfServiceWithClient(service.BackendTypeTemporal)
	defer closeFunc2()

	// start a workflow
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})
	req := apiClient.DefaultApi.ApiV1WorkflowSignalPost(context.Background())
	httpResp, err := req.WorkflowSignalRequest(iwfidl.WorkflowSignalRequest{
		WorkflowId:        "",
		SignalChannelName: signal.SignalName,
	}).Execute()

	assertions.Equal(httpResp.StatusCode, http.StatusBadRequest)

	apiErr, ok := err.(*iwfidl.GenericOpenAPIError)
	if !ok {
		log.Fatalf("Should fail to invoke get api %v", err)
	}
	errResp, ok := apiErr.Model().(iwfidl.ErrorResponse)
	if !ok {
		log.Fatalf("should be error response")
	}
	assertions.Equal(iwfidl.WORKFLOW_NOT_EXISTS_SUB_STATUS, errResp.GetSubStatus())
	assertions.Equal("WorkflowId is not set on request.", errResp.GetDetail())
}
