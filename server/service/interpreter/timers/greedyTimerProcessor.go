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
	"fmt"
	"sort"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type GreedyTimerProcessor struct {
	scheduler                   *timerScheduler
	timerInfosByStepExecutionID map[string][]*iwfpb.TimerInfo
	staleSkipTimers             []*iwfpb.StaleSkipTimer
	provider                    interfaces.WorkflowProvider
	logger                      interfaces.UnifiedLogger
}

func NewGreedyTimerProcessor(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	continueAsNewCounter continueAsNewChecker,
	staleSkipTimers []*iwfpb.StaleSkipTimer,
) *GreedyTimerProcessor {
	if provider == nil || continueAsNewCounter == nil {
		panic("GreedyTimerProcessor requires provider and continue-as-new counter")
	}
	return &GreedyTimerProcessor{
		scheduler:                   startGreedyTimerScheduler(ctx, provider, continueAsNewCounter),
		timerInfosByStepExecutionID: map[string][]*iwfpb.TimerInfo{},
		staleSkipTimers:             staleSkipTimers,
		provider:                    provider,
		logger:                      provider.GetLogger(ctx),
	}
}

func (t *GreedyTimerProcessor) SkipTimer(
	stepExecutionID string,
	timerConditionID string,
	timerConditionIndex int,
) bool {
	if t.trySkipTimer(stepExecutionID, timerConditionID, timerConditionIndex) {
		return true
	}
	t.logger.Warn(
		"timer skip did not match a pending timer",
		"stepExecutionId", stepExecutionID,
		"timerConditionId", timerConditionID,
		"timerConditionIndex", timerConditionIndex,
	)
	t.staleSkipTimers = append(t.staleSkipTimers, &iwfpb.StaleSkipTimer{
		StepExecutionId:     stepExecutionID,
		TimerConditionId:    timerConditionID,
		TimerConditionIndex: int32(timerConditionIndex),
	})
	return false
}

func (t *GreedyTimerProcessor) RetryStaleSkipTimer() bool {
	for index, staleSkip := range t.staleSkipTimers {
		if !t.trySkipTimer(
			staleSkip.GetStepExecutionId(),
			staleSkip.GetTimerConditionId(),
			int(staleSkip.GetTimerConditionIndex()),
		) {
			continue
		}
		t.staleSkipTimers = append(t.staleSkipTimers[:index], t.staleSkipTimers[index+1:]...)
		return true
	}
	return false
}

func (t *GreedyTimerProcessor) WaitForTimerFiredOrSkipped(
	ctx interfaces.UnifiedContext,
	stepExecutionID string,
	timerIndex int,
	cancelWaiting *bool,
) (iwfpb.InternalTimerStatus, error) {
	timerInfos := t.timerInfosByStepExecutionID[stepExecutionID]
	if timerIndex < 0 || timerIndex >= len(timerInfos) {
		if cancelWaiting != nil && *cancelWaiting {
			return iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING, nil
		}
		panic(fmt.Sprintf("missing timer %d for step execution %q", timerIndex, stepExecutionID))
	}
	timerInfo := timerInfos[timerIndex]
	if timerCompleted(timerInfo.GetStatus()) {
		return timerInfo.GetStatus(), nil
	}
	if t.RetryStaleSkipTimer() && timerInfo.GetStatus() == iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED {
		return timerInfo.GetStatus(), nil
	}
	if cancelWaiting == nil {
		panic("cancelWaiting is nil")
	}
	err := t.provider.Await(ctx, func() bool {
		return timerCompleted(timerInfo.GetStatus()) ||
			timerInfo.GetFiringUnixTimestampSeconds() <= t.provider.Now(ctx).Unix() ||
			*cancelWaiting
	})
	if err != nil {
		return iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING, err
	}
	if timerInfo.GetStatus() == iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED {
		return timerInfo.GetStatus(), nil
	}
	if timerInfo.GetFiringUnixTimestampSeconds() <= t.provider.Now(ctx).Unix() {
		timerInfo.Status = iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_FIRED
		t.scheduler.removeTimer(timerInfo)
		return timerInfo.GetStatus(), nil
	}
	t.scheduler.removeTimer(timerInfo)
	return iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING, nil
}

