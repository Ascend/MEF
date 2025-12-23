// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package cloudcert
package cloudcert

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

// GetEdgeHubCertInfo gets edge-hub cert path
func GetEdgeHubCertInfo(tempPathFlag ...bool) (*certutils.TlsCertInfo, error) {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return nil, errors.New("get config path manager failed")
	}

	kmcCfg, err := util.GetKmcConfig("")
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config dir error: %v", err)
		return nil, fmt.Errorf("get kmc config dir failed")
	}

	certInfo := &certutils.TlsCertInfo{
		RootCaPath: configPathMgr.GetHubSvrRootCertPath(),
		CertPath:   configPathMgr.GetHubSvrCertPath(),
		KeyPath:    configPathMgr.GetHubSvrKeyPath(),
		CrlPath:    configPathMgr.GetHubSvrCrlPath(),
		SvrFlag:    false,
		KmcCfg:     kmcCfg,
		WithBackup: true,
	}
	if len(tempPathFlag) > 0 && tempPathFlag[0] == true {
		certInfo.CertPath = configPathMgr.GetHubSvrTempCertPath()
		certInfo.KeyPath = configPathMgr.GetHubSvrTempKeyPath()
	}
	return certInfo, nil
}
