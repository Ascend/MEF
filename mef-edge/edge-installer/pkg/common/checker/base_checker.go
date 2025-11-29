// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker this file for base check method
package checker

import (
	"regexp"
)

// RegexStringChecker check string validity according to the matchStr
func RegexStringChecker(str, matchStr string) bool {
	strSlice := regexp.MustCompile(matchStr)
	return strSlice.MatchString(str)
}

// IntChecker check value range validity
func IntChecker(value, min, max int) bool {
	return value >= min && value <= max
}
