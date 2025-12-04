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

// GetReflectValueByName [method] for get reflect value
func GetReflectValueByName(inputStruct interface{}, name string) (*reflect.Value, error) {
	var value reflect.Value
	switch reflectValue := inputStruct.(type) {
	case *reflect.Value:
		value = *reflectValue
	case reflect.Value:
		value = reflectValue
	default:
		value = reflect.ValueOf(inputStruct)
	}

	if value.Kind() == reflect.Struct {
		retValue := value.FieldByName(name)
		if !retValue.IsValid() {
			return nil, &FieldNotExistErr{name: name}
		}
		if retValue.Kind() == reflect.Ptr {
			if retValue.IsNil() {
				return nil, &FieldNotExistErr{name: name}
			}
			retValue = retValue.Elem()
		}
		return &retValue, nil
	}
	return nil, fmt.Errorf("just supported struct for reflect value, not [%v]", value.Kind())
}
