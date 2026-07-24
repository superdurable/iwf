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
	"sort"
	"time"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type continueAsNewChecker interface {
	IsThresholdMet() bool
}

type scheduledTimer struct {
	timerInfo       *iwfpb.TimerInfo
	stepExecutionID string
	conditionIndex  int
}

type timerScheduler struct {
	ctx                                  interfaces.UnifiedContext
	provider                             interfaces.WorkflowProvider
	continueAsNewChecker                 continueAsNewChecker
	pendingScheduling                    []*scheduledTimer
	providerScheduledTimerUnixTimestamps []int64
}

func startGreedyTimerScheduler(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	continueAsNewChecker continueAsNewChecker,
) *timerScheduler {
	scheduler := &timerScheduler{
		ctx:                  ctx,
		provider:             provider,
		continueAsNewChecker: continueAsNewChecker,
	}
	provider.GoNamed(ctx, "greedy-timer-scheduler", scheduler.run)
	return scheduler
}

func (t *timerScheduler) run(ctx interfaces.UnifiedContext) {
	for {
		if err := t.provider.Await(ctx, t.shouldSchedule); err != nil {
			return
		}
		if t.continueAsNewChecker.IsThresholdMet() {
			return
		}
		nextTimer := t.pruneToNextTimer(t.provider.Now(ctx).Unix())
		if nextTimer == nil {
			continue
		}
		fireAt := nextTimer.timerInfo.GetFiringUnixTimestampSeconds()
		duration := time.Duration(fireAt-t.provider.Now(ctx).Unix()) * time.Second
		t.provider.NewTimer(ctx, duration)
		t.providerScheduledTimerUnixTimestamps = append(
			t.providerScheduledTimerUnixTimestamps,
			fireAt,
		)
		sort.Slice(t.providerScheduledTimerUnixTimestamps, func(left, right int) bool {
			return t.providerScheduledTimerUnixTimestamps[left] <
				t.providerScheduledTimerUnixTimestamps[right]
		})
	}
}

func (t *timerScheduler) shouldSchedule() bool {
	if t.continueAsNewChecker.IsThresholdMet() {
		return true
	}
	nextTimer := t.pruneToNextTimer(t.provider.Now(t.ctx).Unix())
	if nextTimer == nil {
		return false
	}
	if len(t.providerScheduledTimerUnixTimestamps) == 0 {
		return true
	}
	return nextTimer.timerInfo.GetFiringUnixTimestampSeconds() <
		t.providerScheduledTimerUnixTimestamps[0]
}

func (t *timerScheduler) addTimer(
	timerInfo *iwfpb.TimerInfo,
	stepExecutionID string,
	conditionIndex int,
) {
	if timerInfo == nil ||
		timerInfo.GetStatus() != iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING {
		panic("invalid timer added")
	}
	for _, pendingTimer := range t.pendingScheduling {
		if pendingTimer.timerInfo == timerInfo {
			return
		}
	}
	t.pendingScheduling = append(t.pendingScheduling, &scheduledTimer{
		timerInfo:       timerInfo,
		stepExecutionID: stepExecutionID,
		conditionIndex:  conditionIndex,
	})
	sort.Slice(t.pendingScheduling, func(left, right int) bool {
		return scheduledTimerLess(t.pendingScheduling[left], t.pendingScheduling[right])
	})
}

func (t *timerScheduler) removeTimer(timerInfo *iwfpb.TimerInfo) {
	for index, pendingTimer := range t.pendingScheduling {
		if pendingTimer.timerInfo != timerInfo {
			continue
		}
		t.pendingScheduling = append(
			t.pendingScheduling[:index],
			t.pendingScheduling[index+1:]...,
		)
		return
	}
}

func (t *timerScheduler) pruneToNextTimer(nowUnixSeconds int64) *scheduledTimer {
	activeTimestamps := t.providerScheduledTimerUnixTimestamps[:0]
	for _, timestamp := range t.providerScheduledTimerUnixTimestamps {
		if timestamp > nowUnixSeconds {
			activeTimestamps = append(activeTimestamps, timestamp)
		}
	}
	t.providerScheduledTimerUnixTimestamps = activeTimestamps

	activeTimers := t.pendingScheduling[:0]
	for _, pendingTimer := range t.pendingScheduling {
		if pendingTimer.timerInfo.GetStatus() ==
			iwfpb.InternalTimerStatus_INTERNAL_TIMER_STATUS_PENDING &&
			pendingTimer.timerInfo.GetFiringUnixTimestampSeconds() > nowUnixSeconds {
			activeTimers = append(activeTimers, pendingTimer)
		}
	}
	t.pendingScheduling = activeTimers
	if len(t.pendingScheduling) == 0 {
		return nil
	}
	return t.pendingScheduling[0]
}

func (t *timerScheduler) pendingTimerInfos() []*iwfpb.TimerInfo {
	timerInfos := make([]*iwfpb.TimerInfo, len(t.pendingScheduling))
	for index, pendingTimer := range t.pendingScheduling {
		timerInfos[index] = pendingTimer.timerInfo
	}
	return timerInfos
}

func scheduledTimerLess(left, right *scheduledTimer) bool {
	if left.timerInfo.GetFiringUnixTimestampSeconds() !=
		right.timerInfo.GetFiringUnixTimestampSeconds() {
		return left.timerInfo.GetFiringUnixTimestampSeconds() <
			right.timerInfo.GetFiringUnixTimestampSeconds()
	}
	if left.stepExecutionID != right.stepExecutionID {
		return left.stepExecutionID < right.stepExecutionID
	}
	return left.conditionIndex < right.conditionIndex
}
