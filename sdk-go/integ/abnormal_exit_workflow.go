package integ

import "github.com/superdurable/iwf/sdk-go/iwf"

type abnormalExitWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyPersistenceSchema
	iwf.EmptyCommunicationSchema
}

func (b abnormalExitWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&abnormalExitWorkflowState1{}),
	}
}
