// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package install

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func WorkingDirMgrTest() {
	convey.Convey("WorkingDirMgr DoInstallPrepare func", WorkingDirMgrDoPrepareTest)
	convey.Convey("prepareRootWorkDir func", PrepareRootWorkDirTest)
	convey.Convey("prepareLibDir func", PrepareLibDirTest)
	convey.Convey("prepareRunSh func", PrepareRunShTest)
	convey.Convey("prepareBinDir func", PrepareBinDirTest)
	convey.Convey("prepareVersionXml func", PrepareVersionXmlTest)
	convey.Convey("prepareComponentWorkDir func", PrepareComponentWorkDirTest)
	convey.Convey("prepareSymlinks func", PrepareSymlinksTest)
}

func WorkingDirMgrDoPrepareTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	convey.Convey("test workingDirCtl struct DoInstallPrepare func success", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "prepareRootWorkDir", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareLibDir", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareRunSh", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareBinDir", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareVersionXml", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareComponentWorkDir", func(_ *WorkingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareSymlinks", func(_ *WorkingDirCtl) error { return nil })

		defer p.Reset()
		convey.So(ins.DoInstallPrepare(), convey.ShouldBeNil)
	})

	convey.Convey("test workingDirCtl struct DoInstallPrepare func failed", func() {
		p := gomonkey.ApplyPrivateMethod(ins, "prepareRootWorkDir",
			func(_ *WorkingDirCtl) error { return ErrTest })
		defer p.Reset()
		convey.So(ins.DoInstallPrepare(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareRootWorkDirTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	convey.Convey("test prepareRootWorkDir func success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil)
		defer p.Reset()
		convey.So(ins.prepareRootWorkDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareRootWorkDir func failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareRootWorkDir(), convey.ShouldResemble, errors.New("create mef root work path failed"))
	})
}

func PrepareLibDirTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}
	var componentMgrIns *util.ComponentMgr

	convey.Convey("test prepareLibDir func success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(fileutils.CopyDir, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareLibDir", nil).
			ApplyFuncReturn(fileutils.CopyDirWithSoftlink, nil)
		defer p.Reset()
		convey.So(ins.prepareLibDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareLibDir func get current path failed", func() {
		p := gomonkey.ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareLibDir(), convey.ShouldResemble, errors.New("get current path failed"))
	})

	convey.Convey("test prepareLibDir func makesure path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareLibDir(), convey.ShouldResemble, errors.New("create lib path failed"))
	})

	convey.Convey("test prepareLibDir func copy dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(fileutils.CopyDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareLibDir(), convey.ShouldResemble, errors.New("copy lib dir failed"))
	})

	convey.Convey("test prepareLibDir func prepare component lib dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(fileutils.CopyDir, nil).
			ApplyFuncReturn(fileutils.CopyDirWithSoftlink, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareLibDir", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareLibDir(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareRunShTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	convey.Convey("test prepareRunSh func success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, nil).ApplyFuncReturn(fileutils.SetPathPermission, nil)
		defer p.Reset()
		convey.So(ins.prepareRunSh(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareRunSh func get current path failed", func() {
		p := gomonkey.ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareRunSh(), convey.ShouldResemble, errors.New("get current path failed"))
	})

	convey.Convey("test prepareRunSh func copy file failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareRunSh(), convey.ShouldResemble, errors.New("copy run scripts dir failed"))
	})

	convey.Convey("test prepareRunSh func change mod failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, nil).ApplyFuncReturn(os.Chmod, ErrTest).
			ApplyFuncReturn(fileutils.SetPathPermission, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareRunSh(), convey.ShouldResemble, errors.New("set run script path mode failed"))
	})
}

func PrepareBinDirTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	convey.Convey("test prepareBinDir func success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).ApplyFuncReturn(fileutils.CopyDir, nil).
			ApplyFuncReturn(fileutils.CopyFile, nil)
		defer p.Reset()
		convey.So(ins.prepareBinDir(), convey.ShouldBeNil)
	})

	convey.Convey("test prepareBinDir func get current path failed", func() {
		p := gomonkey.ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareBinDir(), convey.ShouldResemble, errors.New("get current path failed"))
	})

	convey.Convey("test prepareBinDir func makesure path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareBinDir(), convey.ShouldResemble, errors.New("create sbin work path failed"))
	})

	convey.Convey("test prepareBinDir func copyfile failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).ApplyFuncReturn(fileutils.CopyDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareBinDir(), convey.ShouldResemble, errors.New("copy mef controller failed"))
	})
}

func PrepareVersionXmlTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	convey.Convey("test func prepareVersionXm func success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, nil).ApplyFuncReturn(os.Chmod, nil).
			ApplyFuncReturn(fileutils.SetPathPermission, nil)
		defer p.Reset()
		convey.So(ins.prepareVersionXml(), convey.ShouldBeNil)
	})

	convey.Convey("test func prepareVersionXm get current path failed", func() {
		p := gomonkey.ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareVersionXml(), convey.ShouldResemble, errors.New("get current path failed"))
	})

	convey.Convey("test func prepareVersionXm copy file failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareVersionXml(), convey.ShouldResemble, errors.New("copy version.xml failed"))
	})

	convey.Convey("test func prepareVersionXm func change mod failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CopyFile, nil).ApplyFuncReturn(fileutils.SetPathPermission, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareVersionXml(), convey.ShouldResemble, errors.New("set version.xml path mode failed"))
	})
}

func PrepareComponentWorkDirTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}
	var componentMgrIns *util.ComponentMgr

	convey.Convey("test func prepareComponentWorkDir success", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareSingleComponentDir", nil)
		defer p.Reset()
		convey.So(ins.prepareComponentWorkDir(), convey.ShouldBeNil)
	})

	convey.Convey("test func prepareComponentWorkDir makesure dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareComponentWorkDir(), convey.ShouldResemble, errors.New("create component root work path failed"))
	})

	convey.Convey("test func prepareComponentWorkDir prepare components dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareSingleComponentDir", ErrTest)
		defer p.Reset()
		convey.So(ins.prepareComponentWorkDir(), convey.ShouldResemble, ErrTest)
	})
}

func PrepareSymlinksTest() {
	var ins = &WorkingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}

	convey.Convey("test func prepareSymlinks success", func() {
		p := gomonkey.ApplyFuncReturn(os.Symlink, nil)
		defer p.Reset()
		convey.So(ins.prepareSymlinks(), convey.ShouldBeNil)
	})

	convey.Convey("test func prepareSymlinks failed", func() {
		p := gomonkey.ApplyFuncReturn(os.Symlink, ErrTest)
		defer p.Reset()
		convey.So(ins.prepareSymlinks(), convey.ShouldResemble, errors.New("create work dir symlink failed"))
	})
}
