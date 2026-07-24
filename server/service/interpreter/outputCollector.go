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
	"sort"
	"strconv"
	"strings"

	"github.com/superdurable/iwf/gen/iwfpb"
)

type OutputCollector struct {
	waitForStepTypes           map[string]struct{}
	waitForExecutionIDs        map[string]struct{}
	byExecutionID              map[string]*iwfpb.StepCompletionOutput
	firstExecutionIDByStepType map[string]string
	orderedResults             []*iwfpb.StepCompletionOutput
}

func NewOutputCollector(
	waitForStepTypes []string,
	waitForExecutionIDs []string,
) *OutputCollector {
	return &OutputCollector{
		waitForStepTypes:           stringSet(waitForStepTypes),
		waitForExecutionIDs:        stringSet(waitForExecutionIDs),
		byExecutionID:              map[string]*iwfpb.StepCompletionOutput{},
		firstExecutionIDByStepType: map[string]string{},
	}
}

func RebuildOutputCollector(
	waitForStepTypes []string,
	waitForExecutionIDs []string,
	retainedOutputs []*iwfpb.StepCompletionOutput,
	inFlight map[string]*iwfpb.StepExecutionResumeInfo,
) *OutputCollector {
	collector := NewOutputCollector(waitForStepTypes, waitForExecutionIDs)
	for _, output := range retainedOutputs {
		if output == nil || output.GetCompletedStepExecutionId() == "" {
			panic("retained step output requires an execution ID")
		}
		executionID := output.GetCompletedStepExecutionId()
		if _, duplicated := collector.byExecutionID[executionID]; duplicated {
			panic("duplicate retained step output execution ID")
		}
		collector.byExecutionID[executionID] = output
		collector.orderedResults = append(collector.orderedResults, output)
	}
	collector.sortResults()
	collector.rebuildStepTypeSelections(retainedOutputs, inFlight)
	return collector
}

func (o *OutputCollector) rebuildStepTypeSelections(
	retainedOutputs []*iwfpb.StepCompletionOutput,
	inFlight map[string]*iwfpb.StepExecutionResumeInfo,
) {
	candidates := make(map[string][]string)
	for _, output := range retainedOutputs {
		stepType := output.GetCompletedStepType()
		if _, whitelisted := o.waitForStepTypes[stepType]; whitelisted {
			candidates[stepType] = append(candidates[stepType], output.GetCompletedStepExecutionId())
		}
	}
	for executionID, resumeInfo := range inFlight {
		if resumeInfo == nil || resumeInfo.GetStep() == nil {
			panic("in-flight step output selection requires resume info")
		}
		if resumeInfo.GetStepExecutionId() != executionID {
			panic("in-flight map key does not match step execution ID")
		}
		stepType := resumeInfo.GetStep().GetStepType()
		if _, whitelisted := o.waitForStepTypes[stepType]; whitelisted {
			candidates[stepType] = append(candidates[stepType], executionID)
		}
	}
	for stepType, executionIDs := range candidates {
		sort.Slice(executionIDs, func(left, right int) bool {
			return executionIDLess(executionIDs[left], executionIDs[right])
		})
		o.firstExecutionIDByStepType[stepType] = executionIDs[0]
	}
}

func (o *OutputCollector) RegisterStepStarted(stepType, executionID string) {
	if stepType == "" || executionID == "" {
		panic("step output registration requires type and execution ID")
	}
	if _, whitelisted := o.waitForStepTypes[stepType]; !whitelisted {
		return
	}
	if _, registered := o.firstExecutionIDByStepType[stepType]; !registered {
		o.firstExecutionIDByStepType[stepType] = executionID
	}
}

func (o *OutputCollector) RecordCompletion(
	stepType string,
	executionID string,
	output *iwfpb.Value,
) bool {
	if stepType == "" || executionID == "" {
		panic("step completion requires type and execution ID")
	}
	if _, completed := o.byExecutionID[executionID]; completed {
		panic("step execution completed more than once")
	}
	if !o.isRegistered(stepType, executionID) {
		return false
	}

	completion := &iwfpb.StepCompletionOutput{
		CompletedStepType:        stepType,
		CompletedStepExecutionId: executionID,
		CompletedStepOutput:      output,
	}
	o.byExecutionID[executionID] = completion
	o.orderedResults = append(o.orderedResults, completion)
	o.sortResults()
	return true
}

func (o *OutputCollector) isRegistered(stepType, executionID string) bool {
	if _, explicitlyRegistered := o.waitForExecutionIDs[executionID]; explicitlyRegistered {
		return true
	}
	return o.firstExecutionIDByStepType[stepType] == executionID
}

func (o *OutputCollector) GetByExecutionID(
	executionID string,
) (*iwfpb.StepCompletionOutput, bool) {
	output, completed := o.byExecutionID[executionID]
	return output, completed
}

func (o *OutputCollector) GetByStepType(
	stepType string,
) (*iwfpb.StepCompletionOutput, bool) {
	executionID, registered := o.firstExecutionIDByStepType[stepType]
	if !registered {
		return nil, false
	}
	return o.GetByExecutionID(executionID)
}

func (o *OutputCollector) GetAll() []*iwfpb.StepCompletionOutput {
	return o.orderedResults
}

func (o *OutputCollector) sortResults() {
	sort.SliceStable(o.orderedResults, func(left, right int) bool {
		return executionIDLess(
			o.orderedResults[left].GetCompletedStepExecutionId(),
			o.orderedResults[right].GetCompletedStepExecutionId(),
		)
	})
}

func executionIDLess(left, right string) bool {
	leftNumber, leftHasNumber := executionNumber(left)
	rightNumber, rightHasNumber := executionNumber(right)
	if leftHasNumber && rightHasNumber && leftNumber != rightNumber {
		return leftNumber < rightNumber
	}
	if leftHasNumber != rightHasNumber {
		return leftHasNumber
	}
	return left < right
}

func executionNumber(executionID string) (int64, bool) {
	separatorIndex := strings.LastIndexByte(executionID, '-')
	if separatorIndex < 0 || separatorIndex == len(executionID)-1 {
		return 0, false
	}
	number, err := strconv.ParseInt(executionID[separatorIndex+1:], 10, 64)
	return number, err == nil
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			panic("step output whitelist contains an empty value")
		}
		set[value] = struct{}{}
	}
	return set
}
