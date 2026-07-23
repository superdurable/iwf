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
	anytimersignal "github.com/superdurable/iwf/integ/workflow/any_timer_signal"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestAnyTimerSignalWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeTemporal, nil)
		smallWaitForFastTest()
	}
}

func TestGreedyAnyTimerSignalWorkflowTemporal(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeTemporal, minimumGreedyTimerConfig())
		smallWaitForFastTest()
	}
}

func TestAnyTimerSignalWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeCadence, nil)
		smallWaitForFastTest()
	}
}

func TestGreedyAnyTimerSignalWorkflowCadence(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeCadence, minimumGreedyTimerConfig())
		smallWaitForFastTest()
	}
}

func TestAnyTimerSignalWorkflowTemporalContinueAsNew(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeTemporal, minimumContinueAsNewConfig(true))
		smallWaitForFastTest()
	}
}

func TestGreedyAnyTimerSignalWorkflowTemporalContinueAsNew(t *testing.T) {
	if !*temporalIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeTemporal, greedyTimerConfig(true))
		smallWaitForFastTest()
	}
}

func TestAnyTimerSignalWorkflowCadenceContinueAsNew(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeCadence, minimumContinueAsNewConfig(false))
		smallWaitForFastTest()
	}
}

func TestGreedyAnyTimerSignalWorkflowCadenceContinueAsNew(t *testing.T) {
	if !*cadenceIntegTest {
		t.Skip()
	}
	for i := 0; i < *repeatIntegTest; i++ {
		doTestAnyTimerSignalWorkflow(t, service.BackendTypeCadence, greedyTimerConfig(true))
		smallWaitForFastTest()
	}
}

func doTestAnyTimerSignalWorkflow(t *testing.T, backendType service.BackendType, config *iwfidl.WorkflowConfig) {
	// start test workflow server
	wfHandler := anytimersignal.NewHandler()
	closeFunc1 := startWorkflowWorker(wfHandler, t)
	defer closeFunc1()

	closeFunc2 := startIwfService(backendType)
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
	wfId := anytimersignal.WorkflowType + strconv.Itoa(int(time.Now().UnixNano()))
	req := apiClient.DefaultApi.ApiV1WorkflowStartPost(context.Background())
	_, httpResp, err := req.WorkflowStartRequest(iwfidl.WorkflowStartRequest{
		WorkflowId:             wfId,
		IwfWorkflowType:        anytimersignal.WorkflowType,
		WorkflowTimeoutSeconds: 20,
		IwfWorkerUrl:           "http://localhost:" + testWorkflowServerPort,
		StartStateId:           ptr.Any(anytimersignal.State1),
		WorkflowStartOptions: &iwfidl.WorkflowStartOptions{
			WorkflowConfigOverride: config,
		},
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// Delay for 3 secs and then send the signal
	time.Sleep(time.Second * 3)
	signalValue := iwfidl.EncodedObject{
		Encoding: iwfidl.PtrString("json"),
		Data:     iwfidl.PtrString("test-data-1"),
	}
	req2 := apiClient.DefaultApi.ApiV1WorkflowSignalPost(context.Background())
	httpResp, err = req2.WorkflowSignalRequest(iwfidl.WorkflowSignalRequest{
		WorkflowId:        wfId,
		SignalChannelName: anytimersignal.SignalName,
		SignalValue:       &signalValue,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	// Wait for the workflow to complete
	reqWait := apiClient.DefaultApi.ApiV1WorkflowGetWithWaitPost(context.Background())
	_, httpResp, err = reqWait.WorkflowGetRequest(iwfidl.WorkflowGetRequest{
		WorkflowId: wfId,
	}).Execute()
	failTestAtHttpError(err, httpResp, t)

	history, data := wfHandler.GetTestResult()
	assertions := assert.New(t)
	assertions.Equalf(map[string]int64{
		"S1_start":  2,
		"S1_decide": 2,
		"S2_start":  1,
		"S2_decide": 1,
	}, history, "anytimersignal test fail, %v", history)

	assertions.Equal(anytimersignal.SignalName, data["signalChannelName1"])
	assertions.Equal("signal-cmd-id", data["signalCommandId1"])
	assertions.Equal(iwfidl.WAITING, data["signalStatus1"])

	assertions.Equal(anytimersignal.SignalName, data["signalChannelName2"])
	assertions.Equal("signal-cmd-id", data["signalCommandId2"])
	assertions.Equal(iwfidl.RECEIVED, data["signalStatus2"])
	assertions.Equal(signalValue, data["signalValue2"])
}
