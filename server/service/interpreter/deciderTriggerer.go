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
	"fmt"
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/compatibility"
)

func IsDeciderTriggerConditionMet(
	commandReq iwfidl.CommandRequest,
	completedTimerCmds map[int]service.InternalTimerStatus,
	completedSignalCmds map[int]*iwfidl.EncodedObject,
	completedInterStateChannelCmds map[int]*iwfidl.EncodedObject,
) bool {
	if len(commandReq.GetTimerCommands())+len(commandReq.GetSignalCommands())+len(commandReq.GetInterStateChannelCommands()) > 0 {
		triggerType := compatibility.GetDeciderTriggerType(commandReq)
		if triggerType == iwfidl.ALL_COMMAND_COMPLETED {
			return len(completedTimerCmds) == len(commandReq.GetTimerCommands()) &&
				len(completedSignalCmds) == len(commandReq.GetSignalCommands()) &&
				len(completedInterStateChannelCmds) == len(commandReq.GetInterStateChannelCommands())
		} else if triggerType == iwfidl.ANY_COMMAND_COMPLETED {
			return len(completedTimerCmds)+
				len(completedSignalCmds)+
				len(completedInterStateChannelCmds) > 0
		} else if triggerType == iwfidl.ANY_COMMAND_COMBINATION_COMPLETED {
			var completedCmdIds []string
			for _, idx := range DeterministicKeys(completedTimerCmds) {
				cmdId := commandReq.GetTimerCommands()[idx].CommandId
				completedCmdIds = append(completedCmdIds, *cmdId)
			}
			for _, idx := range DeterministicKeys(completedSignalCmds) {
				cmdId := commandReq.GetSignalCommands()[idx].CommandId
				completedCmdIds = append(completedCmdIds, *cmdId)
			}
			for _, idx := range DeterministicKeys(completedInterStateChannelCmds) {
				cmdId := commandReq.GetInterStateChannelCommands()[idx].CommandId
				completedCmdIds = append(completedCmdIds, *cmdId)
			}

			for _, acceptedComb := range commandReq.GetCommandCombinations() {
				acceptedCmdIds := make(map[string]int)
				for _, cid := range acceptedComb.GetCommandIds() {
					acceptedCmdIds[cid]++
				}

				for _, cid := range completedCmdIds {
					if acceptedCmdIds[cid] > 0 {
						acceptedCmdIds[cid]--
						if acceptedCmdIds[cid] == 0 {
							delete(acceptedCmdIds, cid)
						}
					}
				}
				if len(acceptedCmdIds) == 0 {
					return true
				}
			}
			return false
		} else {
			panic(fmt.Sprintf("unsupported decider trigger type: %v, this shouldn't happen as activity should have validated it", triggerType))
		}
	}
	return true
}
