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
