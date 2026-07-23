package iwf

import "github.com/superdurable/iwf-golang-sdk/gen/iwfidl"

type WorkflowStopOptions struct {
	StopType iwfidl.WorkflowStopType
	Reason   string
}
