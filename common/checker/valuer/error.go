// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package valuer

import (
	"fmt"
	"reflect"
)

// FieldNotExistErr [struct] for Field not exist error
type FieldNotExistErr struct {
	name string
}

// Error [method] for return error message
func (e *FieldNotExistErr) Error() string {
	return fmt.Sprintf("Field [%s] not found", e.name)
}

// CheckIsFieldNotExistErr [method] for check the error is FieldNotExistErr type
func CheckIsFieldNotExistErr(err error) bool {
	errTypeName := reflect.TypeOf(err).Elem().Name()
	return reflect.TypeOf(FieldNotExistErr{}).Name() == errTypeName
}
