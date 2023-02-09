// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

import (
	"fmt"
	"os"
	"strconv"

	"huawei.com/mindx/common/hwlog"
)

// Envs nginx使用到的环境变量
var Envs map[string]string

// InitEnvs 初始化环境变量
func InitEnvs() {
	Envs = make(map[string]string)
	Envs[EdgePortKey] = os.Getenv(EdgePortKey)
	Envs[SoftPortKey] = os.Getenv(SoftPortKey)
	Envs[CertPortKey] = os.Getenv(CertPortKey)
	Envs[UserMgrSvcPortKey] = os.Getenv(UserMgrSvcPortKey)
	Envs[NginxSslPortKey] = os.Getenv(NginxSslPortKey)
}

// GetEnvAsInt 获取int类型的环境变量
func GetEnvAsInt(key string) (int, error) {
	if _, ok := Envs[key]; !ok {
		hwlog.RunLog.Errorf("cannot find env: %s", key)
		return 0, fmt.Errorf("cannot find env: %s", key)
	}
	ret, err := strconv.Atoi(Envs[key])
	if err != nil {
		hwlog.RunLog.Errorf("parse env %s error", key)
		return 0, fmt.Errorf("parse env %s error", key)
	}
	return ret, nil
}
