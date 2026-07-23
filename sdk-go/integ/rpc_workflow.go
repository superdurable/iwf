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
	"fmt"
	"github.com/superdurable/iwf/sdk-go/iwf"
)

type rpcWorkflow struct {
	iwf.WorkflowDefaults
}

func (b rpcWorkflow) GetCommunicationSchema() []iwf.CommunicationMethodDef {
	return []iwf.CommunicationMethodDef{
		iwf.InternalChannelDef("test"),
		iwf.RPCMethodDef(b.TestRPC, nil),
		iwf.RPCMethodDef(b.TestErrorRPC, nil),
	}
}

func (b rpcWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&rpcWorkflowState1{}),
	}
}

func (b rpcWorkflow) TestRPC(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (interface{}, error) {
	var i int
	input.Get(&i)
	i++
	communication.PublishInternalChannel("test", i)
	return i, nil
}

func (b rpcWorkflow) TestErrorRPC(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (interface{}, error) {
	return nil, fmt.Errorf("test error")
}

type rpcWorkflowState1 struct {
	iwf.WorkflowStateDefaults
}

func (b rpcWorkflowState1) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	return iwf.AllCommandsCompletedRequest(
		iwf.NewInternalChannelCommand("", "test"),
	), nil
}

func (b rpcWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	var i int
	input.Get(&i)
	var j int
	commandResults.InternalChannelCommands[0].Value.Get(&j)
	return iwf.GracefulCompleteWorkflow(i + j), nil
}
