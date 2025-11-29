// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for string exclude checker
package checker

import (
	"huawei.com/mindx/common/checker/valuer"
)

// GetSnChecker [method] for get checker to check serial number
func GetSnChecker(field string, required bool) *RegChecker {
	return &RegChecker{
		filed:    field,
		reg:      `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}
