//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils offer the some utils for certificate handling
package utils

import "reflect"

// IsNil check whether the interface is nil, including type or data is nil
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	defer func() {
		recover()
	}()
	return reflect.ValueOf(i).IsNil()
}
