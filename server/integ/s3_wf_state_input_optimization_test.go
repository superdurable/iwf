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
	"strconv"
	"testing"
	"time"

	s3_state_input_optimization "github.com/superdurable/iwf/integ/workflow/s3-state-input-optimization"

	"github.com/superdurable/iwf/service/common/ptr"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
)

func TestS3WorkflowStateInputOptimizationTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}

	for i := 0; i < *repeatIntegTest; i++ {
		doTestWorkflowWithS3StateInputOptimization(t, service.BackendTypeTemporal)
		smallWaitForFastTest()
	}
}

func TestS3WorkflowStateInputOptimizationCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestWorkflowWithS3StateInputOptimization(t, service.BackendTypeCadence)
		smallWaitForFastTest()
	}
}

func doTestWorkflowWithS3StateInputOptimization(t *testing.T, backendType service.BackendType) {
	// start test workflow server
	wfHandler := s3_state_input_optimization.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	_, closeFunc2 := startIwfServiceByConfig(IwfServiceTestConfig{
		BackendType:     backendType,
		S3TestThreshold: 10, // Set low threshold so our test data gets stored in S3
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
	wfId := s3_state_input_optimization.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))

	// Create large input that will be stored in S3
	wfInput := &iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString("\"this-is-a-large-input-that-exceeds-threshold\""), // 50+ bytes
	}

	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	startReq := iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        s3_state_input_optimization.WorkflowType,
		WorkflowTimeoutSeconds: 100,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(s3_state_input_optimization.State1),
		StateInput:             wfInput,
	}
	_, httpResp, err := req.WorkflowStartRequest(startReq).Execute()
	failTestAtHttpError(err, httpResp, t)

	req2 := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	_, httpResp2, err2 := req2.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err2, httpResp2, t)

	assertions := assert.New(t)

	_, history := wfHandler.GetTestResult()

	// Verify all states received the correct input
	assertions.Equal(history["S1_start"], int64(1), "S1_start should be called once")
	assertions.Equal(history["S1_decide"], int64(1), "S1_decide should be called once")
	assertions.Equal(history["S2_start"], int64(1), "S2_start should be called once")
	assertions.Equal(history["S2_decide"], int64(1), "S2_decide should be called once")
	assertions.Equal(history["S3_start"], int64(1), "S3_start should be called once")
	assertions.Equal(history["S3_decide"], int64(1), "S3_decide should be called once")

	// Verify input data was correctly loaded at each state
	expectedData := "\"this-is-a-large-input-that-exceeds-threshold\""
	assertions.Equal(history["S1_input_data"], expectedData, "S1 should receive correct input data")
	assertions.Equal(history["S2_input_data"], expectedData, "S2 should receive correct input data (same as S1)")
	assertions.Equal(history["S3_input_data"], expectedData, "S3 should receive correct input data (same as S1 and S2)")

	// Verify optimization: should only have 1 object in S3 despite being used 3 times
	// because the same data gets reused instead of duplicated
	objectCount, err := globalBlobStore.CountWorkflowObjectsForTesting(context.Background(), wfId)
	assertions.Nil(err)
	assertions.Equal(int64(1), objectCount, "Should only have 1 object in S3 due to deduplication optimization")
}
