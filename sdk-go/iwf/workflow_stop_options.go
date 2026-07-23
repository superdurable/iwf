package iwf

import "github.com/superdurable/iwf/sdk-go/gen/iwfidl"

type WorkflowStopOptions struct {
	StopType iwfidl.WorkflowStopType
	Reason   string
}
