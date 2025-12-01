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

// FloatValuer [struct] for float valuer
type FloatValuer struct {
}

// GetValue [method] for get float64 value
func (fv *FloatValuer) GetValue(data interface{}, name string) (float64, error) {
	if name == "" {
		return fv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return math.MaxFloat64, err
	}

	return fv.getValueFromReflect(value)
}

func (fv *FloatValuer) getValueFromData(data interface{}) (float64, error) {
	switch value := data.(type) {
	case reflect.Value:
		return fv.getValueFromReflect(&value)
	default:
		valueRef := reflect.ValueOf(value)
		return fv.getValueFromReflect(&valueRef)
	}
}

func (fv *FloatValuer) getValueFromReflect(value *reflect.Value) (float64, error) {
	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	default:
		return math.MaxFloat64, fmt.Errorf("get reflect float value failed: the type [%v] not float", value.Kind())
	}
}
