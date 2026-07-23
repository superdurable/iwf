package iwf

import "github.com/superdurable/iwf/sdk-go/gen/iwfidl"

type UnregisteredWorkflowOptions struct {
	WorkflowIdReusePolicy     *iwfidl.IDReusePolicy
	WorkflowCronSchedule      *string
	WorkflowStartDelaySeconds *int32
	WorkflowRetryPolicy       *iwfidl.WorkflowRetryPolicy
	StartStateOptions         *iwfidl.WorkflowStateOptions
	InitialSearchAttributes   []iwfidl.SearchAttribute
}
