// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package valuer of checker module
package valuer

import "fmt"

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
	_, ok := err.(*FieldNotExistErr)
	return ok
}
