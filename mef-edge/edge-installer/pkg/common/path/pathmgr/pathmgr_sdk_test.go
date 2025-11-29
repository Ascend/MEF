// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
