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

package timers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type fakeContinueAsNewChecker struct {
	thresholdMet bool
}

func (c *fakeContinueAsNewChecker) IsThresholdMet() bool {
	return c.thresholdMet
}

type fakeWorkflowProvider struct {
	nowUnixSeconds int64
	awaitErr       error
	schedulerRun   func(interfaces.UnifiedContext)
}

func (p *fakeWorkflowProvider) NewApplicationError(string, interface{}) error {
	return nil
}

func (p *fakeWorkflowProvider) IsApplicationError(error) bool {
	return false
}

func (p *fakeWorkflowProvider) GetApplicationErrorTypeAndDetails(error) (string, string) {
	return "", ""
}

func (p *fakeWorkflowProvider) GetWorkflowInfo(interfaces.UnifiedContext) interfaces.WorkflowInfo {
	return interfaces.WorkflowInfo{}
}

func (p *fakeWorkflowProvider) UpsertSearchAttributes(
	interfaces.UnifiedContext,
	map[string]interface{},
) error {
	return nil
}

func (p *fakeWorkflowProvider) SetQueryHandler(
	interfaces.UnifiedContext,
	string,
	interface{},
) error {
	return nil
}

func (p *fakeWorkflowProvider) ExtendContextWithValue(
	ctx interfaces.UnifiedContext,
	_ string,
	_ interface{},
) interfaces.UnifiedContext {
	return ctx
}

func (p *fakeWorkflowProvider) GoNamed(
	_ interfaces.UnifiedContext,
	_ string,
	run func(interfaces.UnifiedContext),
) {
	p.schedulerRun = run
}

func (p *fakeWorkflowProvider) GetThreadCount() int {
	return 0
}

func (p *fakeWorkflowProvider) GetPendingThreadNames() map[string]int {
	return nil
}

func (p *fakeWorkflowProvider) Await(
	_ interfaces.UnifiedContext,
	condition func() bool,
) error {
	condition()
	return p.awaitErr
}

func (p *fakeWorkflowProvider) WithActivityOptions(
	ctx interfaces.UnifiedContext,
	_ interfaces.ActivityOptions,
) interfaces.UnifiedContext {
	return ctx
}

func (p *fakeWorkflowProvider) ExecuteActivity(
	interface{},
	iwfpb.StepDurability,
	interfaces.UnifiedContext,
	interface{},
	...interface{},
) error {
	return nil
}

func (p *fakeWorkflowProvider) ExecuteLocalActivity(
	interface{},
	interfaces.UnifiedContext,
	interface{},
	...interface{},
) error {
	return nil
}

func (p *fakeWorkflowProvider) Now(interfaces.UnifiedContext) time.Time {
	return time.Unix(p.nowUnixSeconds, 0)
}

func (p *fakeWorkflowProvider) IsReplaying(interfaces.UnifiedContext) bool {
	return false
}

func (p *fakeWorkflowProvider) Sleep(interfaces.UnifiedContext, time.Duration) error {
	return nil
}

func (p *fakeWorkflowProvider) NewTimer(
	interfaces.UnifiedContext,
	time.Duration,
) interfaces.Future {
	return nil
}

func (p *fakeWorkflowProvider) GetSignalChannel(
	interfaces.UnifiedContext,
	string,
) interfaces.ReceiveChannel {
	return nil
}

func (p *fakeWorkflowProvider) GetContextValue(
	interfaces.UnifiedContext,
	string,
) interface{} {
	return nil
}

func (p *fakeWorkflowProvider) GetVersion(
	interfaces.UnifiedContext,
	string,
	int,
	int,
) int {
	return 0
}

func (p *fakeWorkflowProvider) GetUnhandledSignalNames(
	interfaces.UnifiedContext,
) []string {
	return nil
}

func (p *fakeWorkflowProvider) GetBackendType() service.BackendType {
	return service.BackendTypeTemporal
}

func (p *fakeWorkflowProvider) GetLogger(
	interfaces.UnifiedContext,
) interfaces.UnifiedLogger {
	return fakeLogger{}
}

func (p *fakeWorkflowProvider) NewInterpreterContinueAsNewError(
	interfaces.UnifiedContext,
	*iwfpb.InterpreterWorkflowInput,
) error {
	return nil
}

type fakeLogger struct{}

func (fakeLogger) Debug(string, ...interface{}) {}
func (fakeLogger) Info(string, ...interface{})  {}
func (fakeLogger) Warn(string, ...interface{})  {}
func (fakeLogger) Error(string, ...interface{}) {}

func TestTimerScheduler_EqualDeadlinesUseStableTieBreakers(t *testing.T) {
	scheduler := &timerScheduler{}
	timerStepB := pendingTimerInfo("condition-0", 100)
	timerStepASecond := pendingTimerInfo("condition-1", 100)
	timerStepAFirst := pendingTimerInfo("condition-0", 100)

	scheduler.addTimer(timerStepB, "step-b-1", 0)
	scheduler.addTimer(timerStepASecond, "step-a-1", 1)
	scheduler.addTimer(timerStepAFirst, "step-a-1", 0)

	require.Equal(t, []*iwfpb.TimerInfo{
		timerStepAFirst,
		timerStepASecond,
		timerStepB,
	}, scheduler.pendingTimerInfos())
}

