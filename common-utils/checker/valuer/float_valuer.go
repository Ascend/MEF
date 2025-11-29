// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
