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
)

type persistenceWorkflow struct {
	iwf.DefaultWorkflowType
	iwf.EmptyCommunicationSchema
}

const (
	testDataObjectKey  = "test-data-object"
	testDataObjectKey2 = "test-data-object-2"

	testSearchAttributeInt      = "CustomIntField"
	testSearchAttributeDatetime = "CustomDatetimeField"
	testSearchAttributeBool     = "CustomBoolField"
	testSearchAttributeDouble   = "CustomDoubleField"
	testSearchAttributeText     = "CustomStringField"
	testSearchAttributeKeyword  = "CustomKeywordField"
)

func (b persistenceWorkflow) GetWorkflowStates() []iwf.StateDef {
	return []iwf.StateDef{
		iwf.StartingStateDef(&persistenceWorkflowState1{}),
		iwf.NonStartingStateDef(&persistenceWorkflowState2{}),
	}
}

func (b persistenceWorkflow) GetPersistenceSchema() []iwf.PersistenceFieldDef {
	return []iwf.PersistenceFieldDef{
		iwf.DataAttributeDef(testDataObjectKey),
		iwf.DataAttributeDef(testDataObjectKey2),
		iwf.SearchAttributeDef(testSearchAttributeInt, iwfidl.INT),
		iwf.SearchAttributeDef(testSearchAttributeDatetime, iwfidl.DATETIME),
		iwf.SearchAttributeDef(testSearchAttributeBool, iwfidl.BOOL),
		iwf.SearchAttributeDef(testSearchAttributeDouble, iwfidl.DOUBLE),
		iwf.SearchAttributeDef(testSearchAttributeText, iwfidl.TEXT),
		iwf.SearchAttributeDef(testSearchAttributeKeyword, iwfidl.KEYWORD),
	}
}
