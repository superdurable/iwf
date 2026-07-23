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
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/common/ptr"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvalidAnyConditionCombination(t *testing.T) {
	timers, channels := createConditions()
	waiting := &iwfpb.WaitingCondition{
		WaitingConditionType: iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ANY_COMBINATION_COMPLETED,
		TimerConditions:      timers,
		ChannelConditions:    channels,
		ConditionCombinations: []*iwfpb.ConditionCombination{
			{ConditionIds: []string{"timer-cmd1", "signal-cmd1"}},
			{ConditionIds: []string{"timer-cmd1", "invalid"}},
		},
	}

	err := validateWaitingCondition(waiting)
	require.Error(t, err)
	require.Equal(t,
		"ANY_COMBINATION_COMPLETED condition ids must exist on timer/channel conditions",
		err.Error())
}

func TestValidAnyConditionCombination(t *testing.T) {
	timers, channels := createConditions()
	waiting := &iwfpb.WaitingCondition{
		WaitingConditionType: iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ANY_COMBINATION_COMPLETED,
		TimerConditions:      timers,
		ChannelConditions:    channels,
		ConditionCombinations: []*iwfpb.ConditionCombination{
			{ConditionIds: []string{"timer-cmd1", "signal-cmd1"}},
			{ConditionIds: []string{"timer-cmd1", "internal-cmd1"}},
		},
	}

	require.NoError(t, validateWaitingCondition(waiting))
}

func TestValidateWaitingConditionChannelBounds(t *testing.T) {
	waiting := &iwfpb.WaitingCondition{
		WaitingConditionType: iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ALL_COMPLETED,
		ChannelConditions: []*iwfpb.ChannelCondition{
			{ConditionId: "c1", ChannelName: "ch", AtLeast: ptr.Any(int32(2)), AtMost: ptr.Any(int32(1))},
		},
	}
	require.Error(t, validateWaitingCondition(waiting))
}

func TestValidateStepDecisionEmpty(t *testing.T) {
	require.Error(t, validateStepDecision(nil))
	require.Error(t, validateStepDecision(&iwfpb.StepDecision{}))
	require.NoError(t, validateStepDecision(&iwfpb.StepDecision{
		NextSteps: []*iwfpb.StepMovement{{StepType: "s"}},
	}))
}

func createConditions() ([]*iwfpb.TimerCondition, []*iwfpb.ChannelCondition) {
	timers := []*iwfpb.TimerCondition{
		{ConditionId: "timer-cmd1", DurationSeconds: 86400 * 365},
	}
	channels := []*iwfpb.ChannelCondition{
		{ConditionId: "signal-cmd1", ChannelName: "test-signal-name1"},
		{ConditionId: "internal-cmd1", ChannelName: "test-internal-name1"},
	}
	return timers, channels
}

func TestComposeGRPCError_LocalActivity_LongMessage(t *testing.T) {
	longMsg := strings.Repeat("a", 1000)
	provider := &recordingActivityProvider{ret: errors.New("app-err")}

	err := composeGRPCError(true, provider, errors.New(longMsg), "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "1st-attempt-failure", provider.errType)
	require.Equal(t, longMsg[:50]+"...", provider.details)
}

func TestComposeGRPCError_RegularActivity_LongMessage(t *testing.T) {
	longMsg := strings.Repeat("a", 1000)
	provider := &recordingActivityProvider{ret: errors.New("app-err")}

	err := composeGRPCError(false, provider, errors.New(longMsg), "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "test-error-type", provider.errType)
	require.Equal(t, longMsg[:500]+"...", provider.details)
}

func TestComposeGRPCError_LocalActivity_ShortMessage(t *testing.T) {
	shortMsg := strings.Repeat("a", 40)
	provider := &recordingActivityProvider{ret: errors.New("app-err")}

	err := composeGRPCError(true, provider, errors.New(shortMsg), "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "1st-attempt-failure", provider.errType)
	require.Equal(t, shortMsg, provider.details)
}

func TestComposeGRPCError_RegularActivity_ShortMessage(t *testing.T) {
	shortMsg := strings.Repeat("a", 40)
	provider := &recordingActivityProvider{ret: errors.New("app-err")}

	err := composeGRPCError(false, provider, errors.New(shortMsg), "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "test-error-type", provider.errType)
	require.Equal(t, shortMsg, provider.details)
}

func TestComposeGRPCError_LocalActivity_StatusError(t *testing.T) {
	provider := &recordingActivityProvider{ret: errors.New("app-err")}
	stErr := status.Error(codes.Unavailable, "worker down")

	err := composeGRPCError(true, provider, stErr, "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "1st-attempt-failure", provider.errType)
	require.Equal(t, "code: Unavailable, msg: worker down", provider.details)
}

func TestComposeGRPCError_RegularActivity_StatusErrorLong(t *testing.T) {
	provider := &recordingActivityProvider{ret: errors.New("app-err")}
	longDetail := strings.Repeat("b", 600)
	stErr := status.Error(codes.Internal, longDetail)

	err := composeGRPCError(false, provider, stErr, "test-error-type")
	require.Equal(t, provider.ret, err)
	require.Equal(t, "test-error-type", provider.errType)
	expected := fmt.Sprintf("code: Internal, msg: %s", longDetail)
	require.Equal(t, expected[:500]+"...", provider.details)
}

type recordingActivityProvider struct {
	errType string
	details interface{}
	ret     error
}

func (r *recordingActivityProvider) GetLogger(context.Context) interfaces.UnifiedLogger {
	panic("unexpected")
}
func (r *recordingActivityProvider) NewApplicationError(errType string, details interface{}) error {
	r.errType = errType
	r.details = details
	return r.ret
}
func (r *recordingActivityProvider) GetActivityInfo(context.Context) interfaces.ActivityInfo {
	panic("unexpected")
}
func (r *recordingActivityProvider) RecordHeartbeat(context.Context, ...interface{}) {}
