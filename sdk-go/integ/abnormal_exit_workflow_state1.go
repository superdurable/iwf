package integ

import (
	"errors"
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"github.com/superdurable/iwf/sdk-go/iwf"
)

type abnormalExitWorkflowState1 struct {
	iwf.WorkflowStateDefaultsNoWaitUntil
}

func (b abnormalExitWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	return nil, errors.New("abnormal exit state")
}

func (b abnormalExitWorkflowState1) GetStateOptions() *iwf.StateOptions {
	options := &iwf.StateOptions{
		ExecuteApiRetryPolicy: &iwfidl.RetryPolicy{
			InitialIntervalSeconds: iwfidl.PtrInt32(1),
			MaximumAttempts:        iwfidl.PtrInt32(1),
		},
	}

	return options
}
