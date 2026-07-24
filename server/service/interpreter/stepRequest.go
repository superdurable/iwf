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

import "github.com/superdurable/iwf/gen/iwfpb"

type StepRequest struct {
	startRequest  *iwfpb.StepMovement
	resumeRequest *iwfpb.StepExecutionResumeInfo
}

func NewStepStartRequest(movement *iwfpb.StepMovement) StepRequest {
	if movement == nil {
		panic("step start request requires a movement")
	}
	return StepRequest{startRequest: movement}
}

func NewStepResumeRequest(resumeRequest *iwfpb.StepExecutionResumeInfo) StepRequest {
	if resumeRequest == nil || resumeRequest.GetStep() == nil {
		panic("step resume request requires resume info and a movement")
	}
	if resumeRequest.GetStepExecutionId() == "" {
		panic("step resume request requires an execution ID")
	}
	return StepRequest{resumeRequest: resumeRequest}
}

func (request StepRequest) GetStepStartRequest() *iwfpb.StepMovement {
	if request.IsResumeRequest() {
		panic("resume request has no start request")
	}
	return request.startRequest
}

func (request StepRequest) GetStepResumeRequest() *iwfpb.StepExecutionResumeInfo {
	if !request.IsResumeRequest() {
		panic("start request has no resume request")
	}
	return request.resumeRequest
}

func (request StepRequest) IsResumeRequest() bool {
	return request.resumeRequest != nil
}

func (request StepRequest) GetStepMovement() *iwfpb.StepMovement {
	if request.IsResumeRequest() {
		return request.resumeRequest.GetStep()
	}
	if request.startRequest == nil {
		panic("invalid empty StepRequest")
	}
	return request.startRequest
}

func (request StepRequest) GetStepType() string {
	return request.GetStepMovement().GetStepType()
}
