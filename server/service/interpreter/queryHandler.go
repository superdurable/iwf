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
	"github.com/superdurable/iwf/gen/iwfidl"
	"github.com/superdurable/iwf/service"
	"github.com/superdurable/iwf/service/interpreter/config"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

func SetQueryHandlers(
	ctx interfaces.UnifiedContext,
	provider interfaces.WorkflowProvider,
	timerProcessor interfaces.TimerProcessor,
	persistenceManager *PersistenceManager,
	internalChannel *InternalChannel,
	signalReceiver *SignalReceiver,
	continueAsNewer *ContinueAsNewer,
	workflowConfiger *config.WorkflowConfiger,
	basicInfo service.BasicInfo,
) error {
	err := provider.SetQueryHandler(ctx, service.GetDataAttributesWorkflowQueryType, func(req service.GetDataAttributesQueryRequest) (service.GetDataAttributesQueryResponse, error) {
		dos := persistenceManager.GetDataAttributesByKey(req)
		return dos, nil
	})
	if err != nil {
		return err
	}
	err = provider.SetQueryHandler(ctx, service.GetSearchAttributesWorkflowQueryType, func() ([]iwfidl.SearchAttribute, error) {
		return persistenceManager.GetAllSearchAttributes(), nil
	})
	if err != nil {
		return err
	}
	err = continueAsNewer.SetQueryHandlersForContinueAsNew(ctx)
	if err != nil {
		return err
	}
	err = provider.SetQueryHandler(ctx, service.DebugDumpQueryType, func() (*service.DebugDumpResponse, error) {
		return &service.DebugDumpResponse{
			Config:                     workflowConfiger.Get(),
			Snapshot:                   continueAsNewer.GetSnapshot(),
			FiringTimersUnixTimestamps: timerProcessor.GetTimerStartedUnixTimestamps(),
		}, nil
	})
	if err != nil {
		return err
	}
	err = provider.SetQueryHandler(ctx, service.PrepareRpcQueryType, func(req service.PrepareRpcQueryRequest) (service.PrepareRpcQueryResponse, error) {
		info := provider.GetWorkflowInfo(ctx) // TODO use firstRunId instead

		return service.PrepareRpcQueryResponse{
			DataObjects:              persistenceManager.LoadDataAttributes(ctx, req.DataObjectsLoadingPolicy),
			SearchAttributes:         persistenceManager.LoadSearchAttributes(ctx, req.SearchAttributesLoadingPolicy),
			WorkflowRunId:            info.WorkflowExecution.RunID,
			WorkflowStartedTimestamp: info.WorkflowStartTime.Unix(),
			IwfWorkflowType:          basicInfo.IwfWorkflowType,
			IwfWorkerUrl:             basicInfo.IwfWorkerUrl,
			SignalChannelInfo:        signalReceiver.GetInfos(),
			InternalChannelInfo:      internalChannel.GetInfos(),
		}, nil
	})
	if err != nil {
		return err
	}

	err = provider.SetQueryHandler(ctx, service.GetCurrentTimerInfosQueryType, func() (service.GetCurrentTimerInfosQueryResponse, error) {
		return service.GetCurrentTimerInfosQueryResponse{
			StateExecutionCurrentTimerInfos: timerProcessor.GetTimerInfos(),
		}, nil
	})

	if err != nil {
		return err
	}

	return nil
}
