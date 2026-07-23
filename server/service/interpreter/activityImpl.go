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
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/common/blobstore"
	"github.com/superdurable/iwf/service/common/event"
	"github.com/superdurable/iwf/service/common/log"
	"github.com/superdurable/iwf/service/common/rpc"
	"github.com/superdurable/iwf/service/common/workerclient"
	"github.com/superdurable/iwf/service/interpreter/env"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	errTypeWorkerAPIFail  = "WorkerAPIFailure"
	errTypeServerInternal = "ServerInternalError"
)

// InvokeWaitForMethod calls WorkerService.InvokeWaitForMethod.
func InvokeWaitForMethod(
	ctx context.Context, input *iwfpb.InvokeWaitForMethodActivityInput,
) (*iwfpb.InvokeWaitForMethodActivityOutput, error) {
	startMs := time.Now().UnixMilli()
	backendType := backendTypeFromProto(input.GetBackendType())
	provider := interfaces.GetActivityProviderByType(backendType)
	logger := provider.GetLogger(ctx)
	logger.Info("InvokeWaitForMethodActivity", "input", log.ToJsonAndTruncateForLogging(input))

	activityInfo := provider.GetActivityInfo(ctx)
	req := input.GetRequest()
	if req == nil {
		return nil, provider.NewApplicationError(errTypeWorkerAPIFail, "nil InvokeWaitForMethodRequest")
	}
	if req.Context == nil {
		req.Context = &iwfpb.Context{}
	}
	req.Context.Attempt = activityInfo.Attempt
	req.Context.FirstAttemptTimestamp = activityInfo.ScheduledTime.Unix()

	if err := hydrateWorkerRequestValues(ctx, req.GetStepInput(), req.GetAttributes()); err != nil {
		logLocalActivityWarn(logger, activityInfo, "InvokeWaitForMethod", req.GetContext().GetStepExecutionId(), err)
		return nil, err
	}

	client, callCtx, release, err := env.GetWorkerPool().Acquire(ctx, input.GetWorkerTarget())
	if err != nil {
		return nil, composeWorkerDialError(provider, activityInfo.IsLocalActivity, err)
	}
	defer release()

	resp, err := client.InvokeWaitForMethod(callCtx, req)
	printDebugMsg(logger, err, input.GetWorkerTarget())
	if err != nil {
		appErr := composeGRPCError(activityInfo.IsLocalActivity, provider, err, errTypeWorkerAPIFail)
		emitStepEvent(req, activityInfo, "WAIT_FOR_ATTEMPT_FAIL", startMs, appErr)
		logLocalActivityWarn(logger, activityInfo, "InvokeWaitForMethod", req.GetContext().GetStepExecutionId(), appErr)
		return nil, appErr
	}
	if err := validateWaitingCondition(resp.GetWaitingCondition()); err != nil {
		appErr := provider.NewApplicationError(errTypeWorkerAPIFail, err.Error())
		emitStepEvent(req, activityInfo, "WAIT_FOR_ATTEMPT_FAIL", startMs, appErr)
		return nil, appErr
	}
	if err := validateWorkerWaitForResponse(resp); err != nil {
		return nil, provider.NewApplicationError(errTypeWorkerAPIFail, err.Error())
	}

	if activityInfo.IsLocalActivity {
		resp.LocalActivityInput = composeInputForDebug(req.GetContext().GetStepExecutionId())
	}
	if err := offloadWorkerAttributeWrites(ctx, resp.GetUpsertAttributes(), activityInfo.WorkflowExecution.ID); err != nil {
		return nil, err
	}

	emitStepEvent(req, activityInfo, "WAIT_FOR_ATTEMPT_SUCC", startMs, nil)
	return &iwfpb.InvokeWaitForMethodActivityOutput{Response: resp}, nil
}

