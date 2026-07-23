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

func TestSetSearchAttributes(t *testing.T) {
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

	var signalVals []iwfidl.SearchAttribute
	signalVals = append(signalVals, iwfidl.SearchAttribute{
		Key:          iwfidl.PtrString(persistence.TestSearchAttributeIntKey),
		ValueType:    ptr.Any(iwfidl.INT),
		IntegerValue: iwfidl.PtrInt64(persistence.TestSearchAttributeIntValue1),
	},
		iwfidl.SearchAttribute{
			Key:         iwfidl.PtrString(persistence.TestSearchAttributeKeywordKey),
			ValueType:   ptr.Any(iwfidl.KEYWORD),
			StringValue: iwfidl.PtrString(persistence.TestSearchAttributeKeywordValue1),
		},
		iwfidl.SearchAttribute{
			Key:              iwfidl.PtrString(persistence.TestSearchAttributeKeywordArrayKey),
			ValueType:        ptr.Any(iwfidl.KEYWORD_ARRAY),
			StringArrayValue: []string{persistence.TestSearchAttributeKeywordValue2, persistence.TestSearchAttributeKeywordValue1},
		})

	setReq := apiClient.DefaultApi.ApiV1WorkflowSearchattributesSetPost(context.Background())
	httpResp2, err := setReq.WorkflowSetSearchAttributesRequest(iwfidl.WorkflowSetSearchAttributesRequest{
		WorkflowId:       wfId,
		SearchAttributes: signalVals,
	}).Execute()

	failTestAtHttpError(err, httpResp2, t)

	// Wait for state to store the search attributes
	time.Sleep(time.Second)

	getReq := apiClient.DefaultApi.ApiV1WorkflowSearchattributesGetPost(context.Background())
	searchResult, httpRespGet, err := getReq.WorkflowGetSearchAttributesRequest(iwfidl.WorkflowGetSearchAttributesRequest{
		WorkflowId: wfId,
		Keys: []iwfidl.SearchAttributeKeyAndType{
			{
				Key:       iwfidl.PtrString(persistence.TestSearchAttributeIntKey),
				ValueType: ptr.Any(iwfidl.INT),
			},
			{
				Key:       iwfidl.PtrString(persistence.TestSearchAttributeKeywordKey),
				ValueType: ptr.Any(iwfidl.KEYWORD),
			},
			{
				Key:       iwfidl.PtrString(persistence.TestSearchAttributeKeywordArrayKey),
				ValueType: ptr.Any(iwfidl.KEYWORD_ARRAY),
			},
		}}).Execute()
	failTestAtHttpError(err, httpRespGet, t)

	assertions.ElementsMatch(signalVals, searchResult.SearchAttributes)

	// Terminate the workflow once tests completed
	stopReq := apiClient.DefaultApi.ApiV1WorkflowStopPost(context.Background())
	_, err = stopReq.WorkflowStopRequest(iwfidl.WorkflowStopRequest{
		WorkflowId: wfId,
		StopType:   iwfidl.TERMINATE.Ptr(),
	}).Execute()
}