func (t *GreedyTimerProcessor) RemovePendingTimersOfStep(stepExecutionID string) {
	for _, timerInfo := range t.timerInfosByStepExecutionID[stepExecutionID] {
		if timerInfo.GetStatus() == iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING {
			t.scheduler.removeTimer(timerInfo)
		}
	}
	delete(t.timerInfosByStepExecutionID, stepExecutionID)
}

func (t *GreedyTimerProcessor) AddTimers(
	stepExecutionID string,
	timerConditions []*iwfpb.TimerCondition,
	completedTimerConditions map[int32]iwfpb.InternalTimerStatus,
) {
	validateCompletedTimerConditions(timerConditions, completedTimerConditions)
	timerInfos := make([]*iwfpb.TimerInfo, len(timerConditions))
	conditionIDs := make(map[string]bool, len(timerConditions))
	for index, timerCondition := range timerConditions {
		if timerCondition == nil {
			panic(fmt.Sprintf("nil timer condition at index %d", index))
		}
		conditionID := timerCondition.GetConditionId()
		if conditionID != "" && conditionIDs[conditionID] {
			panic(fmt.Sprintf("duplicate timer condition id %q", conditionID))
		}
		if conditionID != "" {
			conditionIDs[conditionID] = true
		}
		if timerCondition.GetDurationSeconds() != 0 {
			panic(fmt.Sprintf("timer condition at index %d retains relative duration", index))
		}
		if timerCondition.GetFiringUnixTimestampSeconds() <= 0 {
			panic(fmt.Sprintf("timer condition at index %d is not normalized", index))
		}
		status := iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING
		if completedStatus, ok := completedTimerConditions[int32(index)]; ok {
			status = completedStatus
		}
		timerInfo := &iwfpb.TimerInfo{
			ConditionId:                conditionID,
			FiringUnixTimestampSeconds: timerCondition.GetFiringUnixTimestampSeconds(),
			Status:                     status,
		}
		switch timerInfo.GetStatus() {
		case iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING:
			t.scheduler.addTimer(timerInfo, stepExecutionID, index)
		case iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_FIRED,
			iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED:
		default:
			panic(fmt.Sprintf("invalid restored timer status %s", timerInfo.GetStatus()))
		}
		timerInfos[index] = timerInfo
	}
	t.timerInfosByStepExecutionID[stepExecutionID] = timerInfos
}

func validateCompletedTimerConditions(
	timerConditions []*iwfpb.TimerCondition,
	completedTimerConditions map[int32]iwfpb.InternalTimerStatus,
) {
	indexes := make([]int, 0, len(completedTimerConditions))
	for index := range completedTimerConditions {
		indexes = append(indexes, int(index))
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		if index < 0 || index >= len(timerConditions) {
			panic(fmt.Sprintf("completed timer index %d is out of range", index))
		}
		status := completedTimerConditions[int32(index)]
		if !timerCompleted(status) {
			panic(fmt.Sprintf("completed timer index %d has invalid status %s", index, status))
		}
	}
}

func (t *GreedyTimerProcessor) trySkipTimer(
	stepExecutionID string,
	timerConditionID string,
	timerConditionIndex int,
) bool {
	timerInfo, valid := service.ValidateTimerSkipRequest(
		t.timerInfosByStepExecutionID,
		stepExecutionID,
		timerConditionID,
		timerConditionIndex,
	)
	if !valid {
		return false
	}
	timerInfo.Status = iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED
	t.scheduler.removeTimer(timerInfo)
	return true
}

func (t *GreedyTimerProcessor) Dump() []*iwfpb.StaleSkipTimer {
	return t.staleSkipTimers
}

func (t *GreedyTimerProcessor) GetTimerInfos() map[string][]*iwfpb.TimerInfo {
	return t.timerInfosByStepExecutionID
}

func (t *GreedyTimerProcessor) GetPendingScheduledTimers() []*iwfpb.TimerInfo {
	return t.scheduler.pendingTimerInfos()
}

func (t *GreedyTimerProcessor) GetTimerStartedUnixTimestamps() []int64 {
	return t.scheduler.providerScheduledTimerUnixTimestamps
}

func timerCompleted(status iwfpb.InternalTimerStatus) bool {
	return status == iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_FIRED ||
		status == iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_SKIPPED
}
