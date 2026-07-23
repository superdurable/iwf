package iwf

import "github.com/superdurable/iwf-golang-sdk/gen/iwfidl"

type WorkflowInfo struct {
	Status       iwfidl.WorkflowStatus
	CurrentRunId string
}
