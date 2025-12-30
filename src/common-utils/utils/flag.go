// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils
package utils

import (
	"flag"
	"fmt"
	"os"
)

const maxArgsNum = 100

var (
	requiredFlags []string
)

// MarkFlagRequired marks specific field as required
func MarkFlagRequired(flagName string) {
	requiredFlags = append(requiredFlags, flagName)
}

// IsRequiredFlagNotFound checks whether the required flag is set.
func IsRequiredFlagNotFound() bool {
	if len(os.Args) > maxArgsNum {
		fmt.Println("parameters is too many")
		return true
	}
	seenFlags := make(map[string]struct{}, len(os.Args))
	flag.Visit(func(f *flag.Flag) {
		seenFlags[f.Name] = struct{}{}
	})
	for _, flagName := range requiredFlags {
		if _, ok := seenFlags[flagName]; !ok {
			return true
		}
	}
	return false
}

// IsFlagSet check whether the flag is set by user
func IsFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
