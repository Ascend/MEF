// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
