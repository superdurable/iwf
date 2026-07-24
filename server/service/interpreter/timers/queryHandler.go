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
	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

func SetQueryHandlers(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	timerProcessor interfaces.TimerProcessor,
) error {
	if provider == nil || timerProcessor == nil {
		panic("timer query handlers require provider and processor")
	}
	if err := provider.SetQueryHandler(
		ctx,
		service.GetCurrentTimerInfosQueryType,
		func() (*iwfpb.GetCurrentTimerInfosQueryResponse, error) {
			return currentTimerInfosResponse(timerProcessor), nil
		},
	); err != nil {
		return err
	}
	return provider.SetQueryHandler(
		ctx,
		service.GetScheduledGreedyTimerTimesQueryType,
		func() (*iwfpb.GetScheduledGreedyTimerTimesQueryResponse, error) {
			return &iwfpb.GetScheduledGreedyTimerTimesQueryResponse{
				PendingScheduled: timerProcessor.GetPendingScheduledTimers(),
			}, nil
		},
	)
}

func currentTimerInfosResponse(
	timerProcessor interfaces.TimerProcessor,
) *iwfpb.GetCurrentTimerInfosQueryResponse {
	timerInfoLists := make(
		map[string]*iwfpb.TimerInfoList,
		len(timerProcessor.GetTimerInfos()),
	)
	for stepExecutionID, timerInfos := range timerProcessor.GetTimerInfos() {
		timerInfoLists[stepExecutionID] = &iwfpb.TimerInfoList{Timers: timerInfos}
	}
	return &iwfpb.GetCurrentTimerInfosQueryResponse{
		StepExecutionCurrentTimerInfos: timerInfoLists,
	}
}
