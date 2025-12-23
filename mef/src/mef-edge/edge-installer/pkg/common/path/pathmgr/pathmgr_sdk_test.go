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

// Package pathmgr test for config path manager
package pathmgr

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestConfigPathMgr(t *testing.T) {
	convey.Convey("test PathMgr", t, func() {
		convey.Convey("test ConfigPathMgr method in tags MEFEdge_SDK", func() {
			pathMgr = NewPathMgr(testInstallRootDir, testInstallationPkgDir, testLogRootDir, testLogBackupRootDir)
			configPathMgr := pathMgr.ConfigPathMgr
			configPathMgr.GetTempCrlPath()
			configPathMgr.GetHubSvrCertDir()
			configPathMgr.GetHubSvrCertPath()
			configPathMgr.GetHubSvrKeyPath()
			configPathMgr.GetHubSvrRootCertPath()
			configPathMgr.GetHubSvrRootCertBackupPath()
			configPathMgr.GetHubSvrRootCertPrevBackupPath()
			configPathMgr.GetHubSvrCrlPath()
			configPathMgr.GetHubSvrTempCertPath()
			configPathMgr.GetHubSvrTempKeyPath()
			configPathMgr.GetNetConfigTempDir()
			configPathMgr.GetNetCfgTempRootCertPath()
			configPathMgr.GetNetCfgTempRootCertBackupPath()
		})
	})
}
