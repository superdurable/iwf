package integ

import "github.com/superdurable/iwf/sdk-go/iwf"

type stateApiTimeoutWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyCommunicationSchema
	iwf.EmptyPersistenceSchema
}

func (b stateApiTimeoutWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&stateApiTimeoutWorkflowState1{}),
	}
}
