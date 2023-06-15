// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package config for
package config

import (
	"errors"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
)

var certConfig CertConfigInfo

const (
	minCertExpireTime = 7
	maxCertExpireTime = 100
)

// CertConfigInfo [struct] for save cert config
type CertConfigInfo struct {
	CertExpireTime int
}

// SetConfig to set cert config
func SetConfig(config CertConfigInfo) {
	certConfig = config
}

// GetCertConfig to get cert config
func GetCertConfig() CertConfigInfo {
	return certConfig
}

// CheckCertConfig check cert config
func CheckCertConfig(config CertConfigInfo) error {
	if result := checker.GetIntChecker("CertExpireTime", minCertExpireTime, maxCertExpireTime, true).
		Check(config); !result.Result {
		hwlog.RunLog.Errorf(result.Reason)
		return errors.New(result.Reason)
	}
	return nil
}
