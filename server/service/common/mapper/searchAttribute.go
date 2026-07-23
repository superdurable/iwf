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

package mapper

import (
	"fmt"
	"time"

	"github.com/superdurable/iwf/gen/iwfpb"
	"github.com/superdurable/iwf/service/common/timeparser"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
)

// InferIndexType returns the IndexType for an AttributeWrite, preferring IndexConfig then Value kind.
func InferIndexType(write *iwfpb.AttributeWrite) (iwfpb.IndexType, error) {
	if write == nil {
		return iwfpb.IndexType_INDEX_TYPE_UNSPECIFIED, fmt.Errorf("nil AttributeWrite")
	}
	if write.GetIndexConfig() != nil && write.GetIndexConfig().GetType() != iwfpb.IndexType_INDEX_TYPE_UNSPECIFIED {
		return write.GetIndexConfig().GetType(), nil
	}
	return InferIndexTypeFromValue(write.GetValue())
}

// InferIndexTypeFromValue picks KEYWORD for string/object; INT/DOUBLE/BOOL from scalars.
func InferIndexTypeFromValue(value *iwfpb.Value) (iwfpb.IndexType, error) {
	if value == nil {
		return iwfpb.IndexType_INDEX_TYPE_UNSPECIFIED, fmt.Errorf("nil Value")
	}
	switch value.GetKind().(type) {
	case *iwfpb.Value_StringValue, *iwfpb.Value_ObjValue, *iwfpb.Value_InternalBlobIdForStringValue, *iwfpb.Value_InternalBlobIdForObjValue:
		return iwfpb.IndexType_INDEX_TYPE_KEYWORD, nil
	case *iwfpb.Value_IntValue:
		return iwfpb.IndexType_INDEX_TYPE_INT, nil
	case *iwfpb.Value_DoubleValue:
		return iwfpb.IndexType_INDEX_TYPE_DOUBLE, nil
	case *iwfpb.Value_BoolValue:
		return iwfpb.IndexType_INDEX_TYPE_BOOL, nil
	case *iwfpb.Value_NullValue:
		return iwfpb.IndexType_INDEX_TYPE_UNSPECIFIED, fmt.Errorf("null Value is not indexable")
	default:
		return iwfpb.IndexType_INDEX_TYPE_UNSPECIFIED, fmt.Errorf("unsupported Value kind for indexing")
	}
}

// IndexKey returns IndexConfig.index_key when set, else the attribute key.
func IndexKey(write *iwfpb.AttributeWrite) string {
	if write.GetIndexConfig() != nil && write.GetIndexConfig().GetIndexKey() != "" {
		return write.GetIndexConfig().GetIndexKey()
	}
	return write.GetKey()
}

// MapAttributeWritesToSearchAttributes encodes indexed AttributeWrites for Temporal/Cadence upsert.
// Writes without IndexConfig.enable (or with enable=false) are skipped.
// null_value writes are skipped here; callers delete keys separately.
func MapAttributeWritesToSearchAttributes(writes []*iwfpb.AttributeWrite) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for _, write := range writes {
		if write == nil || write.GetValue() == nil {
			continue
		}
		if _, ok := write.GetValue().GetKind().(*iwfpb.Value_NullValue); ok {
			continue
		}
		cfg := write.GetIndexConfig()
		if cfg == nil || !cfg.GetEnable() {
			continue
		}
		indexType, err := InferIndexType(write)
		if err != nil {
			return nil, err
		}
		key := IndexKey(write)
		backendVal, err := valueToSearchBackend(write.GetValue(), indexType)
		if err != nil {
			return nil, err
		}
		res[key] = backendVal
	}
	return res, nil
}

func valueToSearchBackend(value *iwfpb.Value, indexType iwfpb.IndexType) (interface{}, error) {
	switch indexType {
	case iwfpb.IndexType_INDEX_TYPE_KEYWORD, iwfpb.IndexType_INDEX_TYPE_TEXT:
		if s, ok := value.GetKind().(*iwfpb.Value_StringValue); ok {
			return s.StringValue, nil
		}
		return nil, fmt.Errorf("KEYWORD/TEXT requires string_value")
	case iwfpb.IndexType_INDEX_TYPE_KEYWORD_ARRAY:
		// Stored as string_value JSON array is not in Value; use string for single keyword array entry via string.
		if s, ok := value.GetKind().(*iwfpb.Value_StringValue); ok {
			return []string{s.StringValue}, nil
		}
		return nil, fmt.Errorf("KEYWORD_ARRAY requires string_value (single element) in this mapper")
	case iwfpb.IndexType_INDEX_TYPE_INT:
		if v, ok := value.GetKind().(*iwfpb.Value_IntValue); ok {
			return v.IntValue, nil
		}
		return nil, fmt.Errorf("INT requires int_value")
	case iwfpb.IndexType_INDEX_TYPE_DOUBLE:
		if v, ok := value.GetKind().(*iwfpb.Value_DoubleValue); ok {
			return v.DoubleValue, nil
		}
		return nil, fmt.Errorf("DOUBLE requires double_value")
	case iwfpb.IndexType_INDEX_TYPE_BOOL:
		if v, ok := value.GetKind().(*iwfpb.Value_BoolValue); ok {
			return v.BoolValue, nil
		}
		return nil, fmt.Errorf("BOOL requires bool_value")
	case iwfpb.IndexType_INDEX_TYPE_DATETIME:
		if s, ok := value.GetKind().(*iwfpb.Value_StringValue); ok {
			t, err := timeparser.ParseTime(s.StringValue)
			if err != nil {
				return nil, err
			}
			return time.Unix(0, t), nil
		}
		return nil, fmt.Errorf("DATETIME requires string_value")
	default:
		return nil, fmt.Errorf("unsupported index type %v", indexType)
	}
}

