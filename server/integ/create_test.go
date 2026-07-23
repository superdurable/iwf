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
	"github.com/superdurable/iwf/integ/helpers"
	"github.com/superdurable/iwf/integ/workflow/rpc"
	"github.com/superdurable/iwf/service"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestCreateWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestCreateWithoutStartingState(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
	}
}

func TestCreateWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestCreateWithoutStartingState(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
	}
}

func TestCreateWorkflowTemporalContinueAsNew(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestCreateWithoutStartingState(t, service.BackendTypeTemporal, minimumContinueAsNewConfig(true))
		smallWaitForFastTest()
	}
}

func TestCreateWorkflowCadenceContinueAsNew(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestCreateWithoutStartingState(t, service.BackendTypeCadence, minimumContinueAsNewConfig(false))
		smallWaitForFastTest()
	}
}

func doTestCreateWithoutStartingState(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	assertions := assert.New(t)
	// start test workflow server
	wfHandler := rpc.NewHandler()
	closeFunc1 := startWorkflowWorkerWithRpc(wfHandler, t)
	defer closeFunc1()

	uclient, closeFunc2 := startIwfServiceWithClient(backendType)
	defer closeFunc2()

	// create client
	apiClient := iwfidl.NewAPIClient(&iwfidl.Configuration{
		Servers: []iwfidl.ServerConfiguration{
			{
				URL: "http://localhost:" + testIwfServerPort,
			},
		},
	})

	// start a workflow
	wfId := rpc.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	_, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        rpc.WorkflowType,
		WorkflowTimeoutSeconds: 10,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// workflow shouldn't executed any state
	var dump service.DebugDumpResponse
	err = uclient.QueryWorkflow(context.Background(), &dump, wfId, "", service.DebugDumpQueryType)
	if err != nil {
		helpers.FailTestWithError(err, t)
	}
	assertions.Equal(service.StateExecutionCounterInfo{
		StateIdStartedCount:            make(map[string]int),
		StateIdCurrentlyExecutingCount: make(map[string]int),
		TotalCurrentlyExecutingCount:   0,
	}, dump.Snapshot.StateExecutionCounterInfo)

	// invoke an RPC to trigger the state execution
	reqRpc := apiClient.DefaultApi.ApiV1WorkflowRpcPost(context.Background())
	_, httpResp, err = reqRpc.WorkflowRpcRequest(iwfidl.WorkflowRpcRequest{
		WorkflowId: wfId,
		RpcName:    rpc.RPCName,
		Input:      &rpc.TestInput,
		SearchAttributesLoadingPolicy: &iwfidl.PersistenceLoadingPolicy{
			PersistenceLoadingType: iwfidl.PARTIAL_WITHOUT_LOCKING.Ptr(),
			PartialLoadingKeys: []string{
				rpc.TestSearchAttributeIntKey,
			},
		},
		TimeoutSeconds: iwfidl.PtrInt32(2),
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// Wait for the workflow to complete
	reqWait := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	respWait, httpResp, err := reqWait.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpErrorOrWorkflowUncompleted(err, httpResp, respWait, t)

	history, _ := wfHandler.GetTestResult()
	assertions.Equalf(map[string]int64{
		"S2_start":  1,
		"S2_decide": 1,
	}, history, "rpc test fail, %v", history)
}
