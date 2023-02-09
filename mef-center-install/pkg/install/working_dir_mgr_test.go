// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func WorkingDirMgrTest() {
	Convey("WorkingDirMgr doPrepare func", WorkingDirMgrDoPrepareTest)
	Convey("prepareRootWorkDir func", PrepareRootWorkDirTest)
	Convey("prepareLibDir func", PrepareLibDirTest)
	Convey("prepareRunSh func", PrepareRunShTest)
	Convey("prepareBinDir func", PrepareBinDirTest)
	Convey("prepareVersionXml func", PrepareVersionXmlTest)
	Convey("prepareComponentWorkDir func", PrepareComponentWorkDirTest)
	Convey("prepareSymlinks func", PrepareSymlinksTest)
}

func WorkingDirMgrDoPrepareTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	Convey("test workingDirCtl struct doPrepare func success", func() {
		p := ApplyPrivateMethod(ins, "prepareRootWorkDir", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareLibDir", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareRunSh", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareBinDir", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareVersionXml", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareComponentWorkDir", func(_ *workingDirCtl) error { return nil }).
			ApplyPrivateMethod(ins, "prepareSymlinks", func(_ *workingDirCtl) error { return nil })
		defer p.Reset()
		So(ins.doPrepare(), ShouldBeNil)
	})

	Convey("test workingDirCtl struct doPrepare func failed", func() {
		p := ApplyPrivateMethod(ins, "prepareRootWorkDir",
			func(_ *workingDirCtl) error { return ErrTest })
		defer p.Reset()
		So(ins.doPrepare(), ShouldResemble, ErrTest)
	})
}

func PrepareRootWorkDirTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	Convey("test prepareRootWorkDir func success", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil)
		defer p.Reset()
		So(ins.prepareRootWorkDir(), ShouldBeNil)
	})

	Convey("test prepareRootWorkDir func failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, ErrTest)
		defer p.Reset()
		So(ins.prepareRootWorkDir(), ShouldResemble, errors.New("create mef root work path failed"))
	})
}

func PrepareLibDirTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}
	var componentMgrIns *util.ComponentMgr

	Convey("test prepareLibDir func success", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).
			ApplyFuncReturn(common.CopyDir, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareLibDir", nil)
		defer p.Reset()
		So(ins.prepareLibDir(), ShouldBeNil)
	})

	Convey("test prepareLibDir func get current path failed", func() {
		p := ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		So(ins.prepareLibDir(), ShouldResemble, errors.New("get current path failed"))
	})

	Convey("test prepareLibDir func makesure path failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, ErrTest)
		defer p.Reset()
		So(ins.prepareLibDir(), ShouldResemble, errors.New("create lib path failed"))
	})

	Convey("test prepareLibDir func copy dir failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).
			ApplyFuncReturn(common.CopyDir, ErrTest)
		defer p.Reset()
		So(ins.prepareLibDir(), ShouldResemble, errors.New("copy lib dir failed"))
	})

	Convey("test prepareLibDir func prepare component lib dir failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).
			ApplyFuncReturn(common.CopyDir, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareLibDir", ErrTest)
		defer p.Reset()
		So(ins.prepareLibDir(), ShouldResemble, ErrTest)
	})
}

func PrepareRunShTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	Convey("test prepareRunSh func success", func() {
		p := ApplyFuncReturn(utils.CopyFile, nil).ApplyFuncReturn(os.Chmod, nil)
		defer p.Reset()
		So(ins.prepareRunSh(), ShouldBeNil)
	})

	Convey("test prepareRunSh func get current path failed", func() {
		p := ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		So(ins.prepareRunSh(), ShouldResemble, errors.New("get current path failed"))
	})

	Convey("test prepareRunSh func copy file failed", func() {
		p := ApplyFuncReturn(utils.CopyFile, ErrTest)
		defer p.Reset()
		So(ins.prepareRunSh(), ShouldResemble, errors.New("copy run scripts dir failed"))
	})

	Convey("test prepareRunSh func change mod failed", func() {
		p := ApplyFuncReturn(utils.CopyFile, nil).ApplyFuncReturn(os.Chmod, ErrTest)
		defer p.Reset()
		So(ins.prepareRunSh(), ShouldResemble, errors.New("set run script path mode failed"))
	})
}

func PrepareBinDirTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	Convey("test prepareBinDir func success", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).ApplyFuncReturn(utils.CopyDir, nil).
			ApplyFuncReturn(common.DeleteAllFile, nil)
		defer p.Reset()
		So(ins.prepareBinDir(), ShouldBeNil)
	})

	Convey("test prepareBinDir func get current path failed", func() {
		p := ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		So(ins.prepareBinDir(), ShouldResemble, errors.New("get current path failed"))
	})

	Convey("test prepareBinDir func makesure path failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, ErrTest)
		defer p.Reset()
		So(ins.prepareBinDir(), ShouldResemble, errors.New("create sbin work path failed"))
	})

	Convey("test prepareBinDir func copydir failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).ApplyFuncReturn(utils.CopyDir, ErrTest)
		defer p.Reset()
		So(ins.prepareBinDir(), ShouldResemble, errors.New("copy mef scripts dir failed"))
	})

	Convey("test prepareBinDir func delete all file failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).ApplyFuncReturn(utils.CopyDir, nil).
			ApplyFuncReturn(common.DeleteAllFile, ErrTest)
		defer p.Reset()
		So(ins.prepareBinDir(), ShouldResemble, errors.New("delete installer bin failed"))
	})
}

func PrepareVersionXmlTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  nil,
	}

	Convey("test func prepareVersionXm func success", func() {
		p := ApplyFuncReturn(utils.CopyFile, nil).ApplyFuncReturn(os.Chmod, nil)
		defer p.Reset()
		So(ins.prepareVersionXml(), ShouldBeNil)
	})

	Convey("test func prepareVersionXm get current path failed", func() {
		p := ApplyFuncReturn(filepath.Abs, "", ErrTest)
		defer p.Reset()
		So(ins.prepareVersionXml(), ShouldResemble, errors.New("get current path failed"))
	})

	Convey("test func prepareVersionXm copy file failed", func() {
		p := ApplyFuncReturn(utils.CopyFile, ErrTest)
		defer p.Reset()
		So(ins.prepareVersionXml(), ShouldResemble, errors.New("copy version.xml failed"))
	})

	Convey("test func prepareVersionXm func change mod failed", func() {
		p := ApplyFuncReturn(utils.CopyFile, nil).ApplyFuncReturn(os.Chmod, ErrTest)
		defer p.Reset()
		So(ins.prepareVersionXml(), ShouldResemble, errors.New("set version.xml path mode failed"))
	})
}

func PrepareComponentWorkDirTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}
	var componentMgrIns *util.ComponentMgr

	Convey("test func prepareComponentWorkDir success", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareSingleComponentDir", nil)
		defer p.Reset()
		So(ins.prepareComponentWorkDir(), ShouldBeNil)
	})

	Convey("test func prepareComponentWorkDir makesure dir failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, ErrTest)
		defer p.Reset()
		So(ins.prepareComponentWorkDir(), ShouldResemble, errors.New("create component root work path failed"))
	})

	Convey("test func prepareComponentWorkDir prepare components dir failed", func() {
		p := ApplyFuncReturn(common.MakeSurePath, nil).
			ApplyMethodReturn(componentMgrIns, "PrepareSingleComponentDir", ErrTest)
		defer p.Reset()
		So(ins.prepareComponentWorkDir(), ShouldResemble, ErrTest)
	})
}

func PrepareSymlinksTest() {
	var ins = &workingDirCtl{
		pathMgr:     &util.WorkPathAMgr{},
		mefLinkPath: "",
		components:  []string{"edge-manager"},
	}

	Convey("test func prepareSymlinks success", func() {
		p := ApplyFuncReturn(os.Symlink, nil)
		defer p.Reset()
		So(ins.prepareSymlinks(), ShouldBeNil)
	})

	Convey("test func prepareSymlinks failed", func() {
		p := ApplyFuncReturn(os.Symlink, ErrTest)
		defer p.Reset()
		So(ins.prepareSymlinks(), ShouldResemble, errors.New("create work dir symlink failed"))
	})
}
