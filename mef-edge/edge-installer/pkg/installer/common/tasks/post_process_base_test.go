// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing post precess base
package tasks

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
)

var (
	testPostProcessDir = "/tmp/test_post_process"
	softwareADir       = filepath.Join(testPostProcessDir, "MEFEdge/software_A")
	postProcess        = PostProcessBaseTask{
		WorkPathMgr: pathmgr.NewWorkPathMgr(testPostProcessDir),
		LogPathMgr:  pathmgr.NewLogPathMgr(testPostProcessDir, testPostProcessDir),
	}
)

func setup() error {
	if err := fileutils.DeleteAllFileWithConfusion(testPostProcessDir); err != nil {
		return fmt.Errorf("cleanup [%s] failed, error: %v", testPostProcessDir, err)
	}

	dirNames := []string{constants.SoftwareDirA, constants.SoftwareDirTemp}
	for _, dirName := range dirNames {
		dir := filepath.Join(postProcess.WorkPathMgr.GetMefEdgeDir(), dirName)
		if err := fileutils.CreateDir(dir, constants.Mode755); err != nil {
			return fmt.Errorf("create dir [%s] failed, error: %v", dir, err)
		}
	}

	softwareDirSymlink := postProcess.WorkPathMgr.GetWorkDir()
	if err := os.Symlink(softwareADir, softwareDirSymlink); err != nil {
		return fmt.Errorf("create test software dir symlink failed, error: %v", err)
	}
	return nil
}

func teardown() {
	if err := fileutils.DeleteAllFileWithConfusion(testPostProcessDir); err != nil {
		hwlog.RunLog.Warnf("cleanup [%s] failed, error: %v", testPostProcessDir, err)
	}
}

func TestPostProcessBaseTask(t *testing.T) {
	convey.Convey("test remove upgrade binary by path", t, testRemoveUpgradeBinByPath)
	convey.Convey("test create software symlink", t, testCreateSoftwareSymlink)
	convey.Convey("test update mef service info", t, testUpdateMefServiceInfo)
	convey.Convey("test set software dir immutable", t, testSetSoftwareDirImmutable)
}

func testRemoveUpgradeBinByPath() {
	testUpgradeBin := "/tmp/test_upgrade_bin"
	convey.Convey("remove upgrade binary by path success", func() {
		err := postProcess.RemoveUpgradeBinByPath(testUpgradeBin)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("remove upgrade binary by path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p.Reset()
		err := postProcess.RemoveUpgradeBinByPath(testUpgradeBin)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("remove upgrade binary file [%s] failed,"+
			" error: %v", testUpgradeBin, test.ErrTest))
	})
}

func testCreateSoftwareSymlink() {
	convey.Convey("create software symlink success", func() {
		if err := setup(); err != nil {
			panic(err)
		}
		defer teardown()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("create software symlink failed, get target install dir failed", func() {
		p := gomonkey.ApplyFuncReturn(pathmgr.GetTargetInstallDir, "", test.ErrTest)
		defer p.Reset()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get target software dir failed, error: %v", test.ErrTest))
	})

	p := gomonkey.ApplyFuncReturn(pathmgr.GetTargetInstallDir, softwareADir, nil)
	defer p.Reset()
	convey.Convey("create software symlink failed, remove backup dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldResemble, errors.New("remove backup directory failed"))
	})

	convey.Convey("create software symlink failed, rename upgrade dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.RenameFile, test.ErrTest)
		defer p1.Reset()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("rename upgrade temp directory name to [%s] failed", softwareADir))

	})

	convey.Convey("create software symlink failed, remove old software symlink failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p1.Reset()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("remove old software dir symlink failed, error: %v", test.ErrTest))

	})

	convey.Convey("create software symlink failed, create software symlink failed", func() {
		p1 := gomonkey.ApplyFuncReturn(os.Symlink, test.ErrTest)
		defer p1.Reset()
		err := postProcess.CreateSoftwareSymlink()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("create new software dir symlink failed, error: %v", test.ErrTest))
	})
}

func testUpdateMefServiceInfo() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "UpdateServiceFiles", nil).
		ApplyMethodReturn(common.ComponentMgr{}, "RegisterAllServices", nil)
	defer p.Reset()

	convey.Convey("update mef service info success", func() {
		err := postProcess.UpdateMefServiceInfo()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("update mef service info failed, update service files failed", func() {
		p1 := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "UpdateServiceFiles", test.ErrTest)
		defer p1.Reset()
		err := postProcess.UpdateMefServiceInfo()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("update service files failed, error: %v", test.ErrTest))
	})

	convey.Convey("update mef service info failed, register all services failed", func() {
		p1 := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "RegisterAllServices", test.ErrTest)
		defer p1.Reset()
		err := postProcess.UpdateMefServiceInfo()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("register all services failed, error: %v", test.ErrTest))
	})
}

func testSetSoftwareDirImmutable() {
	convey.Convey("set software immutable success", func() {
		p := gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, softwareADir, nil)
		defer p.Reset()
		err := postProcess.SetSoftwareDirImmutable()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set software immutable failed, get software real path failed", func() {
		p := gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p.Reset()
		err := postProcess.SetSoftwareDirImmutable()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
