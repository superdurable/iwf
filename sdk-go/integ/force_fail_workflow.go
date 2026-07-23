package integ

import "github.com/superdurable/iwf/sdk-go/iwf"

type forceFailWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyCommunicationSchema
	iwf.EmptyPersistenceSchema
}

func (b forceFailWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&forceFailWorkflowState1{}),
	}
}
