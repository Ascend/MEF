// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package components for testing prepare component base
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
)

func TestPrepareCompBase(t *testing.T) {
	convey.Convey("prepare component base should be failed, config dir failed", t, prepareConfigDirFailed)
	convey.Convey("prepare component base should be failed, software dir failed", t, prepareSoftwareDirFailed)
	convey.Convey("prepare component base should be failed, log dir failed", t, prepareLogDirsFailed)
	convey.Convey("prepare component base should be failed, log link failed", t, prepareLogLinksFailed)
	convey.Convey("prepare component base should be failed, config link failed", t, prepareConfigLinkFailed)
	convey.Convey("prepare component base should be failed, set owner and mode failed", t, setOwnerAndModeFailed)
}

func prepareConfigDirFailed() {
	convey.Convey("copy config dir failed", func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.prepareConfigDir("")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("copy %s config dir failed, error: %v",
			prepareCompBase.CompName, test.ErrTest))
	})

	convey.Convey("create dir failed", func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.CopyDir, nil).
			ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.prepareConfigDir(testDir, []string{""}...)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("create dir [%s] failed, error: %v", testDir, test.ErrTest))
	})
}

func prepareSoftwareDirFailed() {
	var p = gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
	defer p.Reset()
	err := prepareCompBase.prepareSoftwareDir()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("copy %s software dir failed, error: %v",
		prepareCompBase.CompName, test.ErrTest))
}

func prepareLogDirsFailed() {
	var p = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
	defer p.Reset()
	compLogDir := prepareCompBase.LogPathMgr.GetComponentLogDir(prepareCompBase.CompName)
	err := prepareCompBase.prepareLogDirs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create log dir %s failed, error: %v", compLogDir, test.ErrTest))
}

func prepareLogLinksFailed() {
	convey.Convey("create var dir failed", func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.prepareLogLinks()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("create %s var dir failed, error: %v",
			prepareCompBase.CompName, test.ErrTest))
	})

	convey.Convey("create log dir symlink failed", func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(os.Symlink, test.ErrTest)
		defer p.Reset()
		symlinkDst := filepath.Join(prepareCompBase.WorkAbsPathMgr.GetCompVarDir(prepareCompBase.CompName), constants.Log)
		err := prepareCompBase.prepareLogLinks()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("create log dir symlink %s failed, error: %v",
			symlinkDst, test.ErrTest))
	})
}

func prepareConfigLinkFailed() {
	var p = gomonkey.ApplyFuncReturn(os.Symlink, test.ErrTest)
	defer p.Reset()
	err := prepareCompBase.prepareConfigLink()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create %s config dir symlink failed, error: %v",
		prepareCompBase.CompName, test.ErrTest))
}

func setOwnerAndModeFailed() {
	convey.Convey("get software absolute dir failed", func() {
		var p = gomonkey.ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetWorkAbsDir", "", test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.setOwnerAndMode()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("set owner and group failed", func() {
		var p = gomonkey.ApplyMethodReturn(&common.PermissionMgr{}, "SetOwnerAndGroup", test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.setOwnerAndMode()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set %s owner failed", prepareCompBase.CompName))
	})

	convey.Convey("set mode failed", func() {
		var p = gomonkey.ApplyMethodReturn(&common.PermissionMgr{}, "SetOwnerAndGroup", nil).
			ApplyMethodReturn(&common.PermissionMgr{}, "SetMode", test.ErrTest)
		defer p.Reset()
		err := prepareCompBase.setOwnerAndMode()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set %s mode failed", prepareCompBase.CompName))
	})
}
