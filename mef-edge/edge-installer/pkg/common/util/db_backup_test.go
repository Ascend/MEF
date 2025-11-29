// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util test for db_backup.go
package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
)

func TestStartBackupEdgeOmDb(t *testing.T) {
	if err := test.InitDb("/tmp/test.db"); err != nil {
		panic("init db failed")
	}
	var patches = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("./"), nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(fileutils.CheckOriginPath, "", nil).
		ApplyFuncReturn(fileutils.CheckMode, true).
		ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
		ApplyMethodReturn(&os.File{}, "Close", nil).
		ApplyMethodReturn(&os.File{}, "Stat", &mockFileInfo{}, nil).
		ApplyMethodReturn(&os.File{}, "Chown", nil).
		ApplyFuncReturn(gorm.Open, test.MockGetDb(), nil).
		ApplyFuncReturn(envutils.RunCommandWithOptions, envutils.CommandResult{Err: nil}).
		ApplyFuncReturn(fileutils.CopyFile, nil)
	defer patches.Reset()

	convey.Convey("test func StartBackupEdgeOmDb success", t, testBackupOmDb)
	convey.Convey("test func StartBackupEdgeOmDb failed, GetConfigPathMgr failed", t, testBackupOmDbErrGetPathMgr)
	convey.Convey("test func StartBackupEdgeOmDb failed, db file doesn't exist failed", t, testBackupOmDbErrExist)
	convey.Convey("test func StartBackupEdgeOmDb failed, check db file path failed", t, testBackupOmDbErrCheckPath)
	convey.Convey("test func StartBackupEdgeOmDb failed, check db file mode failed", t, testBackupOmDbErrCheckMode)
	convey.Convey("test func StartBackupEdgeOmDb failed, close db file failed", t, testBackupOmDbErrClose)
	convey.Convey("test func StartBackupEdgeOmDb failed, get db file stat failed", t, testBackupOmDbErrStat)
	convey.Convey("test func StartBackupEdgeOmDb failed, chown db file failed", t, testBackupOmDbErrChown)
	convey.Convey("test func StartBackupEdgeOmDb failed, open db failed", t, testBackupOmDbErrOpenDb)
	convey.Convey("test func StartBackupEdgeOmDb failed, copy file failed", t, testBackupOmDbErrCopyFile)
}

func testBackupOmDb() {
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldBeNil)
}

func testBackupOmDbErrGetPathMgr() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	expErr := fmt.Errorf("get config path manager failed, %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testBackupOmDbErrExist() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrCheckPath() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrCheckMode() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckMode, false)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrClose() {
	var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Close", test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrStat() {
	var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Stat", &mockFileInfo{}, test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrChown() {
	var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Chown", test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrOpenDb() {
	var p1 = gomonkey.ApplyFuncReturn(gorm.Open, test.MockGetDb(), test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func testBackupOmDbErrCopyFile() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
	defer p1.Reset()
	_, err := StartBackupEdgeOmDb(context.Background())
	convey.So(err, convey.ShouldResemble, errors.New("backup db is broken"))
}

func TestStartBackupEdgeCoreDb(t *testing.T) {
	if err := test.InitDb("/tmp/test.db"); err != nil {
		panic("init db failed")
	}
	var patches = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Geteuid()), nil).
		ApplyPrivateMethod(singleDbBackupTask{}, "checkIntegrity", func(string) error { return nil })
	defer patches.Reset()

	convey.Convey("test func StartBackupEdgeCoreDb success", t, func() {
		_, err := StartBackupEdgeCoreDb(context.Background())
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func StartBackupEdgeCoreDb failed, get mef uid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), test.ErrTest)
		defer p1.Reset()
		_, err := StartBackupEdgeCoreDb(context.Background())
		expErr := fmt.Errorf("failed to get uid of %s, %v", constants.EdgeUserName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestInitEdgeOmDbBackupTask(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(GetMefId, uint32(os.Geteuid()), uint32(os.Getegid()), nil).
		ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("./"), nil)
	defer patches.Reset()
	convey.Convey("test func initEdgeOmDbBackupTask success", t, func() {
		_, err := initEdgeCoreBackupTask()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func initEdgeOmDbBackupTask failed, get mef id failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(GetMefId, uint32(os.Geteuid()), uint32(os.Getegid()), test.ErrTest)
		defer p1.Reset()
		_, err := initEdgeCoreBackupTask()
		expErr := fmt.Errorf("failed to get uid of %s, %v", constants.EdgeUserName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func initEdgeOmDbBackupTask failed, get config path manager failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
		defer p1.Reset()
		_, err := initEdgeCoreBackupTask()
		expErr := fmt.Errorf("get config path manager failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestCheckEdgeDbIntegrity(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(GetMefId, uint32(os.Geteuid()), uint32(os.Getegid()), nil).
		ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("./"), nil).
		ApplyPrivateMethod(singleDbBackupTask{}, "checkIntegrity",
			func(dbPath string) error { return nil })
	defer patches.Reset()

	convey.Convey("test func CheckEdgeDbIntegrity success", t, func() {
		err := CheckEdgeDbIntegrity()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func CheckEdgeDbIntegrity failed, get mef id failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(GetMefId, uint32(os.Geteuid()), uint32(os.Getegid()), test.ErrTest)
		defer p1.Reset()
		err := CheckEdgeDbIntegrity()
		expErr := fmt.Errorf("failed to get uid of %s, %v", constants.EdgeUserName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func CheckEdgeDbIntegrity failed, get config path manager failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
		defer p1.Reset()
		err := CheckEdgeDbIntegrity()
		expErr := fmt.Errorf("get config path manager failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func CheckEdgeDbIntegrity failed, check integrity failed", t, func() {
		var p1 = gomonkey.ApplyPrivateMethod(singleDbBackupTask{}, "checkIntegrity",
			func(dbPath string) error { return test.ErrTest })
		defer p1.Reset()
		err := CheckEdgeDbIntegrity()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
