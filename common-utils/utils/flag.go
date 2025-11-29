// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
