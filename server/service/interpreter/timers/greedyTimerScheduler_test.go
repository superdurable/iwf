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
	"errors"
	"testing"
	"time"

	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/cont"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	awaitErr       error
	now            int64
	newTimerCalled bool
}

func (m *mockProvider) Await(_ interfaces.UnifiedContext, _ func() bool) error {
	return m.awaitErr
}
func (m *mockProvider) Now(_ interfaces.UnifiedContext) time.Time {
	return time.Unix(m.now, 0)
}
func (m *mockProvider) GoNamed(_ interfaces.UnifiedContext, _ string, f func(interfaces.UnifiedContext)) {
	f(nil)
}
func (m *mockProvider) NewTimer(_ interfaces.UnifiedContext, _ time.Duration) interfaces.Future {
	m.newTimerCalled = true
	return nil
}

// The rest are not needed for this test
func (m *mockProvider) NewApplicationError(string, interface{}) error { return nil }
func (m *mockProvider) IsApplicationError(error) bool                 { return false }
func (m *mockProvider) GetWorkflowInfo(interfaces.UnifiedContext) interfaces.WorkflowInfo {
	return interfaces.WorkflowInfo{}
}
func (m *mockProvider) GetSearchAttributes(interfaces.UnifiedContext, []iwfidl.SearchAttributeKeyAndType) (map[string]iwfidl.SearchAttribute, error) {
	return nil, nil
}
func (m *mockProvider) UpsertSearchAttributes(interfaces.UnifiedContext, map[string]interface{}) error {
	return nil
}
func (m *mockProvider) UpsertMemo(interfaces.UnifiedContext, map[string]iwfidl.EncodedObject) error {
	return nil
}
func (m *mockProvider) SetQueryHandler(interfaces.UnifiedContext, string, interface{}) error {
	return nil
}
func (m *mockProvider) SetRpcUpdateHandler(interfaces.UnifiedContext, string, interfaces.UnifiedRpcValidator, interfaces.UnifiedRpcHandler) error {
	return nil
}
func (m *mockProvider) ExtendContextWithValue(interfaces.UnifiedContext, string, interface{}) interfaces.UnifiedContext {
	return nil
}
func (m *mockProvider) GetThreadCount() int                   { return 0 }
func (m *mockProvider) GetPendingThreadNames() map[string]int { return nil }
func (m *mockProvider) WithActivityOptions(interfaces.UnifiedContext, interfaces.ActivityOptions) interfaces.UnifiedContext {
	return nil
}
func (m *mockProvider) ExecuteActivity(interface{}, bool, interfaces.UnifiedContext, interface{}, ...interface{}) error {
	return nil
}
func (m *mockProvider) ExecuteLocalActivity(interface{}, interfaces.UnifiedContext, interface{}, ...interface{}) error {
	return nil
}
func (m *mockProvider) IsReplaying(interfaces.UnifiedContext) bool           { return false }
func (m *mockProvider) Sleep(interfaces.UnifiedContext, time.Duration) error { return nil }
func (m *mockProvider) GetSignalChannel(interfaces.UnifiedContext, string) interfaces.ReceiveChannel {
	return nil
}
func (m *mockProvider) GetContextValue(interfaces.UnifiedContext, string) interface{} { return nil }
func (m *mockProvider) GetVersion(interfaces.UnifiedContext, string, int, int) int    { return 0 }
func (m *mockProvider) GetUnhandledSignalNames(interfaces.UnifiedContext) []string    { return nil }
func (m *mockProvider) GetBackendType() service.BackendType                           { return "" }
func (m *mockProvider) GetLogger(interfaces.UnifiedContext) interfaces.UnifiedLogger  { return nil }
func (m *mockProvider) NewInterpreterContinueAsNewError(interfaces.UnifiedContext, service.InterpreterWorkflowInput) error {
	return nil
}

func TestPruneToNextTimer_PrunesCorrectly_WithTwoScheduled(t *testing.T) {
	ts := &timerScheduler{
		pendingScheduling: []*service.TimerInfo{
			{FiringUnixTimestampSeconds: 1751395615, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395955, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395755, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395555, Status: service.TimerPending},
		},
		providerScheduledTimerUnixTs: []int64{1751395955, 1751395555},
	}

	pruned := ts.pruneToNextTimer(1751395755)
	assert.NotNil(t, pruned)
	assert.Equal(t, int64(1751395955), pruned.FiringUnixTimestampSeconds)
	assert.Equal(t, []int64{1751395955}, ts.providerScheduledTimerUnixTs)
	if assert.Equal(t, 2, len(ts.pendingScheduling)) {
		assert.Equal(t, int64(1751395615), ts.pendingScheduling[0].FiringUnixTimestampSeconds)
		assert.Equal(t, int64(1751395955), ts.pendingScheduling[1].FiringUnixTimestampSeconds)
	}
}

func TestPruneToNextTimer_PrunesCorrectly_WithOneScheduled(t *testing.T) {
	ts := &timerScheduler{
		pendingScheduling: []*service.TimerInfo{
			{FiringUnixTimestampSeconds: 1751395615, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395955, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395755, Status: service.TimerPending},
			{FiringUnixTimestampSeconds: 1751395555, Status: service.TimerPending},
		},
		providerScheduledTimerUnixTs: []int64{1751395555},
	}

	pruned := ts.pruneToNextTimer(1751395755)
	assert.NotNil(t, pruned)
	assert.Equal(t, int64(1751395955), pruned.FiringUnixTimestampSeconds)
	assert.Equal(t, []int64(nil), ts.providerScheduledTimerUnixTs)
	if assert.Equal(t, 2, len(ts.pendingScheduling)) {
		assert.Equal(t, int64(1751395615), ts.pendingScheduling[0].FiringUnixTimestampSeconds)
		assert.Equal(t, int64(1751395955), ts.pendingScheduling[1].FiringUnixTimestampSeconds)
	}
}

func TestStartGreedyTimerScheduler_AwaitErrorBreaksLoop(t *testing.T) {
	provider := &mockProvider{awaitErr: errors.New("test error"), now: 100}
	// Should not panic or loop forever
	_ = startGreedyTimerScheduler(nil, provider, (*cont.ContinueAsNewCounter)(nil))
}
