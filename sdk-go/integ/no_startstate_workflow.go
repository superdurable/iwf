// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integ

import (
	"github.com/superdurable/iwf/sdk-go/iwf"
)

type noStartStateWorkflow struct {
	iwf.WorkflowDefaults
}

func (b noStartStateWorkflow) GetCommunicationSchema() []iwf.CommunicationMethodDef {
	return []iwf.CommunicationMethodDef{
		iwf.RPCMethodDef(b.TestRPC, nil),
	}
}

func (b noStartStateWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.NonStartingStateDef(&noStartStateWorkflowState1{}),
	}
}

func (b noStartStateWorkflow) TestRPC(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (interface{}, error) {
	var i int
	input.Get(&i)
	i++
	communication.TriggerStateMovements(iwf.NewStateMovement(noStartStateWorkflowState1{}, nil))
	return i, nil
}

type noStartStateWorkflowState1 struct {
	iwf.WorkflowStateDefaultsNoWaitUntil
}

func (b noStartStateWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	return iwf.GracefulCompletingWorkflow, nil
}
