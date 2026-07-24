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

	"github.com/superdurable/iwf/gen/iwfpb"
)

type StepRequestQueue struct {
	queue []StepRequest
}

func NewStepRequestQueue() *StepRequestQueue {
	return &StepRequestQueue{}
}

func RebuildStepRequestQueue(
	startRequests []*iwfpb.StepMovement,
	resumeRequests map[string]*iwfpb.StepExecutionResumeInfo,
) *StepRequestQueue {
	queue := make([]StepRequest, 0, len(startRequests)+len(resumeRequests))
	for _, request := range startRequests {
		queue = append(queue, NewStepStartRequest(request))
	}

	executionIDs := make([]string, 0, len(resumeRequests))
	for executionID := range resumeRequests {
		executionIDs = append(executionIDs, executionID)
	}
	sort.Strings(executionIDs)
	for _, executionID := range executionIDs {
		request := resumeRequests[executionID]
		if request.GetStepExecutionId() != executionID {
			panic("resume map key does not match step execution ID")
		}
		queue = append(queue, NewStepResumeRequest(request))
	}
	return &StepRequestQueue{queue: queue}
}

func (q *StepRequestQueue) IsEmpty() bool {
	return len(q.queue) == 0
}

func (q *StepRequestQueue) TakeAll() []StepRequest {
	requests := q.queue
	q.queue = nil
	return requests
}

func (q *StepRequestQueue) GetAllStepStartRequests() []*iwfpb.StepMovement {
	requests := make([]*iwfpb.StepMovement, 0, len(q.queue))
	for _, request := range q.queue {
		if !request.IsResumeRequest() {
			requests = append(requests, request.GetStepStartRequest())
		}
	}
	return requests
}

func (q *StepRequestQueue) GetAllStepResumeRequests() map[string]*iwfpb.StepExecutionResumeInfo {
	requests := make(map[string]*iwfpb.StepExecutionResumeInfo)
	for _, request := range q.queue {
		if request.IsResumeRequest() {
			resumeRequest := request.GetStepResumeRequest()
			requests[resumeRequest.GetStepExecutionId()] = resumeRequest
		}
	}
	return requests
}

func (q *StepRequestQueue) AddStepStartRequests(requests []*iwfpb.StepMovement) {
	for _, request := range requests {
		q.queue = append(q.queue, NewStepStartRequest(request))
	}
}

func (q *StepRequestQueue) AddStepResumeRequest(request *iwfpb.StepExecutionResumeInfo) {
	q.queue = append(q.queue, NewStepResumeRequest(request))
}

func (q *StepRequestQueue) AddSingleStepStartRequest(
	stepType string,
	input *iwfpb.Value,
	options *iwfpb.StepOptions,
) {
	q.queue = append(q.queue, NewStepStartRequest(&iwfpb.StepMovement{
		StepType:    stepType,
		StepInput:   input,
		StepOptions: options,
	}))
}
