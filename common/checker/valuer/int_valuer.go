// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package valuer

import (
	"fmt"
	"math"
	"reflect"
)

// IntValuer [struct] for int valuer
type IntValuer struct {
}

// GetValue [method] for get int value
func (iv *IntValuer) GetValue(data interface{}, name string) (int64, error) {
	if name == "" {
		return iv.getValueFromData(data)
	}

	value, err := GetReflectValueByName(data, name)
	if err != nil {
		return math.MaxInt64, err
	}

	return iv.getValueFromReflect(value)
}

func (iv *IntValuer) getValueFromData(data interface{}) (int64, error) {
	switch value := data.(type) {
	case reflect.Value:
		return iv.getValueFromReflect(&value)
	default:
		valueRef := reflect.ValueOf(value)
		return iv.getValueFromReflect(&valueRef)
	}
}

func (iv *IntValuer) getValueFromReflect(value *reflect.Value) (int64, error) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	default:
		return math.MaxInt64, fmt.Errorf("get reflect int value failed: the type [%v] not int", value.Kind())
	}
}
