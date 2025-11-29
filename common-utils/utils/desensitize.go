// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package utils for
package utils

import (
	"fmt"
	"regexp"
)

const (
	sensitiveInfoWildcard = "***"
)

// TrimInfoFromError trim sensitive information from an error, return new error
func TrimInfoFromError(err error) error {
	if err == nil || err.Error() == "" {
		return err
	}
	patterns := []string{
		`:\/\/([a-zA-Z0-9.:\/_\-=?%@&])*`,
	}
	tempMsg := []byte(err.Error())
	for _, pattern := range patterns {
		req, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		tempMsg = req.ReplaceAll(tempMsg, []byte(sensitiveInfoWildcard))
	}

	return fmt.Errorf(string(tempMsg))
}
