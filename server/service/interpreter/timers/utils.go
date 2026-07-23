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
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
)

func removeElement(s []service.StaleSkipTimerSignal, i int) []service.StaleSkipTimerSignal {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// FixTimerCommandFromActivityOutput converts the durationSeconds to firingUnixTimestampSeconds
// doing it right after the activity output so that we don't need to worry about the time drift after continueAsNew
func FixTimerCommandFromActivityOutput(now time.Time, request iwfidl.CommandRequest) iwfidl.CommandRequest {
	var timerCommands []iwfidl.TimerCommand
	for _, cmd := range request.GetTimerCommands() {
		if cmd.HasDurationSeconds() {
			timerCommands = append(timerCommands, iwfidl.TimerCommand{
				CommandId:                  cmd.CommandId,
				FiringUnixTimestampSeconds: iwfidl.PtrInt64(now.Unix() + int64(cmd.GetDurationSeconds())),
			})
		} else {
			timerCommands = append(timerCommands, cmd)
		}
	}
	request.TimerCommands = timerCommands
	return request
}