// InvokeExecuteMethod calls WorkerService.InvokeExecuteMethod.
func InvokeExecuteMethod(
	ctx context.Context, input *iwfpb.InvokeExecuteMethodActivityInput,
) (*iwfpb.InvokeExecuteMethodActivityOutput, error) {
	startMs := time.Now().UnixMilli()
	backendType := backendTypeFromProto(input.GetBackendType())
	provider := interfaces.GetActivityProviderByType(backendType)
	logger := provider.GetLogger(ctx)
	logger.Info("InvokeExecuteMethodActivity", "input", log.ToJsonAndTruncateForLogging(input))

	activityInfo := provider.GetActivityInfo(ctx)
	req := input.GetRequest()
	if req == nil {
		return nil, provider.NewApplicationError(errTypeWorkerAPIFail, "nil InvokeExecuteMethodRequest")
	}
	if req.Context == nil {
		req.Context = &iwfpb.Context{}
	}
	req.Context.Attempt = activityInfo.Attempt
	req.Context.FirstAttemptTimestamp = activityInfo.ScheduledTime.Unix()

	stepInputCopy := cloneValue(req.GetStepInput())
	if err := hydrateWorkerRequestValues(ctx, req.GetStepInput(), req.GetAttributes()); err != nil {
		logLocalActivityWarn(logger, activityInfo, "InvokeExecuteMethod", req.GetContext().GetStepExecutionId(), err)
		return nil, err
	}
	if err := blobstore.HydrateKVs(ctx, req.GetStepExeLocals(), env.GetBlobStore()); err != nil {
		return nil, err
	}

	client, callCtx, release, err := env.GetWorkerPool().Acquire(ctx, input.GetWorkerTarget())
	if err != nil {
		return nil, composeWorkerDialError(provider, activityInfo.IsLocalActivity, err)
	}
	defer release()

	resp, err := client.InvokeExecuteMethod(callCtx, req)
	printDebugMsg(logger, err, input.GetWorkerTarget())
	if err != nil {
		appErr := composeGRPCError(activityInfo.IsLocalActivity, provider, err, errTypeWorkerAPIFail)
		emitExecuteEvent(req, activityInfo, "EXECUTE_ATTEMPT_FAIL", startMs, appErr)
		logLocalActivityWarn(logger, activityInfo, "InvokeExecuteMethod", req.GetContext().GetStepExecutionId(), appErr)
		return nil, appErr
	}
	if err := validateStepDecision(resp.GetStepDecision()); err != nil {
		appErr := provider.NewApplicationError(errTypeWorkerAPIFail, err.Error())
		emitExecuteEvent(req, activityInfo, "EXECUTE_ATTEMPT_FAIL", startMs, appErr)
		return nil, appErr
	}
	if err := validateWorkerExecuteResponse(resp); err != nil {
		return nil, provider.NewApplicationError(errTypeWorkerAPIFail, err.Error())
	}

	if activityInfo.IsLocalActivity {
		resp.LocalActivityInput = composeInputForDebug(req.GetContext().GetStepExecutionId())
	}
	if err := offloadNextStepInputs(ctx, resp.GetStepDecision(), stepInputCopy, activityInfo.WorkflowExecution.ID); err != nil {
		return nil, err
	}
	if err := offloadWorkerAttributeWrites(ctx, resp.GetUpsertAttributes(), activityInfo.WorkflowExecution.ID); err != nil {
		return nil, err
	}

	emitExecuteEvent(req, activityInfo, "EXECUTE_ATTEMPT_SUCC", startMs, nil)
	return &iwfpb.InvokeExecuteMethodActivityOutput{Response: resp}, nil
}

