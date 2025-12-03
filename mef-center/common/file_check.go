// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
