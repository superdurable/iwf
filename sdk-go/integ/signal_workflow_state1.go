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

type signalWorkflowState1 struct {
	iwf.WorkflowStateDefaults
}

func (b signalWorkflowState1) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	return iwf.AnyCommandCompletedRequest(
		iwf.NewSignalCommand("", testChannelName1),
		iwf.NewSignalCommand("", testChannelName2),
	), nil
}

func (b signalWorkflowState1) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	signal0 := commandResults.Signals[0]
	signal1 := commandResults.Signals[1]
	if signal0.CommandId != "" || signal0.ChannelName != testChannelName1 || signal0.Status != iwfidl.WAITING {
		panic(testChannelName1 + " should be waiting....")
	}
	if signal1.CommandId == "" && signal1.ChannelName == testChannelName2 && signal1.Status == iwfidl.RECEIVED {
		var value int
		signal1.SignalValue.Get(&value)
		return iwf.SingleNextState(signalWorkflowState2{}, value), nil
	}
	return nil, fmt.Errorf("%s doesn't receive correct value", testChannelName2)
}