// DumpFlowForContinueAsNew pages ContinueAsNewDump via InternalService.
func DumpFlowForContinueAsNew(
	ctx context.Context, input *iwfpb.DumpFlowForContinueAsNewActivityInput,
) (*iwfpb.DumpFlowForContinueAsNewActivityOutput, error) {
	backendType := backendTypeFromProto(input.GetBackendType())
	provider := interfaces.GetActivityProviderByType(backendType)
	logger := provider.GetLogger(ctx)
	logger.Info("DumpFlowForContinueAsNewActivity", "input", log.ToJsonAndTruncateForLogging(input))

	internal := env.GetInternalClient()
	if internal == nil {
		return nil, provider.NewApplicationError(errTypeServerInternal, "internal client is nil")
	}
	client, callCtx, err := internal.Client(ctx)
	if err != nil {
		return nil, provider.NewApplicationError(errTypeServerInternal, err.Error())
	}
	resp, err := client.DumpFlowForContinueAsNew(callCtx, input.GetRequest())
	if err != nil {
		return nil, composeGRPCError(provider.GetActivityInfo(ctx).IsLocalActivity, provider, err, errTypeServerInternal)
	}
	return &iwfpb.DumpFlowForContinueAsNewActivityOutput{Response: resp}, nil
}

// InvokeWorkerRpcActivity wraps rpc.InvokeWorkerRpc for the activity worker.
func InvokeWorkerRpcActivity(
	ctx context.Context, input *iwfpb.InvokeWorkerRPCActivityInput,
) (*iwfpb.InvokeWorkerRPCActivityOutput, error) {
	backendType := backendTypeFromProto(input.GetBackendType())
	provider := interfaces.GetActivityProviderByType(backendType)
	logger := provider.GetLogger(ctx)
	logger.Info("InvokeWorkerRpcActivity", "input", log.ToJsonAndTruncateForLogging(input))
	activityInfo := provider.GetActivityInfo(ctx)
	sharedCfg := env.GetSharedConfig()

	resp, statusErr := rpc.InvokeWorkerRpc(
		ctx,
		env.GetWorkerPool(),
		input.GetRpcPrep(),
		input.GetRequest(),
		sharedCfg.Api.EffectiveMaxWaitSeconds(),
		env.GetBlobStore(),
		sharedCfg.ExternalStorage,
	)
	out := &iwfpb.InvokeWorkerRPCActivityOutput{Response: resp}
	if statusErr != nil {
		out.Error = &iwfpb.InterpreterError{
			GrpcCode: int32(statusErr.Code),
			Error:    statusErr.Error,
		}
	}
	if activityInfo.IsLocalActivity {
		payload := log.ToJsonAndTruncateForLogging(out)
		if threshold := sharedCfg.Interpreter.LogLocalActivityThresholdBytes; threshold > 0 && len(payload) >= threshold {
			logger.Warn("InvokeWorkerRpc local activity return",
				"workflowId", activityInfo.WorkflowExecution.ID,
				"payloadSize", len(payload))
		}
	}
	return out, nil
}

