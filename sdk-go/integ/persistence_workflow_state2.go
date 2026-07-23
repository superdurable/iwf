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

type persistenceWorkflowState2 struct {
	iwf.WorkflowStateDefaults
}

const testText = "Hail iWF!"

func (b persistenceWorkflowState2) WaitUntil(ctx iwf.WorkflowContext, input iwf.Object, persistence iwf.Persistence, communication iwf.Communication) (*iwf.CommandRequest, error) {
	iv := persistence.GetSearchAttributeInt(testSearchAttributeInt)
	if iv != 1 {
		panic("this value must be 1 because it got set by WaitUntil API")
	}

	var do ExampleDataObjectModel
	persistence.GetDataAttribute(testDataObjectKey, &do)
	dv := persistence.GetSearchAttributeDatetime(testSearchAttributeDatetime)
	bv := persistence.GetSearchAttributeBool(testSearchAttributeBool)
	persistence.SetSearchAttributeDouble(testSearchAttributeDouble, 1.0)
	if dv.Unix() == do.Datetime.Unix() && bv {
		persistence.SetSearchAttributeText(testSearchAttributeText, testText)
		return iwf.EmptyCommandRequest(), nil
	}
	panic("the value of datatime or bool search attribute is incorrect")

}

func (b persistenceWorkflowState2) Execute(ctx iwf.WorkflowContext, input iwf.Object, commandResults iwf.CommandResults, persistence iwf.Persistence, communication iwf.Communication) (*iwf.StateDecision, error) {
	tv := persistence.GetSearchAttributeText(testSearchAttributeText)
	persistence.SetSearchAttributeKeyword(testSearchAttributeKeyword, "iWF")
	if tv == testText {
		return iwf.GracefulCompletingWorkflow, nil
	}
	panic("the value of text search attribute is incorrect")
}
