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
	"errors"
	"fmt"
	"reflect"
)

// StringValuer [struct] for string valuer
type StringValuer struct {
}

// GetValue [method] for get string value
func (sv *StringValuer) GetValue(data interface{}, name string) (string, error) {
	if name == "" {
		return sv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return "", err
	}

	return sv.getValueFromReflect(value)
}

func (sv *StringValuer) getValueFromData(data interface{}) (string, error) {
	switch value := data.(type) {
	case string:
		return value, nil
	case reflect.Value:
		return sv.getValueFromReflect(&value)
	default:
		return "", errors.New("the input data not string or reflect.Value type")
	}
}

func (sv *StringValuer) getValueFromReflect(value *reflect.Value) (string, error) {
	if value.Kind() == reflect.String {
		return value.String(), nil
	}
	return "", fmt.Errorf("get reflect string value failed: the type [%v] not string", value.Kind())
}
