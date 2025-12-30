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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

func TestGetKmcConfig(t *testing.T) {
	const testKmcDir = "/tmp/kmc"
	patches := gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, testKmcDir, nil).
		ApplyFuncReturn(fileutils.CreateDir, nil).
		ApplyFuncReturn(fileutils.MakeSureDir, nil)
	defer patches.Reset()

	convey.Convey("test func GetKmcConfig success", t, func() {
		masterKmcPath := filepath.Join(testKmcDir, constants.KmcMasterName)
		backupKmcPath := filepath.Join(testKmcDir, constants.KmcBackupName)
		const (
			sdpAlgID = kmc.Aes256gcmId
			doMainId = kmc.DefaultDoMainId
		)
		expKmcConfig := &kmc.SubConfig{
			SdpAlgID:       sdpAlgID,
			PrimaryKeyPath: masterKmcPath,
			StandbyKeyPath: backupKmcPath,
			DoMainId:       doMainId,
		}

		kmcConfig, err := GetKmcConfig("")
		convey.So(kmcConfig, convey.ShouldResemble, expKmcConfig)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetKmcConfig failed, get component specific dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", test.ErrTest)
		defer p1.Reset()
		kmcConfig, err := GetKmcConfig("")
		convey.So(kmcConfig, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func GetKmcConfig failed, create dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		kmcConfig, err := GetKmcConfig("")
		convey.So(kmcConfig, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func GetKmcConfig failed, make sure dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, test.ErrTest)
		defer p1.Reset()
		kmcConfig, err := GetKmcConfig("")
		convey.So(kmcConfig, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestGetProcesses(t *testing.T) {
	convey.Convey("test func GetProcesses success", t, func() {
		_, err := GetProcesses()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func GetProcesses failed, read /proc failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.ReadDir, nil, test.ErrTest)
		defer p1.Reset()
		_, err := GetProcesses()
		expErr := fmt.Errorf("read proc directory failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestIsValidVersion(t *testing.T) {
	convey.Convey("test func IsValidVersion", t, func() {
		convey.Convey("valid version", func() {
			isValidVersion, err := IsValidVersion("1.0", "1.0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(isValidVersion, convey.ShouldBeTrue)

			isValidVersion, err = IsValidVersion("1.0", "2.0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(isValidVersion, convey.ShouldBeTrue)

			isValidVersion, err = IsValidVersion("2.0", "1.0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(isValidVersion, convey.ShouldBeTrue)
		})

		convey.Convey("invalid version", func() {
			isValidVersion, err := IsValidVersion("1.0", "3.0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(isValidVersion, convey.ShouldBeFalse)

			isValidVersion, err = IsValidVersion("3.0", "1.0")
			convey.So(err, convey.ShouldBeNil)
			convey.So(isValidVersion, convey.ShouldBeFalse)

			// strconv.Atoi failed
			outputs := []gomonkey.OutputCell{
				{Values: gomonkey.Params{0, test.ErrTest}},

				{Values: gomonkey.Params{0, nil}},
				{Values: gomonkey.Params{0, test.ErrTest}},
			}
			var p1 = gomonkey.ApplyFuncSeq(strconv.Atoi, outputs)
			defer p1.Reset()

			isValidVersion, err = IsValidVersion("1.0", "3.0")
			convey.So(isValidVersion, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldResemble, test.ErrTest)

			isValidVersion, err = IsValidVersion("3.0", "1.0")
			convey.So(isValidVersion, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldResemble, test.ErrTest)
		})
	})
}

func TestGetProcName(t *testing.T) {
	convey.Convey("test func GetProcName success", t, func() {
		const testProcFileData = "test proc file data"
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, []byte("test proc file data"), nil)
		defer p1.Reset()
		procName, err := GetProcName(1)
		expRes := strings.TrimSuffix(testProcFileData, "\n")
		convey.So(procName, convey.ShouldResemble, expRes)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetProcName failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p1.Reset()
		procName, err := GetProcName(1)
		convey.So(procName, convey.ShouldResemble, "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestCheckProcUser(t *testing.T) {
	convey.Convey("TestCheckProcUser", t, func() {
		convey.So(CheckProcUser(1, constants.RootUserName), convey.ShouldBeTrue)
		convey.So(CheckProcUser(1, constants.DockerUserName), convey.ShouldBeFalse)
	})
}

func TestCheckProcGroup(t *testing.T) {
	convey.Convey("TestCheckProcGroup", t, func() {
		convey.So(CheckProcGroup(1, constants.RootUserName), convey.ShouldBeTrue)
		convey.So(CheckProcUser(1, constants.DockerUserName), convey.ShouldBeFalse)

		convey.Convey("test func CheckProcGroup failed, get gid failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.GetGid, uint32(0), test.ErrTest)
			defer p1.Reset()
			convey.So(CheckProcGroup(1, constants.RootUserName), convey.ShouldBeFalse)
		})

		convey.Convey("test func CheckProcGroup failed, read limit bytes failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(CheckProcGroup(1, constants.RootUserName), convey.ShouldBeFalse)
		})
	})
}

func TestIsFlagSet(t *testing.T) {
	convey.Convey("TestIsFlagSet", t, func() {
		convey.So(IsFlagSet(""), convey.ShouldResemble, false)
	})
}

func TestIsProcessActive(t *testing.T) {
	convey.Convey("test func IsProcessActive success", t, func() {
		pid := os.Getpid()
		convey.So(IsProcessActive(pid), convey.ShouldResemble, true)
	})

	convey.Convey("test func IsProcessActive failed, find process failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.FindProcess, nil, test.ErrTest)
		defer p1.Reset()
		pid := os.Getpid()
		convey.So(IsProcessActive(pid), convey.ShouldResemble, false)
	})
}

func TestGetUuid(t *testing.T) {
	patch := gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
	defer patch.Reset()
	convey.Convey("TestGetUuid", t, func() {
		_, err := GetUuid()
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetUuid failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		uuid, err := GetUuid()
		convey.So(uuid, convey.ShouldResemble, "")
		expErr := fmt.Errorf("get uuid failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestWatchAndUpdateCPUTransientUsage(t *testing.T) {
	convey.Convey("test func WatchAndUpdateCPUTransientUsage", t, func() {
		WatchAndUpdateCPUTransientUsage()

		contents, err := fileutils.LoadFile(statInfoPath)
		convey.So(err, convey.ShouldBeNil)

		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil, test.ErrTest}},

			{Values: gomonkey.Params{contents, nil}},
			{Values: gomonkey.Params{nil, test.ErrTest}},

			{Values: gomonkey.Params{contents, nil}, Times: 2},
		}
		var p1 = gomonkey.ApplyFuncSeq(fileutils.LoadFile, outputs)
		defer p1.Reset()
		WatchAndUpdateCPUTransientUsage()
		WatchAndUpdateCPUTransientUsage()

		var p2 = gomonkey.ApplyFuncReturn(strings.SplitN, []string{})
		defer p2.Reset()
		WatchAndUpdateCPUTransientUsage()
	})
}

func TestSystemVariables(t *testing.T) {
	convey.Convey("TestIsSystemMemoryEnough", t, func() {
		_, err := IsSystemMemoryEnough(0)
		convey.So(err, convey.ShouldResemble, nil)

		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p1.Reset()
		res, err := IsSystemMemoryEnough(0)
		convey.So(res, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("IsSystemCPUAvailable", t, func() {
		convey.So(IsSystemCPUAvailable(0), convey.ShouldResemble, false)
	})

	convey.Convey("TestIsSystemStorageEnough", t, func() {
		_, err := IsSystemStorageEnough("/", 0)
		convey.So(err, convey.ShouldResemble, nil)

		var p1 = gomonkey.ApplyFuncReturn(envutils.GetDiskFree, uint64(0), test.ErrTest)
		defer p1.Reset()
		res, err := IsSystemStorageEnough("/", 0)
		convey.So(res, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestRemoveContainer(t *testing.T) {
	convey.Convey("TestRemoveContainer", t, func() {
		testContainerId := "d67946562e34\n3da60b59d0e9\nfd80520edcfe"

		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, testContainerId, nil)
		defer p1.Reset()
		convey.So(RemoveContainer(), convey.ShouldBeNil)

		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{testContainerId, nil}},
			{Values: gomonkey.Params{"", test.ErrTest}, Times: 6},
		}
		var p2 = gomonkey.ApplyFuncSeq(envutils.RunCommand, outputs)
		defer p2.Reset()
		convey.So(RemoveContainer(), convey.ShouldBeNil)
	})

	convey.Convey("test func RemoveContainer success, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		convey.So(RemoveContainer(), convey.ShouldResemble, nil)
	})

	convey.Convey("test func RemoveContainer success, container id is null", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p1.Reset()
		convey.So(RemoveContainer(), convey.ShouldResemble, nil)
	})

}

func TestGetContentMap(t *testing.T) {
	convey.Convey("test func GetContentMap", t, func() {
		content := map[string]string{"key1": "value1", "key2": "value2"}
		bytes, err := json.Marshal(content)
		convey.So(err, convey.ShouldBeNil)

		var ret map[string]interface{}
		err = json.Unmarshal(bytes, &ret)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("test func GetContentMap success", func() {
			// content is byte
			contentMap, err := GetContentMap(bytes)
			convey.So(contentMap, convey.ShouldResemble, ret)
			convey.So(err, convey.ShouldBeNil)

			// content is string
			contentMap, err = GetContentMap(string(bytes))
			convey.So(contentMap, convey.ShouldResemble, ret)
			convey.So(err, convey.ShouldBeNil)

			// content is neither byte nor string
			contentMap, err = GetContentMap(content)
			convey.So(contentMap, convey.ShouldResemble, ret)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("test func GetContentMap failed, marshal failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(json.Marshal, nil, test.ErrTest)
			defer p1.Reset()
			contentMap, err := GetContentMap(content)
			convey.So(contentMap, convey.ShouldBeNil)
			expErr := fmt.Errorf("marshal interface to []byte failed: %v", err)
			convey.So(err, convey.ShouldResemble, expErr)
		})

		convey.Convey("test func GetContentMap failed, unmarshal failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
			defer p1.Reset()
			contentMap, err := GetContentMap(content)
			convey.So(contentMap, convey.ShouldBeNil)
			expErr := errors.New("convert content unmarshal err")
			convey.So(err, convey.ShouldResemble, expErr)
		})
	})
}

func TestGetBoolPointer(t *testing.T) {
	GetBoolPointer(true)
}

func TestEdgeGUidMgr(t *testing.T) {
	testEuid := os.Geteuid()
	testEgid := os.Getegid()
	p := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(testEuid), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(testEgid), nil)
	defer p.Reset()
	mgr := NewEdgeUGidMgr()

	convey.Convey("test func SetEUGidToEdge", t, func() {
		convey.Convey("test func SetEUGidToEdge success", func() {
			convey.So(mgr.SetEUGidToEdge(), convey.ShouldResemble, nil)
		})

		convey.Convey("test func SetEUGidToEdge failed, get uid failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
			defer p1.Reset()
			err := mgr.SetEUGidToEdge()
			expErr := fmt.Errorf("get mef-edge uid/gid failed, %v", test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})

		convey.Convey("test func SetEUGidToEdge failed, get gid failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(envutils.GetGid, uint32(0), test.ErrTest)
			defer p1.Reset()
			err := mgr.SetEUGidToEdge()
			expErr := fmt.Errorf("get mef-edge uid/gid failed, %v", test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})

		convey.Convey("test func SetEUGidToEdge failed, get egid failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(syscall.Setegid, test.ErrTest)
			defer p1.Reset()
			err := mgr.SetEUGidToEdge()
			expErr := fmt.Errorf("set gid to %d failed, %v", testEgid, test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})

		convey.Convey("test func SetEUGidToEdge failed, get euid failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(syscall.Seteuid, test.ErrTest)
			defer p1.Reset()
			err := mgr.SetEUGidToEdge()
			expErr := fmt.Errorf("set uid to %d failed, %v", testEuid, test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})
	})

	convey.Convey("Test EdgeGUidMgr, reset uid/gid to edge should success", t, func() {
		convey.So(mgr.ResetEUGid(), convey.ShouldResemble, nil)
	})
}

func TestSetEuidAndEgid(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(syscall.Seteuid, nil).
		ApplyFuncReturn(syscall.Setegid, nil)
	defer p.Reset()

	convey.Convey("Test SetEuidAndEgid, set uid/gid to edge should success", t, func() {
		convey.So(SetEuidAndEgid(0, 0), convey.ShouldResemble, nil)
	})
}

func TestCheckNecessaryCommands(t *testing.T) {
	convey.Convey("test func CheckNecessaryCommands should be success", t, func() {
		p := gomonkey.ApplyFuncReturn(envutils.CheckCommandAllowedSugid, nil)
		defer p.Reset()
		convey.So(CheckNecessaryCommands(), convey.ShouldBeNil)
	})

	convey.Convey("test func CheckNecessaryCommands should be failed", t, func() {
		p := gomonkey.ApplyFuncReturn(envutils.CheckCommandAllowedSugid, test.ErrTest)
		defer p.Reset()
		convey.So(CheckNecessaryCommands(), convey.ShouldNotBeNil)
	})
}
