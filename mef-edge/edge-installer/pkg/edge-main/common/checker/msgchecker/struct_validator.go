// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgchecker for check struct field
package msgchecker

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const (
	defaultTagName = "validate"
	tagSeparator   = ";"
)

func isMatched(str, matchStr string) bool {
	if matchStr == "" {
		return true
	}
	reg, err := regexp.Compile(matchStr)
	if err != nil {
		return false
	}

	return reg.MatchString(str)
}

func parseMapPattern(str string) (string, string) {
	subStrings := strings.Split(str, tagSeparator)
	var keyPattern, valuePattern string
	if len(subStrings) > 0 {
		keyPattern = subStrings[0]
	}

	if len(subStrings) > 1 {
		valuePattern = subStrings[1]
	}

	return keyPattern, valuePattern
}

func checkSequenceField(fieldNames []string, pattern string, value reflect.Value) error {
	count := value.Len()
	for i := 0; i < count; i++ {
		if err := checkField(fieldNames, pattern, value.Index(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func checkMapField(fieldNames []string, pattern string, value reflect.Value) error {
	keys := value.MapKeys()
	for _, key := range keys {
		keyPattern, valuePattern := parseMapPattern(pattern)
		if key.Kind() == reflect.String {
			if !isMatched(key.String(), keyPattern) {
				return fmt.Errorf("check map field:[%s] key failed", strings.Join(fieldNames, "."))
			}
		}
		if err := checkField(fieldNames, valuePattern, value.MapIndex(key)); err != nil {
			return err
		}
	}

	return nil
}

func checkStructField(fieldNames []string, pattern string, value reflect.Value) error {
	for i := 0; i < value.NumField(); i++ {
		if !value.Type().Field(i).IsExported() {
			continue
		}

		pattern := value.Type().Field(i).Tag.Get(defaultTagName)

		if err := checkField(append(fieldNames, value.Type().Field(i).Name),
			pattern,
			value.Field(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func checkStringField(fieldNames []string, pattern string, value reflect.Value) error {
	if !isMatched(value.String(), pattern) {
		return fmt.Errorf("check field:[%s] failed", strings.Join(fieldNames, "."))
	}

	return nil
}

func checkPointerFiled(fieldNames []string, pattern string, value reflect.Value) error {
	if value.IsNil() {
		return nil
	}
	return checkField(fieldNames, pattern, value.Elem())
}

func checkInterfaceFiled(fieldNames []string, pattern string, value reflect.Value) error {
	return errors.New("struct interface filed type is not supported")
}
func checkFieldValue(fieldNames []string, pattern string, value reflect.Value) error {
	checkItems := map[reflect.Kind]func(fieldNames []string, pattern string, value reflect.Value) error{
		reflect.String:    checkStringField,
		reflect.Struct:    checkStructField,
		reflect.Map:       checkMapField,
		reflect.Slice:     checkSequenceField,
		reflect.Array:     checkSequenceField,
		reflect.Pointer:   checkPointerFiled,
		reflect.Interface: checkInterfaceFiled,
	}
	if check, ok := checkItems[value.Kind()]; ok {
		return check(fieldNames, pattern, value)
	}

	return nil
}

func checkField(fieldNames []string, pattern string, input interface{}) error {
	var value reflect.Value
	switch reflectValue := input.(type) {
	case string:
		return checkStringField(fieldNames, pattern, reflect.ValueOf(reflectValue))
	case *string:
		if reflectValue == nil {
			return nil
		}
		return checkStringField(fieldNames, pattern, reflect.ValueOf(*reflectValue))
	case *reflect.Value:
		if reflectValue.IsNil() {
			return nil
		}
		value = *reflectValue
	case reflect.Value:
		value = reflectValue
	default:
		value = reflect.ValueOf(input)
	}

	return checkFieldValue(fieldNames, pattern, value)
}

// validateStruct [method] check struct string para whether match the regex or not, the regex from tag, default tag is
// validate, if you want to check the map key and value, the key regex and value regex tag separate by ';',
// e.g: map[string]string `validate:"^[a-z]{1-9}$;^[0-9]{5}$"`
// If you need to do a complex verification, please combine validateStruct and gin binding or  validate package
// the regex not support contain ';'
func validateStruct(input interface{}) error {
	if input == nil {
		return nil
	}

	var err error
	func() {
		defer func() {
			if data := recover(); data != nil {
				err = fmt.Errorf("check struct error(%v)", data)
			}
		}()

		val := reflect.ValueOf(input)

		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}

		err = checkField([]string{reflect.TypeOf(val.Interface()).Name()}, "", val)
	}()

	return err
}
