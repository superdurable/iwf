// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iwf

import (
	"fmt"
	"github.com/superdurable/iwf/sdk-go/gen/iwfidl"
	"time"
)

const DateTimeFormat = "2006-01-02T15:04:05-07:00"

type PersistenceFieldDef struct {
	Key       string
	FieldType PersistenceFieldType
	// SearchAttributeType is optional and only required for PersistenceFieldTypeSearchAttribute
	SearchAttributeType *iwfidl.SearchAttributeValueType
}

type PersistenceFieldType string

const (
	PersistenceFieldTypeDataObject      PersistenceFieldType = "DataAttribute"
	PersistenceFieldTypeSearchAttribute PersistenceFieldType = "SearchAttribute"
)

func DataAttributeDef(key string) PersistenceFieldDef {
	return PersistenceFieldDef{
		Key:       key,
		FieldType: PersistenceFieldTypeDataObject,
	}
}
func SearchAttributeDef(key string, saType iwfidl.SearchAttributeValueType) PersistenceFieldDef {
	return PersistenceFieldDef{
		Key:                 key,
		FieldType:           PersistenceFieldTypeSearchAttribute,
		SearchAttributeType: &saType,
	}
}

func getSearchAttributeValue(sa iwfidl.SearchAttribute) (interface{}, error) {
	switch *sa.ValueType {
	case iwfidl.TEXT, iwfidl.KEYWORD:
		return *sa.StringValue, nil
	case iwfidl.KEYWORD_ARRAY:
		return sa.StringArrayValue, nil
	case iwfidl.DOUBLE:
		return *sa.DoubleValue, nil
	case iwfidl.BOOL:
		return *sa.BoolValue, nil
	case iwfidl.DATETIME:
		t, err := time.Parse(DateTimeFormat, *sa.StringValue)
		if err != nil {
			return nil, err
		}
		return t, nil
	case iwfidl.INT:
		return *sa.IntegerValue, nil
	default:
		return nil, fmt.Errorf("unsupported search attribute type %v", sa.GetValueType())
	}
}
