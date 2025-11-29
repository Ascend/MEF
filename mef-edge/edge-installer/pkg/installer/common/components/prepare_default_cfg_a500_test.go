// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package components for testing prepare default config backup
package components

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

func TestPrepareCfgBackupDir(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
	defer p.Reset()

	convey.Convey("prepare default config backup dir should be failed, create dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		err := prepareCompBase.prepareDefaultCfgBackupDirBase()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("create dir [%s] failed, error: %v",
			prepareCompBase.SoftwarePathMgr.ConfigPathMgr.GetConfigBackupTempDir(), test.ErrTest))
	})

	convey.Convey("prepare default config backup dir should be failed, prepare config dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
		defer p1.Reset()
		err := prepareCompBase.prepareDefaultCfgBackupDirBase()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("prepare %s config backup dir failed",
			prepareCompBase.CompName))
	})
}
