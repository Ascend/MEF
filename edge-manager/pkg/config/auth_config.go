// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
