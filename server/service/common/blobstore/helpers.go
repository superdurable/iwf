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

package blobstore

import (
	"context"

	"github.com/superdurable/iwf/gen/iwfidl"
)

func WriteDataObjectsToExternalStorage(ctx context.Context, dataObjects []iwfidl.KeyValue, workflowId string, threashold int, blobStore BlobStore, isExternalStorageEnabled bool) error {
	if !isExternalStorageEnabled {
		return nil
	}

	for i := range dataObjects {
		if dataObjects[i].Value != nil && dataObjects[i].Value.Data != nil &&
			len(*dataObjects[i].Value.Data) > threashold {
			// Save data to external storage
			storeId, path, writeErr := blobStore.WriteObject(ctx, workflowId, *dataObjects[i].Value.Data)
			if writeErr != nil {
				return writeErr
			}
			dataObjects[i].Value.ExtStoreId = &storeId
			dataObjects[i].Value.ExtPath = &path
			dataObjects[i].Value.Data = nil // Clear data since it's now in external storage
		}
	}
	return nil
}

func LoadDataObjectsFromExternalStorage(ctx context.Context, dataObjects []iwfidl.KeyValue, blobStore BlobStore) error {
	for i := range dataObjects {
		if dataObjects[i].Value != nil && dataObjects[i].Value.ExtStoreId != nil && dataObjects[i].Value.ExtPath != nil {
			data, err := blobStore.ReadObject(ctx, *dataObjects[i].Value.ExtStoreId, *dataObjects[i].Value.ExtPath)
			if err != nil {
				return err
			}

			dataObjects[i].Value.Data = &data
			dataObjects[i].Value.ExtPath = nil
			dataObjects[i].Value.ExtStoreId = nil
		}
	}
	return nil
}
