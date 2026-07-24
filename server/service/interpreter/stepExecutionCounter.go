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
	"fmt"
	"sort"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/config"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type StepExecutionCounter struct {
	ctx      interfaces.UnifiedContext
	provider interfaces.WorkflowProvider
	configer *config.FlowConfiger

	stepTypeStartedCount            map[string]int32
	stepTypeCurrentlyExecutingCount map[string]int32
	totalCurrentlyExecutingCount    int32
}

func NewStepExecutionCounter(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	configer *config.FlowConfiger,
) *StepExecutionCounter {
	if provider == nil {
		panic("StepExecutionCounter requires a WorkflowProvider")
	}
	if configer == nil {
		panic("StepExecutionCounter requires a FlowConfiger")
	}
	return &StepExecutionCounter{
		ctx:                             ctx,
		provider:                        provider,
		configer:                        configer,
		stepTypeStartedCount:            map[string]int32{},
		stepTypeCurrentlyExecutingCount: map[string]int32{},
	}
}

func RebuildStepExecutionCounter(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	configer *config.FlowConfiger,
	counterInfo *iwfpb.StepExecutionCounterInfo,
) *StepExecutionCounter {
	if counterInfo == nil {
		panic("StepExecutionCounter restore requires counter info")
	}
	counter := NewStepExecutionCounter(ctx, provider, configer)
	counter.stepTypeStartedCount = counterInfo.GetStepTypeStartedCount()
	counter.stepTypeCurrentlyExecutingCount = counterInfo.GetStepTypeCurrentlyExecutingCount()
	counter.totalCurrentlyExecutingCount = counterInfo.GetTotalCurrentlyExecutingCount()
	counter.validateCounts()
	return counter
}

func (e *StepExecutionCounter) validateCounts() {
	if e.totalCurrentlyExecutingCount < 0 {
		panic("negative total currently executing count")
	}
	for stepType, count := range e.stepTypeStartedCount {
		if stepType == "" || count < 0 {
			panic("invalid restored step started count")
		}
	}
	var trackedCount int32
	for stepType, count := range e.stepTypeCurrentlyExecutingCount {
		if stepType == "" || count <= 0 {
			panic("invalid restored active step count")
		}
		trackedCount += count
	}
	if trackedCount > e.totalCurrentlyExecutingCount {
		panic("active step count exceeds total executing count")
	}
}

func (e *StepExecutionCounter) Dump() *iwfpb.StepExecutionCounterInfo {
	return &iwfpb.StepExecutionCounterInfo{
		StepTypeStartedCount:            e.stepTypeStartedCount,
		StepTypeCurrentlyExecutingCount: e.stepTypeCurrentlyExecutingCount,
		TotalCurrentlyExecutingCount:    e.totalCurrentlyExecutingCount,
	}
}

func (e *StepExecutionCounter) CreateNextExecutionID(stepType string) string {
	if stepType == "" {
		panic("step execution ID requires a step type")
	}
	e.stepTypeStartedCount[stepType]++
	return fmt.Sprintf("%s-%d", stepType, e.stepTypeStartedCount[stepType])
}

func (e *StepExecutionCounter) MarkStepsExecuting(requests []StepRequest) error {
	additions := make(map[string]int32)
	var newExecutions int32
	for _, request := range requests {
		if request.IsResumeRequest() {
			continue
		}
		step := request.GetStepStartRequest()
		newExecutions++
		if e.shouldTrackActiveStep(step) {
			additions[step.GetStepType()]++
		}
	}
	if newExecutions == 0 {
		return nil
	}

	if activeStepSetChangesOnAdd(e.stepTypeCurrentlyExecutingCount, additions) {
		activeStepTypes := activeStepTypesAfterAdd(e.stepTypeCurrentlyExecutingCount, additions)
		if err := e.upsertActiveStepTypes(activeStepTypes); err != nil {
			return err
		}
	}
	for stepType, count := range additions {
		e.stepTypeCurrentlyExecutingCount[stepType] += count
	}
	e.totalCurrentlyExecutingCount += newExecutions
	return nil
}

func (e *StepExecutionCounter) shouldTrackActiveStep(step *iwfpb.StepMovement) bool {
	switch e.configer.EffectiveActiveStepSearchMode() {
	case iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_DISABLED:
		return false
	case iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_ALL:
		return true
	case iwfpb.ActiveStepSearchMode_ACTIVE_STEP_SEARCH_MODE_ENABLED_FOR_STEPS_WITH_WAIT_FOR:
		return !step.GetStepOptions().GetSkipWaitFor()
	default:
		panic("FlowConfiger returned an invalid active step search mode")
	}
}

func activeStepSetChangesOnAdd(current, additions map[string]int32) bool {
	for stepType, count := range additions {
		if count > 0 && current[stepType] == 0 {
			return true
		}
	}
	return false
}

func activeStepTypesAfterAdd(current, additions map[string]int32) []string {
	stepTypes := make([]string, 0, len(current)+len(additions))
	for stepType := range current {
		stepTypes = append(stepTypes, stepType)
	}
	for stepType, count := range additions {
		if count > 0 && current[stepType] == 0 {
			stepTypes = append(stepTypes, stepType)
		}
	}
	sort.Strings(stepTypes)
	return stepTypes
}

func (e *StepExecutionCounter) MarkStepExecutionCompleted(step *iwfpb.StepMovement) error {
	if step == nil || step.GetStepType() == "" {
		panic("step completion requires a movement")
	}
	if e.totalCurrentlyExecutingCount <= 0 {
		panic("step execution count underflow")
	}

	stepType := step.GetStepType()
	var trackedCount int32
	if e.shouldTrackActiveStep(step) {
		trackedCount = e.stepTypeCurrentlyExecutingCount[stepType]
		if trackedCount == 0 {
			panic("active step execution count underflow")
		}
	}
	if trackedCount == 1 {
		activeStepTypes := activeStepTypesWithout(e.stepTypeCurrentlyExecutingCount, stepType)
		if err := e.upsertActiveStepTypes(activeStepTypes); err != nil {
			return err
		}
	}

	e.totalCurrentlyExecutingCount--
	if trackedCount > 1 {
		e.stepTypeCurrentlyExecutingCount[stepType]--
	} else if trackedCount == 1 {
		delete(e.stepTypeCurrentlyExecutingCount, stepType)
	}
	return nil
}

func activeStepTypesWithout(current map[string]int32, removedStepType string) []string {
	stepTypes := make([]string, 0, len(current))
	for stepType := range current {
		if stepType != removedStepType {
			stepTypes = append(stepTypes, stepType)
		}
	}
	sort.Strings(stepTypes)
	return stepTypes
}

func (e *StepExecutionCounter) upsertActiveStepTypes(stepTypes []string) error {
	if err := e.provider.UpsertSearchAttributes(e.ctx, map[string]interface{}{
		service.SearchAttributeExecutingStateIds: stepTypes,
	}); err != nil {
		return fmt.Errorf("upsert active step search attribute: %w", err)
	}
	return nil
}

func (e *StepExecutionCounter) GetTotalCurrentlyExecutingCount() int32 {
	return e.totalCurrentlyExecutingCount
}
