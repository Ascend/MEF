// Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

// 因为无root权限,且想不到好的办法 所以全部打桩
func TestIsService(t *testing.T) {
	patch := gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
	defer patch.Reset()
	convey.Convey("IsServiceInSystemd", t, func() {
		if !fileutils.IsExist(constants.SystemdServiceDir) {
			convey.So(IsServiceInSystemd(""), convey.ShouldResemble, false)
		}
		convey.So(IsServiceInSystemd(""), convey.ShouldResemble, true)
	})

	convey.Convey("test func ReloadServiceDaemon", t, func() {
		convey.Convey("test func ReloadServiceDaemon success", func() {
			convey.So(ReloadServiceDaemon(), convey.ShouldResemble, nil)
		})
		convey.Convey("test func ReloadServiceDaemon failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(ReloadServiceDaemon(), convey.ShouldResemble, test.ErrTest)
		})
	})

	convey.Convey("test func ResetFailedService", t, func() {
		convey.Convey("test func ResetFailedService success", func() {
			convey.So(ResetFailedService(), convey.ShouldResemble, nil)
		})
		convey.Convey("test func ResetFailedService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(ResetFailedService(), convey.ShouldResemble, test.ErrTest)
		})
	})
}

func TestInitService(t *testing.T) {
	patch := gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
	defer patch.Reset()

	convey.Convey("test func StartService", t, func() {
		convey.Convey("test func StartService success", func() {
			convey.So(StartService("test.service"), convey.ShouldResemble, nil)
		})
		convey.Convey("test func StartService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(StartService("test.service"), convey.ShouldResemble, test.ErrTest)
		})
	})

	convey.Convey("test func StopService", t, func() {
		convey.Convey("test func StopService success", func() {
			convey.So(StopService("test.service"), convey.ShouldResemble, nil)
		})
		convey.Convey("test func StopService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(StopService("test.service"), convey.ShouldResemble, test.ErrTest)
		})
	})

	convey.Convey("test func RestartService", t, func() {
		convey.Convey("test func RestartService success", func() {
			convey.So(RestartService("test.service"), convey.ShouldResemble, nil)
		})
		convey.Convey("test func RestartService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(RestartService("test.service"), convey.ShouldResemble, test.ErrTest)
		})
	})
}

func TestDoService(t *testing.T) {
	patch := gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
	defer patch.Reset()

	convey.Convey("test func EnableService", t, func() {
		convey.Convey("test func EnableService success", func() {
			convey.So(EnableService("test.service"), convey.ShouldResemble, nil)
		})
		convey.Convey("test func EnableService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(EnableService("test.service"), convey.ShouldResemble, test.ErrTest)
		})
	})

	convey.Convey("test func DisableService", t, func() {
		convey.Convey("test func DisableService success", func() {
			convey.So(DisableService("test.service"), convey.ShouldResemble, nil)
		})
		convey.Convey("test func DisableService failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(DisableService("test.service"), convey.ShouldResemble, test.ErrTest)
		})
	})

	convey.Convey("test func IsServiceEnabled", t, func() {
		convey.Convey("test func IsServiceEnabled success", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, constants.SystemctlEnabled, nil)
			defer p1.Reset()
			convey.So(IsServiceEnabled("test.service"), convey.ShouldBeTrue)
		})
		convey.Convey("test func IsServiceEnabled failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(IsServiceEnabled("test.service"), convey.ShouldBeFalse)
		})
	})

	convey.Convey("test func IsServiceActive", t, func() {
		convey.Convey("test func IsServiceActive success", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, constants.SystemctlStatusActive, nil)
			defer p1.Reset()
			convey.So(IsServiceActive("test.service"), convey.ShouldBeTrue)
		})
		convey.Convey("test func IsServiceActive failed, run command failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(IsServiceActive("test.service"), convey.ShouldBeFalse)
		})
	})
}

