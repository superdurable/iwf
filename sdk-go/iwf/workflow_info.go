package iwf

import "github.com/superdurable/iwf/sdk-go/gen/iwfidl"

type WorkflowInfo struct {
	Status       iwfidl.WorkflowStatus
	CurrentRunId string
}
