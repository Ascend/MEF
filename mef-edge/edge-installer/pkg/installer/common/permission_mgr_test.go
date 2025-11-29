// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for testing permission manager
package common

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var permMgr = PermissionMgr{
	ConfigPathMgr:  pathmgr.NewConfigPathMgr(testDir),
	WorkAbsPathMgr: pathmgr.NewWorkAbsPathMgr(filepath.Join(edgeDir, constants.SoftwareDir)),
	LogPathMgr:     pathmgr.NewLogPathMgr(logDir, logDir),
}

func TestPermissionMgr(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil)
	defer p.Reset()

	convey.Convey("test set owner and group should be success", t, setOwnerAndGroupSuccess)
	convey.Convey("test set owner and group should be failed", t, setOwnerAndGroupFailed)
	convey.Convey("test set mode should be success", t, setModeSuccess)
	convey.Convey("test set mode should be failed", t, setModeFailed)
}

func setOwnerAndGroupSuccess() {
	for _, comp := range compNames {
		permMgr.CompName = comp
		err := permMgr.SetOwnerAndGroup()
		convey.So(err, convey.ShouldBeNil)
	}
}

func setOwnerAndGroupFailed() {
	permMgr.CompName = constants.EdgeMain
	convey.Convey("get user configuration map failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.GetMefId, uint32(0), uint32(0), test.ErrTest)
		defer p1.Reset()
		err := permMgr.SetOwnerAndGroup()
		convey.So(err, convey.ShouldResemble, errors.New("get user configuration map failed"))
	})

	convey.Convey("set path owner and group failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
		defer p1.Reset()
		err := permMgr.SetOwnerAndGroup()
		convey.So(err, convey.ShouldResemble, errors.New("set path uid/gid failed"))
	})
}

func setModeSuccess() {
	for _, comp := range compNames {
		permMgr.CompName = comp
		err := permMgr.SetMode()
		convey.So(err, convey.ShouldBeNil)
	}
}

func setModeFailed() {
	permMgr.CompName = constants.EdgeInstaller
	convey.Convey("get matched files failed", func() {
		p1 := gomonkey.ApplyFuncReturn(filepath.Glob, []string{}, test.ErrTest)
		defer p1.Reset()
		err := permMgr.SetMode()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("set path permission failed", func() {
		p1 := gomonkey.ApplyFuncSeq(fileutils.SetPathPermission, []gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},

			{Values: gomonkey.Params{nil}, Times: 4},
			{Values: gomonkey.Params{test.ErrTest}},
		})
		defer p1.Reset()

		err := permMgr.SetMode()
		convey.So(err, convey.ShouldResemble, errors.New("set default mode failed"))

		err = permMgr.SetMode()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
