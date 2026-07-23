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
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
)

type StateRequestQueue struct {
	queue []StateRequest
}

func NewStateRequestQueue() *StateRequestQueue {
	return &StateRequestQueue{}
}

func NewStateRequestQueueWithResumeRequests(startReqs []iwfidl.StateMovement, resumeReqs map[string]service.StateExecutionResumeInfo) *StateRequestQueue {
	var queue []StateRequest
	for _, r := range startReqs {
		queue = append(queue, NewStateStartRequest(r))
	}

	for _, k := range DeterministicKeys(resumeReqs) {
		queue = append(queue, NewStateResumeRequest(resumeReqs[k]))
	}

	return &StateRequestQueue{
		queue: queue,
	}
}

func (srq *StateRequestQueue) IsEmpty() bool {
	return len(srq.queue) == 0
}

func (srq *StateRequestQueue) TakeAll() []StateRequest {
	// copy the whole slice(pointer)
	res := srq.queue
	//reset to empty slice since each iteration will process all current states in the queue
	srq.queue = nil
	return res
}

func (srq *StateRequestQueue) GetAllStateStartRequests() []iwfidl.StateMovement {
	var res []iwfidl.StateMovement
	for _, r := range srq.queue {
		if r.IsResumeRequest() {
			continue
		}
		res = append(res, r.GetStateStartRequest())
	}
	return res
}

func (srq *StateRequestQueue) GetAllStateResumeRequests() []service.StateExecutionResumeInfo {
	var res []service.StateExecutionResumeInfo
	for _, r := range srq.queue {
		if !r.IsResumeRequest() {
			continue
		}
		res = append(res, r.GetStateResumeRequest())
	}
	return res
}

func (srq *StateRequestQueue) AddStateStartRequests(reqs []iwfidl.StateMovement) {
	for _, r := range reqs {
		srq.queue = append(srq.queue, NewStateStartRequest(r))
	}
}

func (srq *StateRequestQueue) AddSingleStateStartRequest(stateId string, input *iwfidl.EncodedObject, options *iwfidl.WorkflowStateOptions) {
	srq.queue = append(srq.queue, NewStateStartRequest(
		iwfidl.StateMovement{
			StateId:      stateId,
			StateInput:   input,
			StateOptions: options,
		},
	))
}
