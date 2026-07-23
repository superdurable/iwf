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
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"github.com/superdurable/iwf/sdk-go/iwf"
	"time"
)

type stateApiTimeoutWorkflowState1 struct {
	iwf.DefaultStateId
}

func (b stateApiTimeoutWorkflowState1) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	time.Sleep(time.Minute)
	return nil, nil
}

func (b stateApiTimeoutWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	return iwf.ForceFailWorkflow("a failing message"), nil
}

func (b stateApiTimeoutWorkflowState1) GetStateOptions() *iwf.StateOptions {
	return &iwf.StateOptions{
		WaitUntilApiRetryPolicy: &iwfidl.RetryPolicy{
			MaximumAttempts: iwfidl.PtrInt32(1),
		},
		WaitUntilApiTimeoutSeconds: iwfidl.PtrInt32(1),
	}
}
