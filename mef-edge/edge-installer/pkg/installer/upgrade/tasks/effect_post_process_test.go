// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing post effect process
package tasks

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/tasks"
)

func TestPostEffectProcessTask(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(&util.EdgeGUidMgr{}, "SetEUGidToEdge", nil).
		ApplyFuncReturn(deleteMetaByType, nil).
		ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "CreateSoftwareSymlink", nil).
		ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "UpdateMefServiceInfo", nil).
		ApplyFuncReturn(common.CopyResetScriptToP7, nil).
		ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "SetSoftwareDirImmutable", nil).
		ApplyFuncReturn(config.SmoothEdgeCoreConfigPipePath, nil).
		ApplyFuncReturn(config.SmoothEdgeCoreSafeConfig, nil).
		ApplyFuncReturn(config.SmoothEdgeOmContainerConfig, nil).
		ApplyFuncReturn(config.SmoothAlarmConfigDB, nil).
		ApplyFuncReturn(fileutils.RenameFile, nil).
		ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil).
		ApplyMethodReturn(backuputils.NewBackupDirMgr(""), "BackUp", test.ErrTest).
		ApplyFuncReturn(util.IsServiceActive, true).
		ApplyFuncReturn(util.RestartService, nil).
		ApplyMethod(common.ComponentMgr{}, "CheckAllServiceActive", func(common.ComponentMgr) {})
	defer p.Reset()

	convey.Convey("post effect process should be success", t, postEffectProcessTaskSuccess)
	convey.Convey("post effect process should be failed, clear alarm in db failed", t, clearAlarmInDBFailed)
	convey.Convey("post effect process should be failed, copy reset script failed", t, copyResetScriptFailed)
	convey.Convey("post effect process should be failed, smooth config failed", t, smoothConfigFailed)
	convey.Convey("post effect process should be failed, backup config failed", t, backUpConfigFailed)
	convey.Convey("post effect process should be failed, restart failed", t, restartFailed)
}

func postEffectProcessTaskSuccess() {
	convey.Convey("old service is not running, no need to restart", func() {
		p1 := gomonkey.ApplyFuncReturn(util.IsServiceActive, false)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("restart all services success", func() {
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldBeNil)
	})
}

func clearAlarmInDBFailed() {
	convey.Convey("set euid and egid failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&util.EdgeGUidMgr{}, "SetEUGidToEdge", test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set euid and egid to %s failed, error: %v",
			constants.EdgeUserName, test.ErrTest))
	})

	convey.Convey("reset euid and egid failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&util.EdgeGUidMgr{}, "ResetEUGid", test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldBeNil)
	})
}

func copyResetScriptFailed() {
	p1 := gomonkey.ApplyFuncReturn(common.CopyResetScriptToP7, test.ErrTest)
	defer p1.Reset()
	err := postEffectProcess.Run()
	convey.So(err, convey.ShouldBeNil)
}

func smoothConfigFailed() {
	convey.Convey("smooth edge core config pipe path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(config.SmoothEdgeCoreConfigPipePath, test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldResemble, errors.New("smooth pipe config to edge core config file failed"))
	})

	convey.Convey("smooth safe config to edge core config file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(config.SmoothEdgeCoreSafeConfig, test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldResemble, errors.New("smooth safe config to edge core config file failed"))
	})

	convey.Convey("smooth container config to edge om config file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(config.SmoothEdgeOmContainerConfig, test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldResemble, errors.New("smooth edge_om container config to edge om config file failed"))
	})
}

func backUpConfigFailed() {
	convey.Convey("create backup files failed", func() {
		p1 := gomonkey.ApplyMethodReturn(backuputils.NewBackupDirMgr(""), "BackUp", test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set owner for backup files failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, test.ErrTest)
		defer p1.Reset()
		err := postEffectProcess.Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set [%s] confg dir owner for backup files failed,"+
			" error: %v", constants.EdgeMain, test.ErrTest))
	})
}

func restartFailed() {
	p1 := gomonkey.ApplyFuncReturn(util.RestartService, test.ErrTest)
	defer p1.Reset()
	err := postEffectProcess.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("restart target [%s] failed", constants.MefEdgeTargetFile))
}

func TestDeleteMetaByType(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(&config.DbMgr{}, "InitDB", nil)
	defer p.Reset()

	convey.Convey("delete meta by type success", t, deleteMetaByTypeSuccess)
	convey.Convey("delete meta by type failed", t, deleteMetaByTypeFailed)
}

func deleteMetaByTypeSuccess() {
	convey.Convey("delete meta by type success, table does not exist", func() {
		err := deleteMetaByType("", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("delete meta by type success, record not found", func() {
		err := test.MockGetDb().AutoMigrate(&Meta{})
		convey.So(err, convey.ShouldBeNil)
		err = deleteMetaByType("", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("delete meta by type success, delete alarm from database success", func() {
		tx := test.MockGetDb().Create(Meta{Type: constants.MetaAlarmKey, Value: ""})
		convey.So(tx.Error, convey.ShouldBeNil)
		err := deleteMetaByType("", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldBeNil)
	})
}

func deleteMetaByTypeFailed() {
	convey.Convey("delete meta by type failed, init database failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&config.DbMgr{}, "InitDB", test.ErrTest)
		defer p1.Reset()
		err := deleteMetaByType("", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("init database [%s] failed", constants.DbEdgeMainPath))
	})

	convey.Convey("delete meta by type failed, get database failed", func() {
		p1 := gomonkey.ApplyFuncReturn(database.GetDb, nil)
		defer p1.Reset()
		err := deleteMetaByType("/tmp", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get database [%s] failed", constants.DbEdgeMainPath))
	})

	convey.Convey("delete meta by type failed, find alarm in database failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&gorm.DB{}, "First", &gorm.DB{Error: test.ErrTest})
		defer p1.Reset()
		err := deleteMetaByType("/tmp", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("find [%s] in database failed", constants.MetaAlarmKey))
	})

	convey.Convey("delete meta by type failed, delete alarm from database failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&gorm.DB{}, "First", &gorm.DB{Error: nil}).
			ApplyMethodReturn(&gorm.DB{}, "Delete", &gorm.DB{Error: test.ErrTest})
		defer p1.Reset()
		err := deleteMetaByType("/tmp", constants.DbEdgeMainPath, constants.MetaAlarmKey)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("delete [%s] from database failed", constants.MetaAlarmKey))
	})
}
