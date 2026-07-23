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
	"fmt"

	"github.com/superdurable/iwf/gen/iwfpb"
)

// OffloadLargeAttributeWrites replaces oversized string/object Value arms with server-minted blob ids.
func OffloadLargeAttributeWrites(
	ctx context.Context,
	writes []*iwfpb.AttributeWrite,
	flowId string,
	threshold int,
	blobStore BlobStore,
	enabled bool,
) error {
	if !enabled || threshold <= 0 {
		return nil
	}
	for _, write := range writes {
		if write == nil || write.GetValue() == nil {
			continue
		}
		if err := offloadValue(ctx, write.Value, flowId, threshold, blobStore); err != nil {
			return err
		}
	}
	return nil
}

// OffloadLargeValue offloads a single Value when over threshold.
func OffloadLargeValue(
	ctx context.Context, value *iwfpb.Value, flowId string, threshold int, blobStore BlobStore, enabled bool,
) error {
	if !enabled || threshold <= 0 || value == nil {
		return nil
	}
	return offloadValue(ctx, value, flowId, threshold, blobStore)
}

func offloadValue(ctx context.Context, value *iwfpb.Value, flowId string, threshold int, blobStore BlobStore) error {
	switch kind := value.GetKind().(type) {
	case *iwfpb.Value_StringValue:
		if len(kind.StringValue) <= threshold {
			return nil
		}
		storeId, path, err := blobStore.WriteObject(ctx, flowId, kind.StringValue)
		if err != nil {
			return err
		}
		blobId := formatBlobId(storeId, path)
		value.Kind = &iwfpb.Value_InternalBlobIdForStringValue{InternalBlobIdForStringValue: blobId}
		return nil
	case *iwfpb.Value_ObjValue:
		if kind.ObjValue == nil || len(kind.ObjValue.GetPayload()) <= threshold {
			return nil
		}
		storeId, path, err := blobStore.WriteObject(ctx, flowId, string(kind.ObjValue.GetPayload()))
		if err != nil {
			return err
		}
		// Preserve encoding alongside blob id by storing "storeId|path|encoding" is avoided;
		// encoding stays only if we keep obj metadata — mint blob id and drop payload.
		_ = kind.ObjValue.GetEncoding()
		blobId := formatBlobId(storeId, path)
		value.Kind = &iwfpb.Value_InternalBlobIdForObjValue{InternalBlobIdForObjValue: blobId}
		return nil
	default:
		return nil
	}
}

// HydrateValues replaces internal_blob_id_for_* arms with concrete string/object values.
func HydrateValues(ctx context.Context, values []*iwfpb.Value, blobStore BlobStore) error {
	for _, value := range values {
		if err := HydrateValue(ctx, value, blobStore); err != nil {
			return err
		}
	}
	return nil
}

// HydrateAttributeWrites hydrates Value arms on AttributeWrites / KVs.
func HydrateAttributeWrites(ctx context.Context, writes []*iwfpb.AttributeWrite, blobStore BlobStore) error {
	for _, write := range writes {
		if write == nil {
			continue
		}
		if err := HydrateValue(ctx, write.GetValue(), blobStore); err != nil {
			return err
		}
	}
	return nil
}

// HydrateKVs hydrates Value arms on KV pairs.
func HydrateKVs(ctx context.Context, kvs []*iwfpb.KV, blobStore BlobStore) error {
	for _, kv := range kvs {
		if kv == nil {
			continue
		}
		if err := HydrateValue(ctx, kv.GetValue(), blobStore); err != nil {
			return err
		}
	}
	return nil
}

// HydrateValue hydrates a single Value in place.
func HydrateValue(ctx context.Context, value *iwfpb.Value, blobStore BlobStore) error {
	if value == nil {
		return nil
	}
	switch kind := value.GetKind().(type) {
	case *iwfpb.Value_InternalBlobIdForStringValue:
		storeId, path, err := parseBlobId(kind.InternalBlobIdForStringValue)
		if err != nil {
			return err
		}
		data, err := blobStore.ReadObject(ctx, storeId, path)
		if err != nil {
			return err
		}
		value.Kind = &iwfpb.Value_StringValue{StringValue: data}
		return nil
	case *iwfpb.Value_InternalBlobIdForObjValue:
		storeId, path, err := parseBlobId(kind.InternalBlobIdForObjValue)
		if err != nil {
			return err
		}
		data, err := blobStore.ReadObject(ctx, storeId, path)
		if err != nil {
			return err
		}
		value.Kind = &iwfpb.Value_ObjValue{ObjValue: &iwfpb.EncodedObject{Payload: []byte(data)}}
		return nil
	default:
		return nil
	}
}

func formatBlobId(storeId, path string) string {
	return storeId + "|" + path
}

func parseBlobId(blobId string) (storeId, path string, err error) {
	for i := 0; i < len(blobId); i++ {
		if blobId[i] == '|' {
			return blobId[:i], blobId[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid blob id %q", blobId)
}
