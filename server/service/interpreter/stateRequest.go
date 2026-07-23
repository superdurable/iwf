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

type StateRequest struct {
	stateStartRequest  iwfidl.StateMovement
	isResumeRequest    bool
	stateResumeRequest service.StateExecutionResumeInfo
}

func NewStateStartRequest(movement iwfidl.StateMovement) StateRequest {
	return StateRequest{
		stateStartRequest: movement,
	}
}

func NewStateResumeRequest(resumeRequest service.StateExecutionResumeInfo) StateRequest {
	return StateRequest{
		stateResumeRequest: resumeRequest,
		isResumeRequest:    true,
	}
}

func (sq StateRequest) GetStateStartRequest() iwfidl.StateMovement {
	return sq.stateStartRequest
}

func (sq StateRequest) GetStateResumeRequest() service.StateExecutionResumeInfo {
	return sq.stateResumeRequest
}

func (sq StateRequest) IsResumeRequest() bool {
	return sq.isResumeRequest
}

func (sq StateRequest) GetStateMovement() iwfidl.StateMovement {
	if sq.isResumeRequest {
		return sq.stateResumeRequest.State
	} else {
		return sq.stateStartRequest
	}
}

func (sq StateRequest) GetStateId() string {
	if sq.IsResumeRequest() {
		return sq.stateResumeRequest.State.StateId
	}
	return sq.stateStartRequest.StateId
}
