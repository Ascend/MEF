// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
