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
	"errors"

	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"github.com/superdurable/iwf/sdk-go/iwf"
)

type executeApiFailRecoveryWorkflowState1 struct {
	iwf.WorkflowStateDefaultsNoWaitUntil
}

func (b executeApiFailRecoveryWorkflowState1) GetStateId() string {
	return "execute_api_fail_recovery_workflow_state1"
}

func (b executeApiFailRecoveryWorkflowState1) GetStateOptions() *iwf.StateOptions {
	options := &iwf.StateOptions{
		ExecuteApiRetryPolicy: &iwfidl.RetryPolicy{
			InitialIntervalSeconds: iwfidl.PtrInt32(1),
			MaximumAttempts:        iwfidl.PtrInt32(1),
		},
		ExecuteApiFailureProceedState: &executeApiFailRecoveryWorkflowState2{},
	}

	return options
}

func (b executeApiFailRecoveryWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	return nil, errors.New("error")
}
