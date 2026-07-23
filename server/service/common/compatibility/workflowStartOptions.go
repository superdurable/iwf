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

package compatibility

import (
	"github.com/superdurable/iwf/gen/iwfidl"
)

func GetWorkflowIdReusePolicy(options iwfidl.WorkflowStartOptions) *iwfidl.WorkflowIDReusePolicy {
	if options.HasIdReusePolicy() {
		newType := options.GetIdReusePolicy()
		switch newType {
		case iwfidl.ALLOW_IF_NO_RUNNING:
			return iwfidl.ALLOW_DUPLICATE.Ptr()
		case iwfidl.ALLOW_IF_PREVIOUS_EXISTS_ABNORMALLY, iwfidl.ALLOW_IF_PREVIOUS_EXITS_ABNORMALLY:
			// Keeping typo enum for backwards compatibility. Both old and corrected enums return the same result.
			return iwfidl.ALLOW_DUPLICATE_FAILED_ONLY.Ptr()
		case iwfidl.DISALLOW_REUSE:
			return iwfidl.REJECT_DUPLICATE.Ptr()
		case iwfidl.ALLOW_TERMINATE_IF_RUNNING:
			return iwfidl.TERMINATE_IF_RUNNING.Ptr()
		default:
			panic("invalid id reuse policy to convert:" + string(newType))
		}
	}
	return options.WorkflowIDReusePolicy
}
