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

package replayTests

import (
	"testing"

	"github.com/superdurable/iwf/service/interpreter/temporal"
	"github.com/stretchr/testify/assert"

	"go.temporal.io/sdk/worker"
)

var jsonHistoryFiles = []string{
	"v1-persistence.json",
	"v2-persistence.json",
	"v2-basic.json",
	"v2-basic-disable-system-searchattributes.json",
	"v2-any-timer-signal.json",
	"v3-any-timer-signal-continue-as-new.json",
	"v3-basic.json",
	"v3-skip-start.json",
	"v3-bug-no-state-stuck.json",
	"v4-continue-as-new-on-no-state.json",
	"v4-continued-as-new-before-versioning-optimization.json",
	"v4-local-activity-optimization.json",
	"v5-basic.json",
	"v6-search-attributes-optimization-enabled-for-all.json",
	"v6-search-attributes-optimization-default.json",
	"v7-no-global-version-search-attribute.json",
	"v8-yield-on-conditional-complete.json",
	"v8-activity-for-sync-updates-rpcs.json",
	"v9-force-local-activity-for-sync-updates-rpcs.json",
	"v9-pass-search-attributes-in-execute-activity.json",
	"v10-command-thread-completion-CAN1.json",
	"v10-command-thread-completion-CAN2.json",
	"v10-command-thread-completion-wf-finish.json",
	"v10-any-command-thread-completion-CAN1.json",
	"v10-any-command-thread-completion-wf-finish.json",
}

func TestTemporalReplay(t *testing.T) {
	worker.EnableVerboseLogging(true)

	replayer, err := worker.NewWorkflowReplayerWithOptions(
		worker.WorkflowReplayerOptions{
			EnableLoggingInReplay: true,
		})

	if err != nil {
		panic(err)
	}

	replayer.RegisterWorkflow(temporal.Interpreter)

	for _, f := range jsonHistoryFiles {
		err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "history/"+f)
		assertions := assert.New(t)
		assertions.Nil(err, "fail at replay history for: "+f)
	}

}

// NOTE: set TEMPORAL_DEBUG=true
//func TestDebugTemporalReplay(t *testing.T) {
//	replayer := worker.NewWorkflowReplayer()
//
//	replayer.RegisterWorkflow(temporal.Interpreter)
//
//	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "history/debug.json")
//	assertions := assert.New(t)
//	assertions.Nil(err)
//}