// CleanupBlobStore deletes blob objects for flows that no longer exist in the backend.
func CleanupBlobStore(
	ctx context.Context, input *iwfpb.CleanupBlobStoreActivityInput,
) (*iwfpb.CleanupBlobStoreActivityOutput, error) {
	store := env.GetBlobStore()
	backendType := backendTypeFromProto(input.GetBackendType())
	provider := interfaces.GetActivityProviderByType(backendType)
	logger := provider.GetLogger(ctx)
	logger.Info("CleanupBlobStore started")
	client := env.GetUnifiedClient()

	var continueToken *string
	var totalDeleted int32
	for {
		listOutput, err := store.ListWorkflowPaths(ctx, blobstore.ListObjectPathsInput{
			StoreId:           input.GetStoreId(),
			ContinuationToken: continueToken,
		})
		if err != nil {
			return &iwfpb.CleanupBlobStoreActivityOutput{TotalDeleted: totalDeleted}, err
		}
		continueToken = listOutput.ContinuationToken
		for _, workflowPath := range listOutput.WorkflowPaths {
			_, valid := blobstore.ExtractYyyymmddToUnixSeconds(workflowPath)
			if !valid {
				logger.Info("CleanupBlobStore skipped workflow path", "path", workflowPath)
				continue
			}
			flowId := blobstore.MustExtractWorkflowId(workflowPath)
			_, err := client.DescribeWorkflowExecution(ctx, flowId, "", nil)
			if client.IsNotFoundError(err) {
				if err := store.DeleteWorkflowObjects(ctx, input.GetStoreId(), workflowPath); err != nil {
					logger.Error("CleanupBlobStore failed to delete workflow objects", "workflowPath", workflowPath, "error", err)
					return &iwfpb.CleanupBlobStoreActivityOutput{TotalDeleted: totalDeleted}, err
				}
				totalDeleted++
				logger.Info("CleanupBlobStore deleted workflow objects", "workflowPath", workflowPath)
			} else if err != nil {
				logger.Error("CleanupBlobStore failed to describe workflow", "workflowPath", workflowPath, "error", err)
				return &iwfpb.CleanupBlobStoreActivityOutput{TotalDeleted: totalDeleted}, err
			}
			provider.RecordHeartbeat(ctx)
		}
		if continueToken == nil {
			break
		}
	}
	logger.Info("CleanupBlobStore completed", "totalDeleted", totalDeleted)
	return &iwfpb.CleanupBlobStoreActivityOutput{TotalDeleted: totalDeleted}, nil
}

func backendTypeFromProto(t iwfpb.BackendType) service.BackendType {
	switch t {
	case iwfpb.BackendType_BACKEND_TYPE_CADENCE:
		return service.BackendTypeCadence
	case iwfpb.BackendType_BACKEND_TYPE_TEMPORAL:
		return service.BackendTypeTemporal
	default:
		panic(fmt.Sprintf("unspecified backend type %v", t))
	}
}

func hydrateWorkerRequestValues(ctx context.Context, stepInput *iwfpb.Value, attrs []*iwfpb.KV) error {
	if err := blobstore.HydrateValue(ctx, stepInput, env.GetBlobStore()); err != nil {
		return err
	}
	return blobstore.HydrateKVs(ctx, attrs, env.GetBlobStore())
}

func offloadWorkerAttributeWrites(ctx context.Context, writes []*iwfpb.AttributeWrite, flowId string) error {
	cfg := env.GetSharedConfig()
	if !cfg.ExternalStorage.Enabled || env.GetBlobStore() == nil {
		return nil
	}
	return blobstore.OffloadLargeAttributeWrites(
		ctx, writes, flowId, cfg.ExternalStorage.ThresholdInBytes, env.GetBlobStore(), true,
	)
}

func offloadNextStepInputs(
	ctx context.Context, decision *iwfpb.StepDecision, previousInput *iwfpb.Value, flowId string,
) error {
	cfg := env.GetSharedConfig()
	if decision == nil || !cfg.ExternalStorage.Enabled || env.GetBlobStore() == nil {
		return nil
	}
	for _, step := range decision.GetNextSteps() {
		if step == nil || step.GetStepInput() == nil {
			continue
		}
		if err := maybeReuseOrOffloadStepInput(ctx, step.StepInput, previousInput, flowId); err != nil {
			return err
		}
	}
	return nil
}

func maybeReuseOrOffloadStepInput(
	ctx context.Context, next *iwfpb.Value, previous *iwfpb.Value, flowId string,
) error {
	cfg := env.GetSharedConfig()
	threshold := cfg.ExternalStorage.ThresholdInBytes
	if next == nil {
		return nil
	}
	// Reuse prior blob id when the concrete payload still matches the hydrated previous value.
	if previous != nil {
		switch prev := previous.GetKind().(type) {
		case *iwfpb.Value_InternalBlobIdForStringValue:
			if s, ok := next.GetKind().(*iwfpb.Value_StringValue); ok && len(s.StringValue) > threshold {
				// Cannot reuse without comparing to hydrated bytes; fall through to offload.
				_ = prev
			}
		}
	}
	return blobstore.OffloadLargeValue(ctx, next, flowId, threshold, env.GetBlobStore(), true)
}

