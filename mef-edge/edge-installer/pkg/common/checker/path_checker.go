// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker this file for base path check method
package checker

import (
	"strings"

	"edge-installer/pkg/common/constants"
)

// IsPathValid check whether the path is valid
func IsPathValid(path string) bool {
	if len(path) > constants.MaxPathLength || !RegexStringChecker(path, constants.PathMatchStr) ||
		strings.Contains(path, "..") {
		return false
	}
	return true
}
