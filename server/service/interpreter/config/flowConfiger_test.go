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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
)

func TestNewFlowConfiger_PanicsOnNil(t *testing.T) {
	require.Panics(t, func() { NewFlowConfiger(nil) })
}

func TestNewFlowConfiger_PanicsOnInvalid(t *testing.T) {
	require.Panics(t, func() { NewFlowConfiger(&iwfpb.FlowConfig{ContinueAsNewThreshold: -1}) })
}

func TestFlowConfiger_GetReturnsRetainedConfig(t *testing.T) {
	in := &iwfpb.FlowConfig{ContinueAsNewThreshold: 5}
	fc := NewFlowConfiger(in)
	assert.Equal(t, int32(5), fc.Get().GetContinueAsNewThreshold())
	assert.Equal(t, int32(5), fc.EffectiveContinueAsNewThreshold())
}

func TestFlowConfiger_ZeroDefaults(t *testing.T) {
	fc := NewFlowConfiger(&iwfpb.FlowConfig{})

	assert.Equal(t, int32(0), fc.EffectiveContinueAsNewThreshold()) // 0 disables automatic CAN
	assert.Equal(t, int32(service.DefaultContinueAsNewPageSizeInBytes), fc.EffectiveContinueAsNewPageSizeInBytes())
	assert.Equal(t,
		iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_STEPS_WITH_WAIT_FOR,
		fc.EffectiveActiveStepSearchMode())
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_SYNC, fc.ResolveWaitForDurability(nil))
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_SYNC, fc.ResolveExecuteDurability(nil))
}

func TestFlowConfiger_DurabilityPrecedence(t *testing.T) {
	fc := NewFlowConfiger(&iwfpb.FlowConfig{StepDurability: iwfpb.StepDurability_STEP_DURABILITY_ASYNC})

	// no StepOptions override -> flow-level durability.
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_ASYNC, fc.ResolveWaitForDurability(&iwfpb.StepOptions{}))
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_ASYNC, fc.ResolveExecuteDurability(&iwfpb.StepOptions{}))

	// StepOptions override wins over the flow-level durability, independently per phase.
	waitOverride := &iwfpb.StepOptions{WaitForDurabilityOverride: iwfpb.StepDurability_STEP_DURABILITY_SYNC}
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_SYNC, fc.ResolveWaitForDurability(waitOverride))
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_ASYNC, fc.ResolveExecuteDurability(waitOverride))
}

func TestFlowConfiger_UpdateByAPIFullReplacement(t *testing.T) {
	fc := NewFlowConfiger(&iwfpb.FlowConfig{
		ContinueAsNewThreshold: 5,
		StepDurability:         iwfpb.StepDurability_STEP_DURABILITY_ASYNC,
	})

	require.NoError(t, fc.UpdateByAPI(&iwfpb.FlowConfig{ContinueAsNewThreshold: 9}))

	// Full replacement, not a field patch: the omitted durability resets to
	// UNSPECIFIED and therefore resolves to SYNC, not the previous ASYNC.
	assert.Equal(t, int32(9), fc.EffectiveContinueAsNewThreshold())
	assert.Equal(t, iwfpb.StepDurability_STEP_DURABILITY_SYNC, fc.ResolveExecuteDurability(nil))
}

func TestFlowConfiger_UpdateByAPIRejectsInvalidAndKeepsState(t *testing.T) {
	fc := NewFlowConfiger(&iwfpb.FlowConfig{ContinueAsNewThreshold: 7})
	assert.Error(t, fc.UpdateByAPI(&iwfpb.FlowConfig{ContinueAsNewThreshold: -5}))
	assert.Error(t, fc.UpdateByAPI(nil))
	assert.Equal(t, int32(7), fc.EffectiveContinueAsNewThreshold())
}

func TestValidateFlowConfig(t *testing.T) {
	assert.NoError(t, ValidateFlowConfig(&iwfpb.FlowConfig{}))
	assert.Error(t, ValidateFlowConfig(&iwfpb.FlowConfig{ContinueAsNewThreshold: -1}))
	assert.Error(t, ValidateFlowConfig(&iwfpb.FlowConfig{ContinueAsNewPageSizeInBytes: -1}))
	assert.Error(t, ValidateFlowConfig(&iwfpb.FlowConfig{StepDurability: iwfpb.StepDurability(99)}))
	assert.Error(t, ValidateFlowConfig(&iwfpb.FlowConfig{ActiveStepSearchMode: iwfpb.ActiveStepSearchMode(99)}))
}
