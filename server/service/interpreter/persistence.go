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
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/common/mapper"
	"github.com/superdurable/iwf/service/interpreter/interfaces"
)

type PersistenceManager struct {
	provider interfaces.WorkflowProvider

	attributes map[string]*iwfpb.Value

	lockedKeys map[string]bool
}

type attributeMutation struct {
	key   string
	value *iwfpb.Value
}

func NewPersistenceManager(
	provider interfaces.WorkflowProvider,
	initialAttributes []*iwfpb.KV,
) (*PersistenceManager, error) {
	if provider == nil {
		panic("PersistenceManager requires a WorkflowProvider")
	}

	attributes := make(map[string]*iwfpb.Value, len(initialAttributes))
	seenKeys := make(map[string]struct{}, len(initialAttributes))
	for index, attribute := range initialAttributes {
		if err := validateKV(attribute); err != nil {
			return nil, fmt.Errorf("initial attribute %d: %w", index, err)
		}
		if _, duplicated := seenKeys[attribute.GetKey()]; duplicated {
			return nil, fmt.Errorf("duplicate attribute key %q", attribute.GetKey())
		}
		seenKeys[attribute.GetKey()] = struct{}{}
		if isNullValue(attribute.GetValue()) {
			continue
		}
		attributes[attribute.GetKey()] = attribute.GetValue()
	}

	return &PersistenceManager{
		provider:   provider,
		attributes: attributes,
		lockedKeys: map[string]bool{},
	}, nil
}

func (p *PersistenceManager) ApplyAttributeWrites(
	ctx interfaces.UnifiedContext,
	writes []*iwfpb.AttributeWrite,
) (bool, error) {
	if len(writes) == 0 {
		return true, nil
	}

	mutations, indexedUpdates, applicable, err := p.planAttributeWrites(writes)
	if err != nil {
		return false, err
	}
	if !applicable {
		return false, nil
	}
	if len(mutations) == 0 && len(indexedUpdates) == 0 {
		return true, nil
	}
	if len(indexedUpdates) > 0 {
		if err := p.provider.UpsertSearchAttributes(ctx, indexedUpdates); err != nil {
			return false, fmt.Errorf("upsert indexed attributes: %w", err)
		}
	}

	for _, mutation := range mutations {
		if mutation.value == nil {
			delete(p.attributes, mutation.key)
		} else {
			p.attributes[mutation.key] = mutation.value
		}
	}
	return true, nil
}

func (p *PersistenceManager) planAttributeWrites(
	writes []*iwfpb.AttributeWrite,
) ([]attributeMutation, map[string]interface{}, bool, error) {
	seenKeys := make(map[string]struct{}, len(writes))
	for index, write := range writes {
		if err := validateAttributeWrite(write); err != nil {
			return nil, nil, false, fmt.Errorf("attribute %d: %w", index, err)
		}
		if _, duplicated := seenKeys[write.GetKey()]; duplicated {
			return nil, nil, false, fmt.Errorf("duplicate attribute key %q", write.GetKey())
		}
		seenKeys[write.GetKey()] = struct{}{}
	}
	for _, write := range writes {
		if p.lockedKeys[write.GetKey()] {
			return nil, nil, false, nil
		}
	}

	mutations := make([]attributeMutation, 0, len(writes))
	indexedUpdates := make(map[string]interface{})
	for _, write := range writes {
		if err := addIndexedUpdate(indexedUpdates, write); err != nil {
			return nil, nil, false, fmt.Errorf("attribute %q: %w", write.GetKey(), err)
		}

		existing, exists := p.attributes[write.GetKey()]
		if isNullValue(write.GetValue()) {
			if exists {
				mutations = append(mutations, attributeMutation{key: write.GetKey()})
			}
			continue
		}
		if exists && attributeValuesEqual(existing, write.GetValue()) {
			continue
		}
		mutations = append(mutations, attributeMutation{key: write.GetKey(), value: write.GetValue()})
	}
	return mutations, indexedUpdates, true, nil
}

func validateAttributeWrite(write *iwfpb.AttributeWrite) error {
	if write == nil {
		return fmt.Errorf("write is nil")
	}
	if write.GetKey() == "" {
		return fmt.Errorf("key is empty")
	}
	if write.GetValue() == nil || write.GetValue().GetKind() == nil {
		return fmt.Errorf("value is missing")
	}
	return nil
}

func validateKV(attribute *iwfpb.KV) error {
	if attribute == nil {
		return fmt.Errorf("attribute is nil")
	}
	if attribute.GetKey() == "" {
		return fmt.Errorf("key is empty")
	}
	if attribute.GetValue() == nil || attribute.GetValue().GetKind() == nil {
		return fmt.Errorf("value is missing")
	}
	return nil
}

func addIndexedUpdate(updates map[string]interface{}, write *iwfpb.AttributeWrite) error {
	if indexedKey(write) == "" {
		return nil
	}
	if isNullValue(write.GetValue()) {
		updates[indexedKey(write)] = nil
		return nil
	}
	mapped, err := mapper.MapAttributeWritesToSearchAttributes([]*iwfpb.AttributeWrite{write})
	if err != nil {
		return err
	}
	for key, value := range mapped {
		updates[key] = value
	}
	return nil
}

