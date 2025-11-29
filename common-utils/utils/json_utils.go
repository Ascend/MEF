// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
)

// ObjectConvert converts the generic types to req type
func ObjectConvert(input interface{}, req interface{}) error {
	reqType := reflect.TypeOf(req)
	reqValue := reflect.ValueOf(req)
	if reqType.Kind() == reflect.Pointer &&
		reflect.TypeOf(input).AssignableTo(reqType.Elem()) &&
		!reqValue.IsNil() &&
		reqValue.Elem().CanSet() {
		reqValue.Elem().Set(reflect.ValueOf(input))
		return nil
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return errors.New("encode error")
	}

	if err := json.Unmarshal(jsonBytes, req); err != nil {
		return errors.New("decode error")
	}
	return nil
}

// JsonConvert converts json to req type. This func only supports []byte, json.RawMessage, string
func JsonConvert(input interface{}, req interface{}) error {
	var reader io.Reader
	switch input.(type) {
	case json.RawMessage:
		reader = bytes.NewReader(input.(json.RawMessage))
	case []byte:
		reader = bytes.NewReader(input.([]byte))
	case string:
		reader = strings.NewReader(input.(string))
	default:
		return errors.New("type error")
	}
	dec := json.NewDecoder(reader)
	if err := dec.Decode(req); err != nil {
		return errors.New("decode error")
	}
	return nil
}