func TestCopyServiceFileToSystemd(t *testing.T) {
	tempFile := CreateLogflie(t)
	defer func() {
		if err := os.RemoveAll(tempFile); err != nil {
			t.Fatal(err)
		}
	}()
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	patch := gomonkey.ApplyFunc(fileutils.CheckOwnerAndPermission, mockCheckOwnerAndPermission).
		//无权限对system写操作
		ApplyFunc(fileutils.CopyFile, mocCopyFile)
	defer patch.Reset()

	convey.Convey("test func CopyServiceFileToSystemd success", t, func() {
		convey.So(CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username), convey.ShouldResemble, nil)
	})

	convey.Convey("test func CopyServiceFileToSystemd failed, file is not exist", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p1.Reset()
		err = CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("service file[%s] not exist", tempFile))
	})

	convey.Convey("test func CopyServiceFileToSystemd failed, get uid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err = CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get user id faild, error: %v", test.ErrTest))
	})

	convey.Convey("test func CopyServiceFileToSystemd failed, check owner and permission failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p1.Reset()
		err = CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username)
		expErr := fmt.Errorf("check service file owner and permission failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func CopyServiceFileToSystemd failed, eval symlink failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		err = CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username)
		expErr := fmt.Errorf("get abs srv path failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func CopyServiceFileToSystemd failed, copy file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p1.Reset()
		err = CopyServiceFileToSystemd(tempFile, constants.Mode700, currentUser.Username)
		expErr := fmt.Errorf("copy service file to systemd failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func mocCopyFile(_, _ string, _ ...fileutils.FileChecker) error {
	return nil
}

func TestRemoveServiceFileInSystemd(t *testing.T) {
	tempFile := CreateLogflie(t)
	defer func() {
		if err := os.RemoveAll(tempFile); err != nil {
			t.Fatal(err)
		}
	}()
	patch := gomonkey.ApplyFunc(fileutils.DeleteFile, mocDeleteFile).
		ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
	defer patch.Reset()

	convey.Convey("TestRemoveServiceFileInSystemd", t, func() {
		err := RemoveServiceFileInSystemd("")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func RemoveServiceFileInSystemd failed, eval symlink failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		err := RemoveServiceFileInSystemd("")
		expErr := fmt.Errorf("get abs service dir failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func RemoveServiceFileInSystemd failed, file is not exist", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p1.Reset()
		err := RemoveServiceFileInSystemd("")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func RemoveServiceFileInSystemd failed, delete file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p1.Reset()
		err := RemoveServiceFileInSystemd("")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func mocDeleteFile(_ string, _ ...fileutils.FileChecker) error {
	return nil
}

func TestReplaceValueInService(t *testing.T) {

	tempFile := CreateJsonFile(t)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()
	replaceDic := map[string]string{
		constants.InstallEdgeDir:     tempFile.Name(),
		constants.LogEdgeDir:         tempFile.Name(),
		constants.LogBackupDirName:   tempFile.Name(),
		constants.InstallSoftWareDir: tempFile.Name(),
	}
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	patch := gomonkey.ApplyFunc(fileutils.CheckOwnerAndPermission, mockCheckOwnerAndPermission)
	defer patch.Reset()

	convey.Convey("TestReplaceValueInService", t, func() {
		err = ReplaceValueInService(tempFile.Name(), constants.Mode700, currentUser.Username, replaceDic)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func ReplaceValueInService failed, get uid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err = ReplaceValueInService(tempFile.Name(), constants.Mode700, currentUser.Username, replaceDic)
		expErr := fmt.Errorf("get user id faild, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func ReplaceValueInService failed, check owner and permission failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p1.Reset()
		err = ReplaceValueInService(tempFile.Name(), constants.Mode700, currentUser.Username, replaceDic)
		expErr := fmt.Errorf("systemd service file[%s] is invalid, error: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetExecStartInService(t *testing.T) {
	tempFile := CreateJsonFile(t)
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()
	currentUser, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	uid, err := envutils.GetUid(currentUser.Username)
	if err != nil {
		t.Fatal(err)
	}
	patch := gomonkey.ApplyFunc(fileutils.CheckOwnerAndPermission, mockCheckOwnerAndPermission)
	defer patch.Reset()

	convey.Convey("TestGetExecStartInService", t, func() {
		_, err = GetExecStartInService(tempFile.Name(), constants.Mode700, uid)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetExecStartInService failed, check owner and permission failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p1.Reset()
		_, err = GetExecStartInService(tempFile.Name(), constants.Mode700, uid)
		expErr := fmt.Errorf("systemd service file[%s] is invalid, error: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetExecStartInService failed, load file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte(""), test.ErrTest)
		defer p1.Reset()
		_, err = GetExecStartInService(tempFile.Name(), constants.Mode700, uid)
		expErr := fmt.Errorf("load systemd service file[%s] failed, error: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
