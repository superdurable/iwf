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

func GetDeciderTriggerType(commandRequest iwfidl.CommandRequest) iwfidl.DeciderTriggerType {
	if commandRequest.HasCommandWaitingType() {
		newType := commandRequest.GetCommandWaitingType()
		switch newType {
		case iwfidl.ALL_COMPLETED:
			return iwfidl.ALL_COMMAND_COMPLETED
		case iwfidl.ANY_COMPLETED:
			return iwfidl.ANY_COMMAND_COMPLETED
		case iwfidl.ANY_COMBINATION_COMPLETED:
			return iwfidl.ANY_COMMAND_COMBINATION_COMPLETED
		default:
			panic("invalid waiting type to convert:" + string(newType))
		}
	}
	return commandRequest.GetDeciderTriggerType()
}