// MapCadenceIndexedFieldsToValues decodes requested Cadence indexed fields into Values.
func MapCadenceIndexedFieldsToValues(
	searchAttributes *shared.SearchAttributes, indexedAttrTypes map[string]iwfpb.IndexType,
) (map[string]*iwfpb.Value, error) {
	if searchAttributes == nil || len(indexedAttrTypes) == 0 {
		return nil, nil
	}
	result := make(map[string]*iwfpb.Value, len(indexedAttrTypes))
	for key, indexType := range indexedAttrTypes {
		field, ok := searchAttributes.IndexedFields[key]
		if !ok {
			continue
		}
		var object interface{}
		if err := client.NewValue(field).Get(&object); err != nil {
			return nil, err
		}
		if object == nil {
			continue
		}
		val, err := backendObjectToValue(object, indexType, true)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

// MapTemporalIndexedFieldsToValues decodes requested Temporal indexed fields into Values.
func MapTemporalIndexedFieldsToValues(
	searchAttributes *common.SearchAttributes, indexedAttrTypes map[string]iwfpb.IndexType,
) (map[string]*iwfpb.Value, error) {
	if searchAttributes == nil || len(indexedAttrTypes) == 0 {
		return nil, nil
	}
	result := make(map[string]*iwfpb.Value, len(indexedAttrTypes))
	for key, indexType := range indexedAttrTypes {
		field, ok := searchAttributes.IndexedFields[key]
		if !ok {
			continue
		}
		var object interface{}
		if err := converter.GetDefaultDataConverter().FromPayload(field, &object); err != nil {
			return nil, err
		}
		if object == nil {
			continue
		}
		val, err := backendObjectToValue(object, indexType, false)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func backendObjectToValue(object interface{}, indexType iwfpb.IndexType, cadenceJSONNumber bool) (*iwfpb.Value, error) {
	switch indexType {
	case iwfpb.IndexType_INDEX_TYPE_KEYWORD, iwfpb.IndexType_INDEX_TYPE_TEXT, iwfpb.IndexType_INDEX_TYPE_DATETIME:
		s, ok := object.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for %v", indexType)
		}
		return &iwfpb.Value{Kind: &iwfpb.Value_StringValue{StringValue: s}}, nil
	case iwfpb.IndexType_INDEX_TYPE_INT:
		switch v := object.(type) {
		case int64:
			return &iwfpb.Value{Kind: &iwfpb.Value_IntValue{IntValue: v}}, nil
		case float64:
			return &iwfpb.Value{Kind: &iwfpb.Value_IntValue{IntValue: int64(v)}}, nil
		default:
			return nil, fmt.Errorf("expected int for INT")
		}
	case iwfpb.IndexType_INDEX_TYPE_DOUBLE:
		switch v := object.(type) {
		case float64:
			return &iwfpb.Value{Kind: &iwfpb.Value_DoubleValue{DoubleValue: v}}, nil
		case int64:
			return &iwfpb.Value{Kind: &iwfpb.Value_DoubleValue{DoubleValue: float64(v)}}, nil
		default:
			return nil, fmt.Errorf("expected float for DOUBLE")
		}
	case iwfpb.IndexType_INDEX_TYPE_BOOL:
		b, ok := object.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool for BOOL")
		}
		return &iwfpb.Value{Kind: &iwfpb.Value_BoolValue{BoolValue: b}}, nil
	case iwfpb.IndexType_INDEX_TYPE_KEYWORD_ARRAY:
		switch v := object.(type) {
		case []string:
			if len(v) == 1 {
				return &iwfpb.Value{Kind: &iwfpb.Value_StringValue{StringValue: v[0]}}, nil
			}
			return nil, fmt.Errorf("KEYWORD_ARRAY multi-element decode not represented in Value yet")
		case []interface{}:
			if len(v) == 1 {
				s, ok := v[0].(string)
				if !ok {
					return nil, fmt.Errorf("KEYWORD_ARRAY element not string")
				}
				return &iwfpb.Value{Kind: &iwfpb.Value_StringValue{StringValue: s}}, nil
			}
			_ = cadenceJSONNumber
			return nil, fmt.Errorf("KEYWORD_ARRAY multi-element decode not represented in Value yet")
		default:
			return nil, fmt.Errorf("expected string array for KEYWORD_ARRAY")
		}
	default:
		return nil, fmt.Errorf("unsupported index type %v", indexType)
	}
}
