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
