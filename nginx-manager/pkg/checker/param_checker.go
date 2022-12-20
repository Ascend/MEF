// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker this file is for check parameter
package checker

import (
	"fmt"
	"regexp"
)

// CheckType the type for check type
type CheckType int32

const (
	// Env 环境变量
	Env CheckType = 0
	// NginxConfig nginx配置文件
	NginxConfig CheckType = 1
)

var checkers = map[CheckType]func(param interface{}) error{
	Env:         checkEnv,
	NginxConfig: checkNginxConfig,
}

// Check do the check job
func Check(cType CheckType, param interface{}) error {
	if c, ok := checkers[cType]; ok {
		return c(param)
	}
	return fmt.Errorf("no checker found %d", cType)
}

// RegexStringChecker use regexp to check str
func RegexStringChecker(str, matchStr string) bool {
	strSlice := regexp.MustCompile(matchStr)
	return strSlice.MatchString(str)
}