func TestGreedyTimerProcessor_SkipBeforeRegister(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 10}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		[]*iwfpb.StaleSkipTimer{{
			StepExecutionId:     "step-1",
			TimerConditionId:    "timer-1",
			TimerConditionIndex: 0,
		}},
	)

	processor.AddTimers("step-1", []*iwfpb.TimerCondition{{
		ConditionId:                "timer-1",
		FiringUnixTimestampSeconds: 30,
	}}, nil)

	require.True(t, processor.RetryStaleSkipTimer())
	require.Empty(t, processor.Dump())
	require.Equal(
		t,
		iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED,
		processor.GetTimerInfos()["step-1"][0].GetStatus(),
	)
}

func TestGreedyTimerProcessor_SkipWinsAgainstFiring(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 10}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		nil,
	)
	processor.AddTimers("step-1", []*iwfpb.TimerCondition{{
		ConditionId:                "timer-1",
		FiringUnixTimestampSeconds: 10,
	}}, nil)

	require.True(t, processor.SkipTimer("step-1", "timer-1", 0))
	status, err := processor.WaitForTimerFiredOrSkipped(nil, "step-1", 0, ptr.Any(false))

	require.NoError(t, err)
	require.Equal(t, iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED, status)
}

func TestGreedyTimerProcessor_RemovesLosingPendingTimers(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 10}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		nil,
	)
	processor.AddTimers("step-1", []*iwfpb.TimerCondition{
		{ConditionId: "timer-1", FiringUnixTimestampSeconds: 30},
		{ConditionId: "timer-2", FiringUnixTimestampSeconds: 40},
	}, nil)

	processor.RemovePendingTimersOfStep("step-1")

	require.Empty(t, processor.GetPendingScheduledTimers())
	require.NotContains(t, processor.GetTimerInfos(), "step-1")
}

func TestGreedyTimerProcessor_AllowsEmptyConditionIDs(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 10}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		nil,
	)
	processor.AddTimers("step-1", []*iwfpb.TimerCondition{
		{FiringUnixTimestampSeconds: 30},
		{FiringUnixTimestampSeconds: 40},
	}, nil)

	timerInfos := processor.GetTimerInfos()["step-1"]
	require.Len(t, timerInfos, 2)
	require.Empty(t, timerInfos[0].GetConditionId())
	require.Empty(t, timerInfos[1].GetConditionId())
}

func TestGreedyTimerProcessor_RestoresAbsoluteTimers(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 100}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		nil,
	)
	processor.AddTimers(
		"step-1",
		[]*iwfpb.TimerCondition{
			{ConditionId: "timer-1", FiringUnixTimestampSeconds: 500},
			{ConditionId: "timer-2", FiringUnixTimestampSeconds: 90},
		},
		map[int32]iwfpb.InternalTimerStatus{
			1: iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_FIRED,
		},
	)

	require.Equal(t, int64(500), processor.GetTimerInfos()["step-1"][0].GetFiringUnixTimestampSeconds())
	require.Equal(
		t,
		iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_FIRED,
		processor.GetTimerInfos()["step-1"][1].GetStatus(),
	)
	require.Equal(t, []*iwfpb.TimerInfo{
		processor.GetTimerInfos()["step-1"][0],
	}, processor.GetPendingScheduledTimers())
}

func TestGreedyTimerProcessor_RejectsUnnormalizedRestore(t *testing.T) {
	provider := &fakeWorkflowProvider{nowUnixSeconds: 100}
	processor := NewGreedyTimerProcessor(
		nil,
		provider,
		&fakeContinueAsNewChecker{},
		nil,
	)

	require.Panics(t, func() {
		processor.AddTimers("step-1", []*iwfpb.TimerCondition{{
			ConditionId:     "timer-1",
			DurationSeconds: 10,
		}}, nil)
	})
	require.Panics(t, func() {
		processor.AddTimers("step-1", []*iwfpb.TimerCondition{{
			ConditionId: "timer-1",
		}}, nil)
	})
}

func TestNormalizeTimerConditionsFromActivityOutput(t *testing.T) {
	waitingCondition := &iwfpb.WaitingCondition{
		TimerConditions: []*iwfpb.TimerCondition{
			{ConditionId: "timer-1", DurationSeconds: 20},
			{ConditionId: "timer-2"},
		},
	}

	require.NoError(
		t,
		NormalizeTimerConditionsFromActivityOutput(time.Unix(100, 0), waitingCondition),
	)

	require.Equal(t, int64(120), waitingCondition.GetTimerConditions()[0].GetFiringUnixTimestampSeconds())
	require.Equal(t, int64(100), waitingCondition.GetTimerConditions()[1].GetFiringUnixTimestampSeconds())
	require.Zero(t, waitingCondition.GetTimerConditions()[0].GetDurationSeconds())
	require.Zero(t, waitingCondition.GetTimerConditions()[1].GetDurationSeconds())
}

func pendingTimerInfo(conditionID string, firingUnixSeconds int64) *iwfpb.TimerInfo {
	return &iwfpb.TimerInfo{
		ConditionId:                conditionID,
		FiringUnixTimestampSeconds: firingUnixSeconds,
		Status:                     iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING,
	}
}
