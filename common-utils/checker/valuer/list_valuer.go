// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package valuer

import (
	"fmt"
	"reflect"
)

// ListValuer [struct] for list valuer
type ListValuer struct {
}

// GetValue [method] for get value
func (lv *ListValuer) GetValue(data interface{}, name string) (*reflect.Value, error) {
	if name == "" {
		return lv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return nil, err
	}
	return lv.getValueFromReflect(value)
}

func (lv *ListValuer) getValueFromData(data interface{}) (*reflect.Value, error) {
	switch value := data.(type) {
	case reflect.Value:
		return lv.getValueFromReflect(&value)
	default:
		valueRef := reflect.ValueOf(value)
		return lv.getValueFromReflect(&valueRef)
	}
}

func (lv *ListValuer) getValueFromReflect(value *reflect.Value) (*reflect.Value, error) {
	valueType := value.Kind()
	if valueType != reflect.Array && valueType != reflect.Slice {
		return nil, fmt.Errorf("the value type [%v] not array or slice", valueType)
	}
	return value, nil
}
