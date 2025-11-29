// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package util

import (
	"huawei.com/mindx/common/checker"
)

const (
	pathMatchStr   = "^/[a-z0-9A-Z_./-]{1,511}$"
	invalidPathStr = ".."
)

// GetPathChecker [method] return a checker which check out if the path is valid
func GetPathChecker(filed string, required bool) *checker.ModelChecker {
	modelChecker := checker.ModelChecker{
		Field:    "",
		Required: required,
		Checker: checker.GetAndChecker(
			checker.GetRegChecker(filed, pathMatchStr, true),
			checker.GetStringExcludeChecker(filed, []string{invalidPathStr}, true),
		),
	}
	return &modelChecker
}
