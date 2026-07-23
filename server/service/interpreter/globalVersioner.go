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

package interpreter

import (
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

const globalChangeId = "global"

// StartingVersionUsingGlobalVersioning First global version
const StartingVersionUsingGlobalVersioning = 1

// StartingVersionOptimizedUpsertSearchAttribute Optimized upserting SAs
const StartingVersionOptimizedUpsertSearchAttribute = 2

// StartingVersionRenamedStateApi Renamed state API
// see: https://github.com/superdurable/iwf/pull/242/files
const StartingVersionRenamedStateApi = 3

// StartingVersionContinueAsNewOnNoStates Fix ContinueAsNew bug
const StartingVersionContinueAsNewOnNoStates = 4

// StartingVersionTemporal26SDK Upgraded Temporal SDK version which brought changes to update handler
// see: https://github.com/superdurable/iwf/releases/tag/v1.11.0
const StartingVersionTemporal26SDK = 5

// StartingVersionExecutingStateIdMode Changed default rule of upserting SAs
const StartingVersionExecutingStateIdMode = 6

// StartingVersionNoIwfGlobalVersionSearchAttribute Removed upserting IwfGlobalWorkflowVersion SA
const StartingVersionNoIwfGlobalVersionSearchAttribute = 7

// StartingVersionYieldOnConditionalComplete Bug fix to where published messages could be lost
const StartingVersionYieldOnConditionalComplete = 8

// SyncUpdateRPCUseLocalActivity Always use local activities for sync update based RPC
const SyncUpdateRPCUseLocalActivity = 9

// StartingVersionWaitingCommandThreads waits for all command threads to complete before taking a snapshot.
// This ensures that commands don't get lost during continueAsNew operations.
const StartingVersionWaitingCommandThreads = 10

const MaxOfAllVersions = StartingVersionWaitingCommandThreads

// GlobalVersioner see https://stackoverflow.com/questions/73941723/what-is-a-good-way-pattern-to-use-temporal-cadence-versioning-api
type GlobalVersioner struct {
	workflowProvider interfaces.WorkflowProvider
	ctx              interfaces.UnifiedContext
	version          int
}

func NewGlobalVersioner(
	workflowProvider interfaces.WorkflowProvider, ctx interfaces.UnifiedContext,
) (*GlobalVersioner, error) {
	version := workflowProvider.GetVersion(ctx, globalChangeId, 0, MaxOfAllVersions)

	return &GlobalVersioner{
		workflowProvider: workflowProvider,
		ctx:              ctx,
		version:          version,
	}, nil
}

// methods checking version number

func (p *GlobalVersioner) IsAfterVersionOfContinueAsNewOnNoStates() bool {
	return p.version >= StartingVersionContinueAsNewOnNoStates
}

func (p *GlobalVersioner) IsAfterVersionOfUsingGlobalVersioning() bool {
	return p.version >= StartingVersionUsingGlobalVersioning
}

func (p *GlobalVersioner) IsAfterVersionOfOptimizedUpsertSearchAttribute() bool {
	return p.version >= StartingVersionOptimizedUpsertSearchAttribute
}

func (p *GlobalVersioner) IsAfterVersionOfExecutingStateIdMode() bool {
	return p.version >= StartingVersionExecutingStateIdMode
}

func (p *GlobalVersioner) IsAfterVersionOfRenamedStateApi() bool {
	return p.version >= StartingVersionRenamedStateApi
}

func (p *GlobalVersioner) IsAfterVersionOfTemporal26SDK() bool {
	return p.version >= StartingVersionTemporal26SDK
}

func (p *GlobalVersioner) IsAfterVersionOfNoIwfGlobalVersionSearchAttribute() bool {
	return p.version >= StartingVersionNoIwfGlobalVersionSearchAttribute
}

func (p *GlobalVersioner) IsAfterVersionOfYieldOnConditionalComplete() bool {
	return p.version >= StartingVersionYieldOnConditionalComplete
}

func (p *GlobalVersioner) IsAfterVersionOfSyncUpdateRPCUseLocalActivity() bool {
	return p.version >= SyncUpdateRPCUseLocalActivity
}

func (p *GlobalVersioner) IsAfterVersionOfWaitingCommandThreads() bool {
	return p.version >= StartingVersionWaitingCommandThreads
}

// methods checking feature/functionality availability

func (p *GlobalVersioner) IsUsingGlobalVersionSearchAttribute() bool {
	return p.version >= StartingVersionUsingGlobalVersioning && p.version < StartingVersionNoIwfGlobalVersionSearchAttribute
}

func (p *GlobalVersioner) UpsertGlobalVersionSearchAttribute() error {
	if p.IsUsingGlobalVersionSearchAttribute() &&
		p.workflowProvider.GetBackendType() != service.BackendTypeCadence {
		// Note that there was bug in Cadence SDK may cause concurrent writes hence we never upsert for Cadence
		// https://github.com/uber-go/cadence-client/issues/1198

		return p.workflowProvider.UpsertSearchAttributes(p.ctx, map[string]interface{}{
			service.SearchAttributeGlobalVersion: MaxOfAllVersions,
		})
	}
	return nil
}
