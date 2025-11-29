// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package components for testing prepare edge installer
package components

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

var prepareInstallerTask = NewPrepareInstaller(pathMgr, workAbsPathMgr)

func TestPrepareInstallerPrepareCfgDir(t *testing.T) {
	convey.Convey("prepare installer config should be success", t, func() {
		err := prepareInstallerTask.PrepareCfgDir()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPrepareInstallerRun(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil)
	defer p.Reset()

	convey.Convey("prepare installer should be success", t, prepareInstallerSuccess)
	convey.Convey("prepare installer should be failed, prepare version file failed", t, prepareVersionFileFailed)
	convey.Convey("prepare installer should be failed, prepare run sh failed", t, prepareRunShFailed)
	convey.Convey("prepare installer should be failed, prepare lib failed", t, prepareLibFailed)
	convey.Convey("prepare installer should be failed, remove unnecessary files failed", t, removeUnnecessaryFilesFailed)
}

func prepareInstallerSuccess() {
	err := prepareInstallerTask.Run()
	convey.So(err, convey.ShouldBeNil)
}

func prepareVersionFileFailed() {
	convey.Convey("copy version.xml failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p1.Reset()
		err := prepareInstallerTask.Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("copy %s failed, error: %v", constants.VersionXml, test.ErrTest))
	})

	convey.Convey("set file mode failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, test.ErrTest)
		defer p1.Reset()
		err := prepareInstallerTask.Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set file [%s] mode failed: %v",
			prepareInstallerTask.WorkAbsPathMgr.GetVersionXmlPath(), test.ErrTest))
	})
}

func prepareRunShFailed() {
	convey.Convey("copy run.sh failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p1.Reset()
		err := prepareInstallerTask.prepareRunSh()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("copy file [%s] failed,"+
			" error: %v", constants.RunScript, test.ErrTest))
	})

	convey.Convey("set file mode failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, test.ErrTest)
		defer p1.Reset()
		err := prepareInstallerTask.prepareRunSh()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set file [%s] mode failed: %v", constants.RunScript, test.ErrTest))
	})
}

func prepareLibFailed() {
	convey.Convey("copy lib dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CopyDirWithSoftlink, test.ErrTest)
		defer p1.Reset()
		err := prepareInstallerTask.prepareLib()
		convey.So(err, convey.ShouldResemble, errors.New("copy lib dir failed"))
	})

	convey.Convey("set mode failed", func() {
		libDst := prepareInstallerTask.WorkAbsPathMgr.GetLibDir()
		p1 := gomonkey.ApplyFuncSeq(fileutils.SetPathPermission, []gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},

			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		})
		defer p1.Reset()

		err := prepareInstallerTask.prepareLib()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set lib files mode in dir [%s] failed: %v", libDst, test.ErrTest))

		err = prepareInstallerTask.prepareLib()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set dir [%s] mode failed: %v", libDst, test.ErrTest))
	})
}

func removeUnnecessaryFilesFailed() {
	file := prepareInstallerTask.WorkAbsPathMgr.GetInstallBinaryPath()
	p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
	defer p1.Reset()
	err := prepareInstallerTask.removeUnnecessaryFiles()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("remove file [%s] failed, error: %v", file, test.ErrTest))
}
