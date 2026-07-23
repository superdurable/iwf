// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package compatibility

import (
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service/common/ptr"
)

func GetStartApiTimeoutSeconds(stateOptions *iwfidl.WorkflowStateOptions) int32 {
	if stateOptions == nil {
		return 0
	}
	if stateOptions.HasStartApiTimeoutSeconds() {
		return stateOptions.GetStartApiTimeoutSeconds()
	}
	return stateOptions.GetWaitUntilApiTimeoutSeconds()
}

func GetDecideApiTimeoutSeconds(stateOptions *iwfidl.WorkflowStateOptions) int32 {
	if stateOptions == nil {
		return 0
	}
	if stateOptions.HasDecideApiTimeoutSeconds() {
		return stateOptions.GetDecideApiTimeoutSeconds()
	}
	return stateOptions.GetExecuteApiTimeoutSeconds()
}

func GetStartApiRetryPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.RetryPolicy {
	if stateOptions == nil {
		return nil
	}
	if stateOptions.HasStartApiRetryPolicy() {
		return stateOptions.StartApiRetryPolicy
	}
	return stateOptions.WaitUntilApiRetryPolicy
}

func GetDecideApiRetryPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.RetryPolicy {
	if stateOptions == nil {
		return nil
	}
	if stateOptions.HasDecideApiRetryPolicy() {
		return stateOptions.DecideApiRetryPolicy
	}
	return stateOptions.ExecuteApiRetryPolicy
}

func GetWaitUntilApiDataAttributesLoadingPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.PersistenceLoadingPolicy {
	if stateOptions == nil {
		return nil
	}

	if stateOptions.HasWaitUntilApiDataAttributesLoadingPolicy() {
		return stateOptions.WaitUntilApiDataAttributesLoadingPolicy
	}

	if stateOptions.HasDataAttributesLoadingPolicy() {
		return stateOptions.DataAttributesLoadingPolicy
	}

	return stateOptions.DataObjectsLoadingPolicy
}

func GetExecuteApiDataAttributesLoadingPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.PersistenceLoadingPolicy {
	if stateOptions == nil {
		return nil
	}

	if stateOptions.HasExecuteApiDataAttributesLoadingPolicy() {
		return stateOptions.ExecuteApiDataAttributesLoadingPolicy
	}

	if stateOptions.HasDataAttributesLoadingPolicy() {
		return stateOptions.DataAttributesLoadingPolicy
	}

	return stateOptions.DataObjectsLoadingPolicy
}

func GetWaitUntilApiSearchAttributesLoadingPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.PersistenceLoadingPolicy {
	if stateOptions == nil {
		return nil
	}

	if stateOptions.HasWaitUntilApiSearchAttributesLoadingPolicy() {
		return stateOptions.WaitUntilApiSearchAttributesLoadingPolicy

	}

	return stateOptions.SearchAttributesLoadingPolicy
}

func GetExecuteApiSearchAttributesLoadingPolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.PersistenceLoadingPolicy {
	if stateOptions == nil {
		return nil
	}

	if stateOptions.HasExecuteApiSearchAttributesLoadingPolicy() {
		return stateOptions.ExecuteApiSearchAttributesLoadingPolicy

	}

	return stateOptions.SearchAttributesLoadingPolicy
}

func GetStartApiFailurePolicy(stateOptions *iwfidl.WorkflowStateOptions) *iwfidl.StartApiFailurePolicy {
	if stateOptions.HasStartApiFailurePolicy() {
		return stateOptions.StartApiFailurePolicy
	}
	if stateOptions.HasWaitUntilApiFailurePolicy() {
		newPolicy := stateOptions.GetWaitUntilApiFailurePolicy()
		switch newPolicy {
		case iwfidl.FAIL_WORKFLOW_ON_FAILURE:
			return ptr.Any(iwfidl.FAIL_WORKFLOW_ON_START_API_FAILURE)
		case iwfidl.PROCEED_ON_FAILURE:
			return ptr.Any(iwfidl.PROCEED_TO_DECIDE_ON_START_API_FAILURE)
		default:
			panic("invalid policy to convert " + string(newPolicy))
		}
	}
	return nil
}

func GetSkipWaitUntilApi(stateOptions *iwfidl.WorkflowStateOptions) bool {
	if stateOptions.HasSkipStartApi() {
		return stateOptions.GetSkipStartApi()
	}
	return stateOptions.GetSkipWaitUntil()
}