func cloneValue(v *iwfpb.Value) *iwfpb.Value {
	if v == nil {
		return nil
	}
	return proto.Clone(v).(*iwfpb.Value)
}

func validateWorkerWaitForResponse(resp *iwfpb.InvokeWaitForMethodResponse) error {
	if resp == nil {
		return fmt.Errorf("nil InvokeWaitForMethodResponse")
	}
	if err := workerclient.RejectWorkerAttributeWriteBlobIDs(resp.GetUpsertAttributes()); err != nil {
		return err
	}
	if err := workerclient.RejectWorkerKVBlobIDs(resp.GetUpsertStepExeLocals()); err != nil {
		return err
	}
	return workerclient.RejectWorkerKVBlobIDs(resp.GetRecordEvents())
}

func validateWorkerExecuteResponse(resp *iwfpb.InvokeExecuteMethodResponse) error {
	if resp == nil {
		return fmt.Errorf("nil InvokeExecuteMethodResponse")
	}
	if err := workerclient.RejectWorkerAttributeWriteBlobIDs(resp.GetUpsertAttributes()); err != nil {
		return err
	}
	if err := workerclient.RejectWorkerKVBlobIDs(resp.GetUpsertStepExeLocals()); err != nil {
		return err
	}
	if err := workerclient.RejectWorkerKVBlobIDs(resp.GetRecordEvents()); err != nil {
		return err
	}
	if decision := resp.GetStepDecision(); decision != nil {
		for _, step := range decision.GetNextSteps() {
			if step == nil {
				continue
			}
			if err := workerclient.RejectWorkerBlobIDs(step.GetStepInput()); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateStepDecision(decision *iwfpb.StepDecision) error {
	if decision == nil || len(decision.GetNextSteps()) == 0 {
		return fmt.Errorf("empty step decision is not supported")
	}
	return nil
}

func validateWaitingCondition(waiting *iwfpb.WaitingCondition) error {
	if waiting == nil {
		return nil
	}
	hasConditions := len(waiting.GetTimerConditions())+len(waiting.GetChannelConditions()) > 0
	if !hasConditions {
		return nil
	}
	switch waiting.GetWaitingConditionType() {
	case iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ALL_COMPLETED,
		iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ANY_COMPLETED:
		// ok
	case iwfpb.WaitingConditionType_WAITING_CONDITION_TYPE_ANY_COMBINATION_COMPLETED:
		if len(waiting.GetConditionCombinations()) == 0 {
			return fmt.Errorf("ANY_COMBINATION_COMPLETED requires non-empty condition_combinations and condition ids")
		}
		for _, cmd := range waiting.GetTimerConditions() {
			if cmd.GetConditionId() == "" {
				return fmt.Errorf("ANY_COMBINATION_COMPLETED requires condition_id on every timer condition")
			}
		}
		for _, cmd := range waiting.GetChannelConditions() {
			if cmd.GetConditionId() == "" {
				return fmt.Errorf("ANY_COMBINATION_COMPLETED requires condition_id on every channel condition")
			}
			if err := validateChannelBounds(cmd); err != nil {
				return err
			}
		}
		if !areAllConditionCombinationIdsValid(waiting) {
			return fmt.Errorf("ANY_COMBINATION_COMPLETED condition ids must exist on timer/channel conditions")
		}
	default:
		return fmt.Errorf("unsupported waiting_condition_type %v", waiting.GetWaitingConditionType())
	}
	for _, cmd := range waiting.GetChannelConditions() {
		if err := validateChannelBounds(cmd); err != nil {
			return err
		}
	}
	return nil
}

func validateChannelBounds(cmd *iwfpb.ChannelCondition) error {
	if cmd == nil {
		return nil
	}
	if cmd.AtLeast != nil && cmd.GetAtLeast() < 0 {
		return fmt.Errorf("channel condition at_least must be >= 0")
	}
	if cmd.AtMost != nil && cmd.AtLeast != nil && cmd.GetAtMost() > 0 && cmd.GetAtMost() < cmd.GetAtLeast() {
		return fmt.Errorf("channel condition at_most must be >= at_least when set")
	}
	return nil
}

func areAllConditionCombinationIdsValid(waiting *iwfpb.WaitingCondition) bool {
	ids := listConditionIds(waiting)
	for _, combo := range waiting.GetConditionCombinations() {
		for _, id := range combo.GetConditionIds() {
			if !slices.Contains(ids, id) {
				return false
			}
		}
	}
	return true
}

func listConditionIds(waiting *iwfpb.WaitingCondition) []string {
	var ids []string
	for _, cmd := range waiting.GetTimerConditions() {
		ids = append(ids, cmd.GetConditionId())
	}
	for _, cmd := range waiting.GetChannelConditions() {
		ids = append(ids, cmd.GetConditionId())
	}
	return ids
}

func composeInputForDebug(stepExeId string) string {
	return fmt.Sprintf("stepExeId: %s", stepExeId)
}

func printDebugMsg(logger interfaces.UnifiedLogger, err error, target string) {
	if os.Getenv(service.EnvNameDebugMode) != "" {
		logger.Info("check error at worker gRPC request", err, target)
	}
}

func composeWorkerDialError(provider interfaces.ActivityProvider, isLocal bool, err error) error {
	return composeGRPCError(isLocal, provider, err, errTypeWorkerAPIFail)
}

func composeGRPCError(isLocalActivity bool, provider interfaces.ActivityProvider, err error, errType string) error {
	msg := err.Error()
	if st, ok := status.FromError(err); ok {
		msg = fmt.Sprintf("code: %v, msg: %v", st.Code(), st.Message())
	}
	if isLocalActivity {
		errType = "1st-attempt-failure"
		msg = trimText(msg, 50)
	} else {
		msg = trimText(msg, 500)
	}
	return provider.NewApplicationError(errType, msg)
}

func trimText(msg string, maxLength int) string {
	if len(msg) > maxLength {
		return msg[:maxLength] + "..."
	}
	return msg
}

func logLocalActivityWarn(
	logger interfaces.UnifiedLogger,
	activityInfo interfaces.ActivityInfo, name, stepExeId string, err error,
) {
	_ = err
	cfg := env.GetSharedConfig()
	if !activityInfo.IsLocalActivity || cfg.Interpreter.LogLocalActivityThresholdBytes <= 0 {
		return
	}
	logger.Warn(name+" local activity return on error",
		"workflowId", activityInfo.WorkflowExecution.ID,
		"stepExecutionId", stepExeId)
}

func emitStepEvent(
	req *iwfpb.InvokeWaitForMethodRequest, activityInfo interfaces.ActivityInfo, eventType string, startMs int64, appErr error,
) {
	_ = appErr
	event.Handle(event.Event{
		FlowId:          activityInfo.WorkflowExecution.ID,
		RunId:           activityInfo.WorkflowExecution.RunID,
		FlowType:        req.GetFlowType(),
		StepType:        req.GetStepType(),
		StepExecutionId: req.GetContext().GetStepExecutionId(),
		EventType:       eventType,
	})
	_ = startMs
}

func emitExecuteEvent(
	req *iwfpb.InvokeExecuteMethodRequest, activityInfo interfaces.ActivityInfo, eventType string, startMs int64, appErr error,
) {
	_ = appErr
	_ = startMs
	event.Handle(event.Event{
		FlowId:          activityInfo.WorkflowExecution.ID,
		RunId:           activityInfo.WorkflowExecution.RunID,
		FlowType:        req.GetFlowType(),
		StepType:        req.GetStepType(),
		StepExecutionId: req.GetContext().GetStepExecutionId(),
		EventType:       eventType,
	})
}
