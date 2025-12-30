// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config for
package config

import (
	"errors"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
)

var authConfig AuthInfo

const (
	minTokenExpireTime = 1
	maxTokenExpireTime = 60
)

// AuthInfo [struct] for save auth config
type AuthInfo struct {
	TokenExpireTime int
}

// SetConfig to set auth config
func SetConfig(config AuthInfo) {
	authConfig = config
}

// GetAuthConfig to get auth config
func GetAuthConfig() AuthInfo {
	return authConfig
}

// CheckAuthConfig check auth config
func CheckAuthConfig(config AuthInfo) error {
	if result := checker.GetIntChecker("TokenExpireTime", minTokenExpireTime, maxTokenExpireTime, true).
		Check(config); !result.Result {
		hwlog.RunLog.Errorf(result.Reason)
		return errors.New(result.Reason)
	}
	return nil
}
