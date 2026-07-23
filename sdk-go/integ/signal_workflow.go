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

import "github.com/superdurable/iwf/sdk-go/iwf"

type signalWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyPersistenceSchema
}

const testChannelName1 = "test-channel-name-1"
const testChannelName2 = "test-channel-name-2"

func (b signalWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&signalWorkflowState1{}),
		iwf.NonStartingStateDef(&signalWorkflowState2{}),
	}
}

func (b signalWorkflow) GetCommunicationSchema() []iwf.CommunicationMethodDef {
	return []iwf.CommunicationMethodDef{
		iwf.SignalChannelDef(testChannelName1),
		iwf.SignalChannelDef(testChannelName2),
	}
}
