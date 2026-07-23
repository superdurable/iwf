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

package retry

import (
	"github.com/superdurable/iwf/gen/iwfidl"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/cadence/workflow"
	"time"
)

func ConvertCadenceWorkflowRetryPolicy(policy *iwfidl.WorkflowRetryPolicy) *workflow.RetryPolicy {
	if policy == nil {
		return nil
	}
	fillWorkflowRetryPolicyDefault(policy)

	return &workflow.RetryPolicy{
		InitialInterval:    time.Second * time.Duration(policy.GetInitialIntervalSeconds()),
		MaximumInterval:    time.Second * time.Duration(policy.GetMaximumIntervalSeconds()),
		MaximumAttempts:    policy.GetMaximumAttempts(),
		BackoffCoefficient: float64(policy.GetBackoffCoefficient()),
	}
}

func ConvertCadenceActivityRetryPolicy(policy *iwfidl.RetryPolicy) *workflow.RetryPolicy {
	if policy == nil {
		return nil
	}
	fillActivityRetryPolicyDefault(policy)

	// in Cadence, ExpirationInterval is the timeout include all retries
	expirationInterval := time.Duration(0)
	if policy.GetMaximumAttemptsDurationSeconds() > 0 {
		expirationInterval = time.Second * time.Duration(policy.GetMaximumAttemptsDurationSeconds())
	} else {
		// unlimited to match Temporal
		expirationInterval = time.Hour * 24 * 365 * 1
	}

	return &workflow.RetryPolicy{
		InitialInterval:    time.Second * time.Duration(policy.GetInitialIntervalSeconds()),
		MaximumInterval:    time.Second * time.Duration(policy.GetMaximumIntervalSeconds()),
		MaximumAttempts:    policy.GetMaximumAttempts(),
		BackoffCoefficient: float64(policy.GetBackoffCoefficient()),
		ExpirationInterval: expirationInterval,
	}
}

func ConvertTemporalWorkflowRetryPolicy(policy *iwfidl.WorkflowRetryPolicy) *temporal.RetryPolicy {
	if policy == nil {
		return nil
	}
	fillWorkflowRetryPolicyDefault(policy)

	return &temporal.RetryPolicy{
		InitialInterval:    time.Second * time.Duration(policy.GetInitialIntervalSeconds()),
		MaximumInterval:    time.Second * time.Duration(policy.GetMaximumIntervalSeconds()),
		MaximumAttempts:    policy.GetMaximumAttempts(),
		BackoffCoefficient: float64(policy.GetBackoffCoefficient()),
	}
}

func ConvertTemporalActivityRetryPolicy(policy *iwfidl.RetryPolicy) *temporal.RetryPolicy {
	if policy == nil {
		return nil
	}
	fillActivityRetryPolicyDefault(policy)

	return &temporal.RetryPolicy{
		InitialInterval:    time.Second * time.Duration(policy.GetInitialIntervalSeconds()),
		MaximumInterval:    time.Second * time.Duration(policy.GetMaximumIntervalSeconds()),
		MaximumAttempts:    policy.GetMaximumAttempts(),
		BackoffCoefficient: float64(policy.GetBackoffCoefficient()),
	}
}

func fillActivityRetryPolicyDefault(policy *iwfidl.RetryPolicy) {
	if policy.InitialIntervalSeconds == nil {
		policy.InitialIntervalSeconds = iwfidl.PtrInt32(1)
	}
	if policy.BackoffCoefficient == nil {
		policy.BackoffCoefficient = iwfidl.PtrFloat32(2)
	}
	if policy.MaximumIntervalSeconds == nil {
		policy.MaximumIntervalSeconds = iwfidl.PtrInt32(100)
	}
	if policy.MaximumAttempts == nil {
		policy.MaximumAttempts = iwfidl.PtrInt32(0)
	}
	if policy.MaximumAttemptsDurationSeconds == nil {
		policy.MaximumAttemptsDurationSeconds = iwfidl.PtrInt32(0)
	}
}

func fillWorkflowRetryPolicyDefault(policy *iwfidl.WorkflowRetryPolicy) {
	if policy.InitialIntervalSeconds == nil {
		policy.InitialIntervalSeconds = iwfidl.PtrInt32(1)
	}
	if policy.BackoffCoefficient == nil {
		policy.BackoffCoefficient = iwfidl.PtrFloat32(2)
	}
	if policy.MaximumIntervalSeconds == nil {
		policy.MaximumIntervalSeconds = iwfidl.PtrInt32(100)
	}
	if policy.MaximumAttempts == nil {
		policy.MaximumAttempts = iwfidl.PtrInt32(0)
	}
}
