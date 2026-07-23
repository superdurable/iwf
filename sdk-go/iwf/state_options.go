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

package iwf

import (
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
)

type StateOptions struct {
	// apply for both waitUntil and execute API
	DataAttributesLoadingPolicy   *iwfidl.PersistenceLoadingPolicy
	SearchAttributesLoadingPolicy *iwfidl.PersistenceLoadingPolicy
	// below are wait_until API specific options:
	WaitUntilApiTimeoutSeconds                *int32
	WaitUntilApiRetryPolicy                   *iwfidl.RetryPolicy
	WaitUntilApiFailurePolicy                 *iwfidl.WaitUntilApiFailurePolicy
	WaitUntilApiDataAttributesLoadingPolicy   *iwfidl.PersistenceLoadingPolicy
	WaitUntilApiSearchAttributesLoadingPolicy *iwfidl.PersistenceLoadingPolicy
	// below are execute API specific options:
	ExecuteApiTimeoutSeconds                *int32
	ExecuteApiRetryPolicy                   *iwfidl.RetryPolicy
	ExecuteApiFailureProceedState           WorkflowState
	ExecuteApiDataAttributesLoadingPolicy   *iwfidl.PersistenceLoadingPolicy
	ExecuteApiSearchAttributesLoadingPolicy *iwfidl.PersistenceLoadingPolicy
}
