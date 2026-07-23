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
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"github.com/superdurable/iwf/sdk-go/iwf"
)

type interStateWorkflowState1 struct {
	iwf.WorkflowStateDefaults
}

func (b interStateWorkflowState1) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	return iwf.AnyCommandCompletedRequest(
			iwf.NewInternalChannelCommand("id1", interStateChannel1),
			iwf.NewInternalChannelCommand("id2", interStateChannel2)),
		nil
}

func (b interStateWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	var i int
	cmd1 := commandResults.GetInternalChannelCommandResultById("id1")
	cmd2 := commandResults.GetInternalChannelCommandResultById("id2")
	cmd2.Value.Get(&i)

	if cmd1.Status == iwfidl.WAITING && i == 2 {
		return iwf.GracefulCompletingWorkflow, nil
	}
	return nil, fmt.Errorf("error in executing %s", ctx.GetStateExecutionId())
}
