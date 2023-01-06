// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package checker

type checkerIntf interface {
	// Check [interface method] for do check
	Check(data interface{}) CheckResult
}
