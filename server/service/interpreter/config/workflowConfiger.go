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

package config

import (
	"github.com/superdurable/iwf/gen/iwfidl"
)

type WorkflowConfiger struct {
	config iwfidl.WorkflowConfig
}

func NewWorkflowConfiger(config iwfidl.WorkflowConfig) *WorkflowConfiger {
	return &WorkflowConfiger{
		config: config,
	}
}

func (wc *WorkflowConfiger) Get() iwfidl.WorkflowConfig {
	return wc.config
}

func (wc *WorkflowConfiger) ShouldOptimizeActivity() bool {
	return wc.config.GetOptimizeActivity()
}

func (wc *WorkflowConfiger) UpdateByAPI(config iwfidl.WorkflowConfig) {
	if config.DisableSystemSearchAttribute != nil {
		wc.config.DisableSystemSearchAttribute = config.DisableSystemSearchAttribute
	}
	if config.ExecutingStateIdMode != nil {
		wc.config.ExecutingStateIdMode = config.ExecutingStateIdMode
	}
	if config.ContinueAsNewPageSizeInBytes != nil {
		wc.config.ContinueAsNewPageSizeInBytes = config.ContinueAsNewPageSizeInBytes
	}
	if config.ContinueAsNewThreshold != nil {
		wc.config.ContinueAsNewThreshold = config.ContinueAsNewThreshold
	}
	if config.OptimizeActivity != nil {
		wc.config.OptimizeActivity = config.OptimizeActivity
	}
}
