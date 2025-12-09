// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package valuer
package valuer

import (
	"errors"
	"fmt"
	"reflect"
)

// BoolValuer [struct] for boolean valuer
type BoolValuer struct {
}

// GetValue [method] for get bool value
func (bv *BoolValuer) GetValue(data interface{}, name string) (bool, error) {
	if name == "" {
		return bv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return false, err
	}

	return bv.getValueFromReflect(value)
}

func (bv *BoolValuer) getValueFromData(data interface{}) (bool, error) {
	switch value := data.(type) {
	case bool:
		return value, nil
	case reflect.Value:
		return bv.getValueFromReflect(&value)
	default:
		return false, errors.New("the input data not bool or reflect.Value type")
	}
}

func (bv *BoolValuer) getValueFromReflect(value *reflect.Value) (bool, error) {
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), nil
	default:
		return false, fmt.Errorf("get reflect bool value failed: the type [%v] not bool", value.Kind())
	}
}