func indexedKey(write *iwfpb.AttributeWrite) string {
	if write.GetIndexConfig() == nil || !write.GetIndexConfig().GetEnable() {
		return ""
	}
	return mapper.IndexKey(write)
}

func (p *PersistenceManager) TryLockKeys(sortedUniqueKeys []string) bool {
	validateSortedUniqueKeys(sortedUniqueKeys)
	for _, key := range sortedUniqueKeys {
		if p.lockedKeys[key] {
			return false
		}
	}

	for _, key := range sortedUniqueKeys {
		p.lockedKeys[key] = true
	}
	return true
}

func (p *PersistenceManager) UnlockKeys(sortedUniqueKeys []string) {
	validateSortedUniqueKeys(sortedUniqueKeys)
	for _, key := range sortedUniqueKeys {
		if !p.lockedKeys[key] {
			panic(fmt.Sprintf("attribute lock is not held for key %q", key))
		}
	}
	for _, key := range sortedUniqueKeys {
		delete(p.lockedKeys, key)
	}
}

func (p *PersistenceManager) HasAnyLock() bool {
	return len(p.lockedKeys) > 0
}

func validateSortedUniqueKeys(keys []string) {
	if len(keys) == 0 {
		panic("attribute lock requires at least one key")
	}
	for index, key := range keys {
		if key == "" {
			panic("attribute lock key is empty")
		}
		if index > 0 && keys[index-1] >= key {
			panic("attribute lock keys must be sorted and unique")
		}
	}
}

func (p *PersistenceManager) GetAttribute(key string) (*iwfpb.Value, bool) {
	value, ok := p.attributes[key]
	return value, ok
}

func (p *PersistenceManager) GetAllAttributes() []*iwfpb.KV {
	keys := sortedAttributeKeys(p.attributes)
	attributes := make([]*iwfpb.KV, 0, len(keys))
	for _, key := range keys {
		attributes = append(attributes, &iwfpb.KV{Key: key, Value: p.attributes[key]})
	}
	return attributes
}

func (p *PersistenceManager) GetAttributes(
	request *iwfpb.GetAttributesQueryRequest,
) *iwfpb.GetAttributesQueryResponse {
	if request == nil {
		panic("GetAttributes requires a request")
	}

	keys := request.GetKeys()
	if request.GetAllKeys() {
		keys = sortedAttributeKeys(p.attributes)
	} else {
		keys = sortedUniqueStrings(keys)
	}

	attributes := make([]*iwfpb.KV, 0, len(keys))
	for _, key := range keys {
		value, ok := p.attributes[key]
		if !ok {
			continue
		}
		attributes = append(attributes, &iwfpb.KV{Key: key, Value: value})
	}
	return &iwfpb.GetAttributesQueryResponse{Attributes: attributes}
}

func sortedAttributeKeys(attributes map[string]*iwfpb.Value) []string {
	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedUniqueStrings(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		unique[value] = struct{}{}
	}
	values = make([]string, 0, len(unique))
	for value := range unique {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func attributeValuesEqual(left, right *iwfpb.Value) bool {
	if left == nil || right == nil {
		return left == right
	}
	switch leftKind := left.GetKind().(type) {
	case *iwfpb.Value_InternalBlobIdForStringValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_InternalBlobIdForStringValue)
		return ok && leftKind.InternalBlobIdForStringValue == rightKind.InternalBlobIdForStringValue
	case *iwfpb.Value_InternalBlobIdForObjValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_InternalBlobIdForObjValue)
		return ok && leftKind.InternalBlobIdForObjValue == rightKind.InternalBlobIdForObjValue
	case *iwfpb.Value_StringValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_StringValue)
		return ok && leftKind.StringValue == rightKind.StringValue
	case *iwfpb.Value_ObjValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_ObjValue)
		if !ok {
			return false
		}
		return leftKind.ObjValue.GetEncoding() == rightKind.ObjValue.GetEncoding() &&
			bytes.Equal(leftKind.ObjValue.GetPayload(), rightKind.ObjValue.GetPayload())
	case *iwfpb.Value_IntValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_IntValue)
		return ok && leftKind.IntValue == rightKind.IntValue
	case *iwfpb.Value_DoubleValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_DoubleValue)
		return ok && math.Float64bits(leftKind.DoubleValue) == math.Float64bits(rightKind.DoubleValue)
	case *iwfpb.Value_BoolValue:
		rightKind, ok := right.GetKind().(*iwfpb.Value_BoolValue)
		return ok && leftKind.BoolValue == rightKind.BoolValue
	case *iwfpb.Value_NullValue:
		_, ok := right.GetKind().(*iwfpb.Value_NullValue)
		return ok
	default:
		return false
	}
}

func isNullValue(value *iwfpb.Value) bool {
	if value == nil {
		return false
	}
	_, ok := value.GetKind().(*iwfpb.Value_NullValue)
	return ok
}
