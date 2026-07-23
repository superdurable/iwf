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

package temporal

import (
	"context"
	"fmt"
	"log"

	"github.com/superdurable/iwf/config"
	uclient "github.com/superdurable/iwf/service/client"
	"github.com/superdurable/iwf/service/common/blobstore"
	"github.com/superdurable/iwf/service/common/workerclient"
	"github.com/superdurable/iwf/service/interpreter"
	"github.com/superdurable/iwf/service/interpreter/env"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

type InterpreterWorker struct {
	temporalClient client.Client
	worker         worker.Worker
	taskQueue      string
	workerPool     *workerclient.Pool
	internalClient *workerclient.Internal
}

func NewInterpreterWorker(
	config config.Config, temporalClient client.Client, taskQueue string, memoEncryption bool,
	memoEncryptionConverter converter.DataConverter, unifiedClient uclient.UnifiedClient,
	store blobstore.BlobStore,
) *InterpreterWorker {
	pool, internal := interpreter.NewWorkerClients(config)
	env.SetSharedEnv(config, memoEncryption, memoEncryptionConverter, unifiedClient, taskQueue, store, pool, internal)

	return &InterpreterWorker{
		temporalClient: temporalClient,
		taskQueue:      taskQueue,
		workerPool:     pool,
		internalClient: internal,
	}
}

func (iw *InterpreterWorker) Close() {
	if iw.worker != nil {
		iw.worker.Stop()
	}
	if iw.workerPool != nil {
		iw.workerPool.Close()
	}
	if iw.internalClient != nil {
		iw.internalClient.Close()
	}
	iw.temporalClient.Close()
}

func (iw *InterpreterWorker) StartWithStickyCacheDisabledForTest() {
	iw.start(true)
}

func (iw *InterpreterWorker) Start() {
	iw.start(false)
}

func (iw *InterpreterWorker) start(disableStickyCache bool) {
	cfg := env.GetSharedConfig()
	var options worker.Options

	if cfg.Interpreter.Temporal != nil && cfg.Interpreter.Temporal.WorkerOptions != nil {
		options = *cfg.Interpreter.Temporal.WorkerOptions
	}

	if options.MaxConcurrentActivityTaskPollers == 0 {
		options.MaxConcurrentActivityTaskPollers = 10
	}
	if options.MaxConcurrentWorkflowTaskPollers == 0 {
		options.MaxConcurrentWorkflowTaskPollers = 10
	}

	if disableStickyCache {
		worker.SetStickyWorkflowCacheSize(0)
		fmt.Println("Temporal worker: Sticky cache disabled")
	}

	iw.worker = worker.New(iw.temporalClient, iw.taskQueue, options)
	worker.EnableVerboseLogging(cfg.Interpreter.VerboseDebug)

	iw.worker.RegisterWorkflow(Interpreter)
	iw.worker.RegisterWorkflow(BlobStoreCleanup)
	iw.worker.RegisterActivity(interpreter.InvokeWaitForMethod)
	iw.worker.RegisterActivity(interpreter.InvokeExecuteMethod)
	iw.worker.RegisterActivity(interpreter.DumpFlowForContinueAsNew)
	iw.worker.RegisterActivity(interpreter.InvokeWorkerRpcActivity)
	iw.worker.RegisterActivity(interpreter.CleanupBlobStore)

	err := iw.worker.Start()
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}

	if cfg.ExternalStorage.Enabled {
		for _, storeCfg := range cfg.ExternalStorage.SupportedStorages {
			if storeCfg.CleanupCronSchedule != "" {
				err = env.GetUnifiedClient().StartBlobStoreCleanupWorkflow(
					context.Background(), iw.taskQueue,
					"blobstore-cleanup-"+storeCfg.StorageId,
					storeCfg.CleanupCronSchedule,
					storeCfg.StorageId)
				if err != nil {
					if env.GetUnifiedClient().IsWorkflowAlreadyStartedError(err) {
						continue
					}
					log.Fatalln("Unable to start blobstore cleanup workflow", err)
				}
			}
		}
	}
}
