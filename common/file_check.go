// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import "huawei.com/mindx/common/checker"

const (
	fileNameRegex = "^[a-zA-Z0-9_/.-]{1,256}$"
)

// FileNameCheck [method] for check file name
func FileNameCheck(fileName string) checker.CheckResult {
	return checker.GetAndChecker(
		checker.GetRegChecker("", fileNameRegex, true),
		checker.GetStringExcludeChecker("", []string{".."}, true),
	).Check(fileName)
}
