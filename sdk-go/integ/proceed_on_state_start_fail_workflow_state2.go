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

type proceedOnStateStartFailWorkflowState2 struct {
	iwf.WorkflowStateDefaults
	output string
}

func (b *proceedOnStateStartFailWorkflowState2) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	var i string
	input.Get(&i)
	b.output = i + "_state2_start"
	return iwf.EmptyCommandRequest(), nil
}

func (b *proceedOnStateStartFailWorkflowState2) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	b.output += "_state2_decide"
	return iwf.GracefulCompleteWorkflow(b.output), nil
}
