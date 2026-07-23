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
	"context"
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"github.com/superdurable/iwf/sdk-go/iwf/ptr"
)

const (
	WorkflowStateWaitUntilApi = "/api/v1/workflowState/start"
	WorkflowStateExecuteApi   = "/api/v1/workflowState/decide"
	WorkflowWorkerRPCAPI      = "/api/v1/workflowWorker/rpc"
)

type WorkerService interface {
	HandleWorkflowStateWaitUntil(ctx context.Context, request iwfidl.WorkflowStateWaitUntilRequest) (*iwfidl.WorkflowStateWaitUntilResponse, error)
	HandleWorkflowStateExecute(ctx context.Context, request iwfidl.WorkflowStateExecuteRequest) (*iwfidl.WorkflowStateExecuteResponse, error)
	HandleWorkflowWorkerRPC(ctx context.Context, request iwfidl.WorkflowWorkerRpcRequest) (*iwfidl.WorkflowWorkerRpcResponse, error)
}

func NewWorkerService(registry Registry, options *WorkerOptions) WorkerService {
	if options == nil {
		options = ptr.Any(GetDefaultWorkerOptions())
	}
	return &workerServiceImpl{
		registry: registry,
		options:  *options,
	}
}
