package integ

import "github.com/superdurable/iwf/sdk-go/iwf"

type timerWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyCommunicationSchema
	iwf.EmptyPersistenceSchema
}

func (b timerWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&timerWorkflowState1{}),
	}
}
