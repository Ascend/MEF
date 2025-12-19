// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for pod config manager
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
)

func TestLoadPodConfig(t *testing.T) {
	err := os.Remove("/tmp/podCfg/config/edge_om/container-config.json")
	if err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup file failed, error: %v", err)
		return
	}
	err = os.Remove("/tmp/podCfg/config/edge_om/pod-config.json")
	if err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup file failed, error: %v", err)
		return
	}
	prepareContainerCfg()
	preparePodCfg()
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), nil).
		ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetConfigDir", "/tmp/config").
		ApplyPrivateMethod(backuputils.NewBackupFileMgr(""), "BackUp", func() error { return nil }).
		ApplyPrivateMethod(backuputils.NewBackupFileMgr(""), "Restore", func() error { return nil })

	defer p1.Reset()
	convey.Convey("load pod config should be success", t, testLoadPodConfig)
	convey.Convey("load pod config should be success, but check ori path failed", t, testLoadPodConfigErrCheckOriPath)
	convey.Convey("load pod cfg should be failed, get install root dir failed", t, testLoadPodCfgErrGetInstallRootDir)
	convey.Convey("load pod cfg should be failed, init config error", t, testLoadPodConfigErrInitConfig)
	convey.Convey("load pod cfg should be failed, load file error", t, testLoadPodConfigErrLoadFile)
	convey.Convey("load pod cfg should be failed, unmarshal error", t, testLoadPodConfigErrUnmarshal)
}

func testLoadPodConfig() {
	_, err := LoadPodConfig()
	convey.So(err, convey.ShouldBeNil)
}

func testLoadPodConfigErrCheckOriPath() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, nil, testErr)
	defer p1.Reset()
	_, err := LoadPodConfig()
	convey.So(err, convey.ShouldBeNil)
}

func testLoadPodCfgErrGetInstallRootDir() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr)
	defer p1.Reset()
	_, err := LoadPodConfig()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("load pod config failed, err: %s", testErr.Error()))
}

func testLoadPodConfigErrInitConfig() {
	var p1 = gomonkey.ApplyFuncSeq(backuputils.InitConfig, []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},
	})
	defer p1.Reset()
	_, err := LoadPodConfig()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("load pod config failed, err: %s", test.ErrTest.Error()))
	_, err = LoadPodConfig()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("load pod config failed, err: %s", test.ErrTest.Error()))
}

func testLoadPodConfigErrLoadFile() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
	defer p1.Reset()

	_, err := LoadPodConfig()
	innErr1 := fmt.Errorf("load container config file failed, %v", test.ErrTest)
	innErr2 := fmt.Errorf("init config [%s] failed after recovery, %v",
		pathmgr.NewConfigPathMgr("/tmp").GetContainerConfigPath(), innErr1)
	expErr := fmt.Errorf("load pod config failed, err: %s", innErr2.Error())
	convey.So(err, convey.ShouldResemble, expErr)
}

func testLoadPodConfigErrUnmarshal() {
	var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
	defer p1.Reset()

	_, err := LoadPodConfig()
	innErr1 := fmt.Errorf("unmarshal container config failed, %v", test.ErrTest)
	innErr2 := fmt.Errorf("init config [%s] failed after recovery, %v",
		pathmgr.NewConfigPathMgr("/tmp").GetContainerConfigPath(), innErr1)
	expErr := fmt.Errorf("load pod config failed, err: %s", innErr2.Error())
	convey.So(err, convey.ShouldResemble, expErr)
}

func prepareContainerCfg() {
	containerCfg := &ContainerConfig{
		HostPath:                  []string{"/etc/sys_version.conf", "/etc/hdcBasic.cfg"},
		MaxContainerNumber:        16,
		ContainerModelFileNumber:  48,
		TotalModelFileNumber:      512,
		SystemReservedCPUQuota:    1,
		SystemReservedMemoryQuota: 1024,
	}
	cfgContent, err := json.Marshal(containerCfg)
	if err != nil {
		fmt.Printf("marshal container config failed, error: %v\n", err)
		return
	}
	cfgDir := filepath.Join("/tmp", constants.Config, constants.EdgeOm)
	if err = os.MkdirAll(cfgDir, constants.Mode700); err != nil {
		fmt.Printf("make dir [%s] failed, error: %v", cfgDir, err)
		return
	}
	cfgFile := filepath.Join(cfgDir, constants.ContainerCfgFile)
	if err = os.WriteFile(cfgFile, cfgContent, constants.Mode400); err != nil {
		return
	}
}

func preparePodCfg() {
	podCfg := &PodSecurityConfig{
		HostPid:                  false,
		Capability:               false,
		Privileged:               false,
		AllowPrivilegeEscalation: false,
		RunAsRoot:                false,
		UseHostNetwork:           false,
		UseSeccomp:               false,
		UseDefaultContainerCap:   false,
		EmptyDirVolume:           false,
		AllowReadWriteRootFs:     false,
	}
	cfgContent, err := json.Marshal(podCfg)
	if err != nil {
		fmt.Printf("marshal pod config failed, error: %v\n", err)
		return
	}
	cfgDir := filepath.Join("/tmp", constants.Config, constants.EdgeOm)
	if err = os.MkdirAll(cfgDir, constants.Mode700); err != nil {
		fmt.Printf("make dir [%s] failed, error: %v", cfgDir, err)
		return
	}
	cfgFile := filepath.Join(cfgDir, constants.PodCfgFile)
	if err = os.WriteFile(cfgFile, cfgContent, constants.Mode400); err != nil {
		return
	}
}

func TestCheckHostPathMode(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(os.OpenFile, nil, nil).
		ApplyMethodReturn(&fileutils.FileModeChecker{}, "Check", nil).
		ApplyMethodReturn(&fileutils.FileOwnerChecker{}, "Check", nil)
	defer patches.Reset()

	convey.Convey("test func checkHostPathMode and checkHostPathParentOwner success", t, func() {
		err := checkHostPathMode(testPath)
		convey.So(err, convey.ShouldBeNil)

		err = checkHostPathParentOwner(testPath)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func checkHostPathMode and checkHostPathParentOwner failed, open file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.OpenFile, nil, test.ErrTest)
		defer p1.Reset()

		err := checkHostPathMode(testPath)
		expErr := fmt.Errorf("open file %s failed: %s", testPath, test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expErr)

		err = checkHostPathMode(testPath)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func checkHostPathMode and checkHostPathParentOwner failed, file check failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&fileutils.FileModeChecker{}, "Check", test.ErrTest)
		defer p1.Reset()

		err := checkHostPathMode(testPath)
		expErr := fmt.Errorf("check file %s failed: %s", testPath, test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expErr)

		err = checkHostPathMode(testPath)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
