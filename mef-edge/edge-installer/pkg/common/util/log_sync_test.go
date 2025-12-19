// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
)

func TestLogSyncMgr(t *testing.T) {
	logDir := CreateLogflie(t)
	defer func() {
		if err := os.RemoveAll(logDir); err != nil {
			t.Fatal(err)
		}
	}()
	patchVarDir := filepath.Join(logDir, constants.SoftwareDir, constants.EdgeInstaller, constants.Var)
	patchLogLinkDir := filepath.Join(patchVarDir, constants.Log)
	patchLogBackupLinkDir := filepath.Join(patchVarDir, constants.LogBackup)

	patches := gomonkey.ApplyFunc(path.GetInstallRootDir, func() (string, error) { return logDir, nil }).
		ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetCompLogLinkDir", patchLogLinkDir).
		ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetCompLogBackupLinkDir", patchLogBackupLinkDir).
		ApplyFunc(envutils.IsInTmpfs, func(_ string) (bool, error) { return true, nil }).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Geteuid()), nil).
		ApplyFunc(fileutils.CheckOwnerAndPermission, mockCheckOwnerAndPermission).
		ApplyFunc(envutils.RunCommandWithUser, mockRunCommandWithUserForNil)
	defer patches.Reset()

	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method BackupLogs", t, func() {
		err := lsm.BackupLogs()
		convey.So(err, convey.ShouldResemble, nil)

		var p1 = gomonkey.ApplyPrivateMethod(&LogSyncMgr{}, "initLogConfigs",
			func() error { return test.ErrTest })
		defer p1.Reset()
		err = lsm.BackupLogs()
		expErr := fmt.Errorf("backup log failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test LogSyncMgr method RecoverLogs", t, func() {
		err := lsm.RecoverLogs()
		convey.So(err, convey.ShouldResemble, nil)

		var p1 = gomonkey.ApplyPrivateMethod(&LogSyncMgr{}, "initLogConfigs",
			func() error { return test.ErrTest })
		defer p1.Reset()
		err = lsm.RecoverLogs()
		expErr := fmt.Errorf("restore log failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func CreateLogflie(t *testing.T) string {
	logdir, err := os.MkdirTemp("/var", "logDir")
	if err != nil {
		t.Fatal(err)
	}
	varDir := filepath.Join(logdir, constants.SoftwareDir, constants.EdgeInstaller, constants.Var)
	installerLogDir := filepath.Join(varDir, constants.Log)
	realInstallerLogDir := filepath.Join(varDir, "real_log")
	if err := os.MkdirAll(realInstallerLogDir, constants.Mode700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realInstallerLogDir, installerLogDir); err != nil {
		t.Fatal(err)
	}
	realInstallerLogBackupDir := filepath.Join(varDir, "real_bakup_log")
	installerLogBackupDir := filepath.Join(varDir, constants.LogBackup)
	if err := os.MkdirAll(realInstallerLogBackupDir, constants.Mode700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realInstallerLogBackupDir, installerLogBackupDir); err != nil {
		t.Fatal(err)
	}

	for _, loginfo := range logFileList {
		if err := os.MkdirAll(filepath.Join(varDir, loginfo.path()), constants.Mode700); err != nil {
			t.Fatal(err)
		}
	}

	return logdir
}

func mockCheckOwnerAndPermission(path string, _ os.FileMode, _ uint32) (string, error) {
	return path, nil
}

func mockRunCommandWithUserForNil(_ string, _ int, _ uint32, _ uint32, _ ...string) (string, error) {
	return "", nil
}

func TestCopyLogFiles(t *testing.T) {
	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method copyLogFiles failed, judge in tmpfs failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, test.ErrTest)
		defer p1.Reset()
		err := lsm.copyLogFiles()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method copyLogFiles failed, is not in tmpfs", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, nil)
		defer p1.Reset()
		err := lsm.copyLogFiles()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test LogSyncMgr method copyLogFiles failed, copy one file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, true, nil).
			ApplyPrivateMethod(&LogSyncMgr{}, "copyOneFile",
				func(logFileRelPath, user string) error { return test.ErrTest })
		defer p1.Reset()
		err := lsm.copyLogFiles()
		expErr := fmt.Errorf("recover log failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestCopyOneFile(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(checkAndPrepareLogFile, "", nil).
		ApplyFuncReturn(fileutils.IsLexist, true).
		ApplyFuncReturn(fileutils.CopyFile, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer patches.Reset()
	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method copyOneFile failed, checkAndPrepareLogFile failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(checkAndPrepareLogFile, outputs)
		defer p1.Reset()

		err := lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
		err = lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method copyOneFile success, syncFilePath exist", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.IsLexist, false)
		defer p1.Reset()
		err := lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test LogSyncMgr method copyOneFile success, copy file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p1.Reset()
		err := lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method copyOneFile success, get uid and gid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(getUidAndGid, uint32(0), uint32(0), test.ErrTest)
		defer p1.Reset()
		err := lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method copyOneFile success, set path owner group failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
		defer p1.Reset()
		err := lsm.copyOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestSyncLogFiles(t *testing.T) {
	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method syncLogFiles failed, judge in tmpfs failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, test.ErrTest)
		defer p1.Reset()
		err := lsm.syncLogFiles()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method syncLogFiles failed, is not in tmpfs", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, nil)
		defer p1.Reset()
		err := lsm.syncLogFiles()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test LogSyncMgr method syncLogFiles failed, sync one file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, true, nil).
			ApplyPrivateMethod(&LogSyncMgr{}, "syncOneFile",
				func(logFileRelPath, user string) error { return test.ErrTest })
		defer p1.Reset()
		err := lsm.syncLogFiles()
		expErr := fmt.Errorf("recover log failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestSyncOneFile(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(checkAndPrepareLogFile, "", nil).
		ApplyFuncReturn(fileutils.IsLexist, true).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil).
		ApplyFuncReturn(envutils.RunCommandWithUser, "", nil)
	defer patches.Reset()
	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method syncOneFile success, logFilePath exist", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.IsLexist, false)
		defer p1.Reset()
		err := lsm.syncOneFile("", "")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test LogSyncMgr method syncOneFile failed, checkAndPrepareLogFile failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(checkAndPrepareLogFile, outputs)
		defer p1.Reset()

		err := lsm.syncOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
		err = lsm.syncOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method syncOneFile success, get uid and gid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(getUidAndGid, uint32(0), uint32(0), test.ErrTest)
		defer p1.Reset()
		err := lsm.syncOneFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test LogSyncMgr method syncOneFile success, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommandWithUser, "", test.ErrTest)
		defer p1.Reset()
		err := lsm.syncOneFile("", "")
		expErr := fmt.Errorf("execute rsync failed, error: %v, output: %s", test.ErrTest, "")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestInitLogConfigs(t *testing.T) {
	lsm := NewLogSyncMgr()
	convey.Convey("test LogSyncMgr method initLogConfigs success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetCompLogDirs, "", "", nil)
		defer p1.Reset()
		err := lsm.initLogConfigs()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test LogSyncMgr method initLogConfigs failed, get installer log dirs failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetCompLogDirs, "", "", test.ErrTest)
		defer p1.Reset()
		err := lsm.initLogConfigs()
		expErr := errors.New("get component log dirs failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetUidAndGid(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer patches.Reset()
	convey.Convey("test func getUidAndGid success", t, func() {
		_, _, err := getUidAndGid("")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func getUidAndGid failed, get uid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		_, _, err := getUidAndGid("")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func getUidAndGid failed, get gid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetGid, uint32(0), test.ErrTest)
		defer p1.Reset()
		_, _, err := getUidAndGid("")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestCheckAndPrepareLogFile(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil).
		ApplyFuncReturn(fileutils.CheckOriginPath, "/tmp/log", nil).
		ApplyFuncReturn(fileutils.IsLexist, false).
		ApplyFuncReturn(fileutils.CreateDir, nil).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil).
		ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", nil)
	defer patches.Reset()

	convey.Convey("test func checkAndPrepareLogFile failed, get uid or gid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(getUidAndGid, uint32(0), uint32(0), test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkAndPrepareLogFile failed, check log path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "/tmp/log", test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkAndPrepareLogFile failed, create log root path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkAndPrepareLogFile failed, check log root path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkAndPrepareLogFile failed, set log dir owner group failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkAndPrepareLogFile failed, check log dir owner and permission failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p1.Reset()
		_, err := checkAndPrepareLogFile("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
