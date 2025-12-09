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
	"math"
	"reflect"
)

// UintValuer [struct] for uint valuer
type UintValuer struct {
}

// GetValue [method] for get uint64 value
func (uv *UintValuer) GetValue(data interface{}, name string) (uint64, error) {
	if name == "" {
		return uv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return math.MaxUint64, err
	}

	return uv.getValueFromReflect(value)
}

func (uv *UintValuer) getValueFromData(data interface{}) (uint64, error) {
	switch value := data.(type) {
	case reflect.Value:
		return uv.getValueFromReflect(&value)
	default:
		valueRef := reflect.ValueOf(value)
		return uv.getValueFromReflect(&valueRef)
	}
}

func (uv *UintValuer) getValueFromReflect(value *reflect.Value) (uint64, error) {
	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint(), nil
	default:
		return math.MaxUint64, fmt.Errorf("get reflect uint value failed: the type [%v] not uint", value.Kind())
	}
}
