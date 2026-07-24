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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/config"
)

func TestStepExecutionCounterTracksWaitForSteps(t *testing.T) {
	provider := &s2WorkflowProvider{}
	counter := NewStepExecutionCounter(
		nil,
		provider,
		config.NewFlowConfiger(&iwfpb.FlowConfig{}),
	)
	waitStep := &iwfpb.StepMovement{StepType: "wait"}
	skipStep := &iwfpb.StepMovement{
		StepType:    "skip",
		StepOptions: &iwfpb.StepOptions{SkipWaitFor: true},
	}

	require.NoError(t, counter.MarkStepsExecuting([]StepRequest{
		NewStepStartRequest(waitStep),
		NewStepStartRequest(skipStep),
	}))
	require.Equal(t, int32(2), counter.GetTotalCurrentlyExecutingCount())
	require.Equal(t, []string{"wait"}, provider.upserts[0][service.SearchAttributeExecutingStateIds])

	require.Equal(t, "wait-1", counter.CreateNextExecutionID("wait"))
	require.Equal(t, "wait-2", counter.CreateNextExecutionID("wait"))
	require.Equal(t, "skip-1", counter.CreateNextExecutionID("skip"))

	require.NoError(t, counter.MarkStepExecutionCompleted(skipStep))
	require.Len(t, provider.upserts, 1)
	require.NoError(t, counter.MarkStepExecutionCompleted(waitStep))
	require.Equal(t, int32(0), counter.GetTotalCurrentlyExecutingCount())
	require.Equal(t, []string{}, provider.upserts[1][service.SearchAttributeExecutingStateIds])
}

func TestStepExecutionCounterBackendFailureIsAtomic(t *testing.T) {
	provider := &s2WorkflowProvider{upsertErr: errors.New("backend unavailable")}
	counter := NewStepExecutionCounter(
		nil,
		provider,
		config.NewFlowConfiger(&iwfpb.FlowConfig{
			ActiveStepSearchMode: iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_ALL,
		}),
	)

	err := counter.MarkStepsExecuting([]StepRequest{
		NewStepStartRequest(&iwfpb.StepMovement{StepType: "step"}),
	})
	require.ErrorContains(t, err, "backend unavailable")
	require.Equal(t, int32(0), counter.GetTotalCurrentlyExecutingCount())
	require.Empty(t, counter.Dump().GetStepTypeCurrentlyExecutingCount())
}

func TestStepExecutionCounterDisabledModeAndSharedType(t *testing.T) {
	disabledProvider := &s2WorkflowProvider{}
	disabledCounter := NewStepExecutionCounter(
		nil,
		disabledProvider,
		config.NewFlowConfiger(&iwfpb.FlowConfig{
			ActiveStepSearchMode: iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_DISABLED,
		}),
	)
	disabledStep := &iwfpb.StepMovement{StepType: "disabled"}
	require.NoError(t, disabledCounter.MarkStepsExecuting([]StepRequest{
		NewStepStartRequest(disabledStep),
	}))
	require.NoError(t, disabledCounter.MarkStepExecutionCompleted(disabledStep))
	require.Empty(t, disabledProvider.upserts)

	sharedProvider := &s2WorkflowProvider{}
	sharedCounter := NewStepExecutionCounter(
		nil,
		sharedProvider,
		config.NewFlowConfiger(&iwfpb.FlowConfig{
			ActiveStepSearchMode: iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_ALL,
		}),
	)
	first := &iwfpb.StepMovement{StepType: "shared"}
	second := &iwfpb.StepMovement{StepType: "shared"}
	require.NoError(t, sharedCounter.MarkStepsExecuting([]StepRequest{
		NewStepStartRequest(first),
		NewStepStartRequest(second),
	}))
	require.Len(t, sharedProvider.upserts, 1)
	require.NoError(t, sharedCounter.MarkStepExecutionCompleted(first))
	require.Len(t, sharedProvider.upserts, 1)
	require.NoError(t, sharedCounter.MarkStepExecutionCompleted(second))
	require.Len(t, sharedProvider.upserts, 2)
}

func TestRebuildStepExecutionCounterRetainsProtoMaps(t *testing.T) {
	provider := &s2WorkflowProvider{}
	info := &iwfpb.StepExecutionCounterInfo{
		StepTypeStartedCount:            map[string]int32{"step": 2},
		StepTypeCurrentlyExecutingCount: map[string]int32{"step": 1},
		TotalCurrentlyExecutingCount:    1,
	}
	counter := RebuildStepExecutionCounter(
		nil,
		provider,
		config.NewFlowConfiger(&iwfpb.FlowConfig{
			ActiveStepSearchMode: iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_ALL,
		}),
		info,
	)

	info.StepTypeStartedCount["owned"] = 4
	require.Equal(t, int32(4), counter.Dump().StepTypeStartedCount["owned"])
	require.Equal(t, "step-3", counter.CreateNextExecutionID("step"))
	require.NoError(t, counter.MarkStepExecutionCompleted(&iwfpb.StepMovement{StepType: "step"}))
	require.Equal(t, int32(0), counter.GetTotalCurrentlyExecutingCount())
}
