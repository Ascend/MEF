// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
