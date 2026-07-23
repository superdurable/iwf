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

package env

import (
	"github.com/superdurable/iwf/config"
	uclient "github.com/superdurable/iwf/service/client"
	"github.com/superdurable/iwf/service/common/blobstore"
	"go.temporal.io/sdk/converter"
)

// this env package is for creating some handy global objects so that workflow worker can make use of,
// to avoid passing throw worker context(esp hard because we have to do it with both Cadence & Temporal)
// Note it's not a good practice to use global so should keep them minimum

var sharedConfig config.Config

var temporalDataConverter converter.DataConverter

var temporalMemoEncryption bool

var unifiedClient uclient.UnifiedClient

var taskQueue string

var blobStore blobstore.BlobStore

func SetSharedEnv(
	config config.Config,
	memoEncryption bool,
	temporalMemoEncryptionDataConverter converter.DataConverter,
	client uclient.UnifiedClient,
	queue string,
	store blobstore.BlobStore,
) {
	sharedConfig = config
	temporalDataConverter = temporalMemoEncryptionDataConverter
	temporalMemoEncryption = memoEncryption
	unifiedClient = client
	taskQueue = queue
	blobStore = store
}

func GetUnifiedClient() uclient.UnifiedClient {
	return unifiedClient
}

func GetTaskQueue() string {
	return taskQueue
}

func GetSharedConfig() config.Config {
	return sharedConfig
}

func CheckAndGetTemporalMemoEncryptionDataConverter() (converter.DataConverter, bool) {
	return temporalDataConverter, temporalMemoEncryption
}

func GetBlobStore() blobstore.BlobStore {
	return blobStore
}
