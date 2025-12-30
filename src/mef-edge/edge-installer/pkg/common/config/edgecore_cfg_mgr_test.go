// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for edgecore config manager
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	testJson = `
{
  "database": {
    "dataSource": "/var/lib/kubeedge/edgecore.db"
  },
  "kind": "EdgeCore",
  "modules": {
    "deviceTwin": {},
    "edgeHub": {
	  "tlsCaFile": "/etc/kubeedge/ca/rootCA.crt",
	  "tlsCertFile": "/etc/kubeedge/certs/server.crt",
	  "tlsPrivateKeyFile": "/run/edgecore.pipe"
    },
    "edged": {
		"hostnameOverride": "",
		"nodeLabels": {"serialNumber": ""},
		"rootDirectory": "/var/lib/docker/kubelet",
		"tailoredKubeletConfig": {
			"readOnlyPort": 0,
			"serverTLSBootstrap": true,
			"evictionHard": {
			  "imagefs.available": "0%",
			  "memory.available": "0%",
			  "nodefs.available": "10%",
			  "nodefs.inodesFree": "5%"
			},
			"cgroupDriver": "cgroupfs",
			"systemReserved":""
        }
    },
	"eventBus": {
      "enable": false
    }
  }
}`
	serialNumber = "AN428394huawei"
	cgroupDriver = "cgroupfs"
	jsonFilePath = filepath.Join("./", "edgecore.json")
)

func prepareJsonFile() string {
	if err := os.Remove(jsonFilePath); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup edgecore json file failed, error: %v", err)
		return ""
	}
	if err := os.WriteFile(jsonFilePath, []byte(testJson), constants.Mode640); err != nil {
		fmt.Printf("write file failed, error: %v\n", err)
		return ""
	}
	jsonFileRealPath, err := filepath.Abs(jsonFilePath)
	if err != nil {
		fmt.Printf("get json abs path failed: %v\n", err)
	}
	return jsonFileRealPath
}

func TestSetDatabase(t *testing.T) {
	convey.Convey("set database should be success", t, testSetDatabase)
	convey.Convey("set database should be failed, load json file failed", t, testSetDatabaseErrLoadJsonFile)
	convey.Convey("set database should be failed, set json value failed", t, testSetDatabaseErrSetJson)
	convey.Convey("set database should be failed, save json value failed", t, testSetDatabaseErrSaveJson)
}

func testSetDatabase() {
	jsonFile := prepareJsonFile()
	dataSource := filepath.Join("./", "edgecore.db")
	err := SetDatabase(jsonFile, dataSource)
	convey.So(err, convey.ShouldBeNil)
}

func testSetDatabaseErrLoadJsonFile() {
	jsonFile := prepareJsonFile()
	dataSource := filepath.Join("./", "edgecore.db")
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SetDatabase(jsonFile, dataSource)
	convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
}

func testSetDatabaseErrSetJson() {
	jsonFile := prepareJsonFile()
	dataSource := filepath.Join("./", "edgecore.db")
	var p1 = gomonkey.ApplyFunc(util.SetJsonValue,
		func(object map[string]interface{}, value interface{}, names ...string) error {
			return testErr
		})
	defer p1.Reset()
	err := SetDatabase(jsonFile, dataSource)
	convey.So(err, convey.ShouldResemble, errors.New("set database value failed"))
}

func testSetDatabaseErrSaveJson() {
	jsonFile := prepareJsonFile()
	dataSource := filepath.Join("./", "edgecore.db")
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SetDatabase(jsonFile, dataSource)
	convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
}

func TestSetCertPath(t *testing.T) {
	convey.Convey("set cert path should be success", t, testSetCertPath)
	convey.Convey("set cert path should be failed, load json file failed", t, testSetCertPathErrLoadJson)
	convey.Convey("set cert path should be failed, set json value failed", t, testSetCertPathErrSetJson)
	convey.Convey("set cert path should be failed, save json value failed", t, testSetCertPathErrSaveJson)
}

func testSetCertPath() {
	jsonFile := prepareJsonFile()
	err := SetCertPath(jsonFile, pathmgr.NewConfigPathMgr("./"))
	convey.So(err, convey.ShouldBeNil)
}

func testSetCertPathErrLoadJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SetCertPath(jsonFile, pathmgr.NewConfigPathMgr("./"))
	convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
}

func testSetCertPathErrSetJson() {
	jsonFile := prepareJsonFile()
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},
	}

	var p1 = gomonkey.ApplyFuncSeq(util.SetJsonValue, outputs)
	defer p1.Reset()
	err := SetCertPath(jsonFile, pathmgr.NewConfigPathMgr("./"))
	convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edgeHub.tlsCaFile failed"))
	err = SetCertPath(jsonFile, pathmgr.NewConfigPathMgr("./"))
	convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edgeHub.tlsCertFile failed"))
}

func testSetCertPathErrSaveJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SetCertPath(jsonFile, pathmgr.NewConfigPathMgr("./"))
	convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
}

func TestSetHostname(t *testing.T) {
	convey.Convey("set hostname should be success", t, testSetHostname)
	convey.Convey("set hostname should be failed, load json file failed", t, testSetHostnameErrLoadJson)
	convey.Convey("set hostname should be failed, set json value failed", t, testSetHostnameErrSetJson)
	convey.Convey("set hostname should be failed, save json value failed", t, testSetHostnameErrSaveJson)
}

func testSetHostname() {
	jsonFile := prepareJsonFile()
	err := SetHostname(jsonFile, strings.ToLower(serialNumber))
	convey.So(err, convey.ShouldBeNil)
}

func testSetHostnameErrLoadJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SetHostname(jsonFile, strings.ToLower(serialNumber))
	convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
}

func testSetHostnameErrSetJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.SetJsonValue,
		func(object map[string]interface{}, value interface{}, names ...string) error {
			return testErr
		})
	defer p1.Reset()
	err := SetHostname(jsonFile, strings.ToLower(serialNumber))
	convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edged.hostnameOverride failed"))
}

func testSetHostnameErrSaveJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SetHostname(jsonFile, strings.ToLower(serialNumber))
	convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
}

func TestSetSerialNumber(t *testing.T) {
	convey.Convey("set sn should be success", t, testSetSerialNumber)
	convey.Convey("set sn should be failed, load json file failed", t, testSetSerialNumberErrLoadJson)
	convey.Convey("set sn should be failed, set json value failed", t, testSetSerialNumberErrSetJson)
	convey.Convey("set sn should be failed, save json value failed", t, testSetSerialNumberErrSaveJson)
}

func testSetSerialNumber() {
	jsonFile := prepareJsonFile()
	err := SetSerialNumber(jsonFile, serialNumber)
	convey.So(err, convey.ShouldBeNil)
}

func testSetSerialNumberErrLoadJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SetSerialNumber(jsonFile, serialNumber)
	convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
}

func testSetSerialNumberErrSetJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.SetJsonValue,
		func(object map[string]interface{}, value interface{}, names ...string) error {
			return testErr
		})
	defer p1.Reset()
	err := SetSerialNumber(jsonFile, serialNumber)
	convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edged.nodeLabels.serialNumber failed"))
}

func testSetSerialNumberErrSaveJson() {
	jsonFile := prepareJsonFile()
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SetSerialNumber(jsonFile, serialNumber)
	convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
}

func TestSetCgroupDriver(t *testing.T) {
	convey.Convey("test func SetCgroupDriver success", t, func() {
		jsonFile := prepareJsonFile()
		err := SetCgroupDriver(jsonFile, cgroupDriver)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SetCgroupDriver failed, load json file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.LoadJsonFile, nil, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetCgroupDriver(jsonFile, cgroupDriver)
		convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
	})

	convey.Convey("test func SetCgroupDriver failed, set json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SetJsonValue, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetCgroupDriver(jsonFile, cgroupDriver)
		expErr := errors.New("set value for modules.edged.tailoredKubeletConfig.cgroupDriver failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SetCgroupDriver failed, save json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SaveJsonValue, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetCgroupDriver(jsonFile, cgroupDriver)
		convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
	})
}

func TestSetNodeIP(t *testing.T) {
	const nodeIP = "127.0.0.1"
	convey.Convey("test func SetNodeIP success", t, func() {
		jsonFile := prepareJsonFile()
		err := SetNodeIP(jsonFile, nodeIP)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SetCgroupDriver failed, load json file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.LoadJsonFile, nil, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetNodeIP(jsonFile, nodeIP)
		expErr := fmt.Errorf("get edge core config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SetCgroupDriver failed, set json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SetJsonValue, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetNodeIP(jsonFile, nodeIP)
		expErr := fmt.Errorf("set value for modules.edged.nodeIP failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SetCgroupDriver failed, save json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SaveJsonValue, test.ErrTest)
		defer p1.Reset()
		jsonFile := prepareJsonFile()
		err := SetNodeIP(jsonFile, nodeIP)
		expErr := fmt.Errorf("save edge core config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestSmoothEdgeCoreConfigSystemReserve(t *testing.T) {
	containerConfig := ContainerConfig{
		SystemReservedCPUQuota:    0,
		SystemReservedMemoryQuota: 0,
	}
	podConfig := &PodConfig{
		PodSecurityConfig: PodSecurityConfig{},
		ContainerConfig:   containerConfig,
	}
	var patches = gomonkey.ApplyFuncReturn(LoadPodConfig, podConfig, nil).
		ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetEdgeCoreConfigPath", jsonFilePath)
	defer patches.Reset()

	prepareCfgDir()
	convey.Convey("test func SmoothEdgeCoreConfigSystemReserve success", t, func() {
		prepareJsonFile()
		err := SmoothEdgeCoreConfigSystemReserve(testPath, false)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SmoothEdgeCoreConfigSystemReserve failed, load pod config failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(LoadPodConfig, nil, test.ErrTest)
		defer p1.Reset()
		err := SmoothEdgeCoreConfigSystemReserve(testPath, false)
		expErr := fmt.Errorf("set edgecore config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SmoothEdgeCoreConfigSystemReserve failed, load json file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.LoadJsonFile, nil, test.ErrTest)
		defer p1.Reset()
		err := SmoothEdgeCoreConfigSystemReserve(testPath, false)
		convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
	})

	convey.Convey("test func SmoothEdgeCoreConfigSystemReserve failed, set json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SetJsonValue, test.ErrTest)
		defer p1.Reset()
		err := SmoothEdgeCoreConfigSystemReserve(testPath, false)
		convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edged.tailoredKubeletConfig failed"))
	})

	convey.Convey("test func SmoothEdgeCoreConfigSystemReserve failed, save json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SaveJsonValue, test.ErrTest)
		defer p1.Reset()
		err := SmoothEdgeCoreConfigSystemReserve(testPath, false)
		convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
	})
}

func TestSetOldKubeletRootDir(t *testing.T) {
	var patches = gomonkey.ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetEdgeCoreConfigPath", jsonFilePath)
	defer patches.Reset()

	convey.Convey("test func setOldKubeletRootDir success", t, func() {
		prepareJsonFile()
		err := setOldKubeletRootDir(testPath)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func setOldKubeletRootDir failed, load json file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.LoadJsonFile, nil, test.ErrTest)
		defer p1.Reset()
		err := setOldKubeletRootDir(testPath)
		convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
	})

	convey.Convey("test func setOldKubeletRootDir failed, set json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SetJsonValue, test.ErrTest)
		defer p1.Reset()
		err := setOldKubeletRootDir(testPath)
		convey.So(err, convey.ShouldResemble, errors.New("set edge core filed failed"))
	})

	convey.Convey("test func setOldKubeletRootDir failed, save json value failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(util.SaveJsonValue, test.ErrTest)
		defer p1.Reset()
		err := setOldKubeletRootDir(testPath)
		convey.So(err, convey.ShouldResemble, errors.New("save edge core config failed"))
	})
}

func TestEffectToOldestVersionSmooth(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(SmoothEdgeCoreConfigPipePath, nil).
		ApplyFuncReturn(SmoothEdgeCoreConfigSystemReserve, nil).
		ApplyFuncReturn(setOldKubeletRootDir, nil)
	defer patches.Reset()

	convey.Convey("test func EffectToOldestVersionSmooth success", t, func() {
		err := EffectToOldestVersionSmooth(constants.Version5Rc1, testPath)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func EffectToOldestVersionSmooth failed, smooth edgecore config pipe path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(SmoothEdgeCoreConfigPipePath, test.ErrTest)
		defer p1.Reset()
		err := EffectToOldestVersionSmooth(constants.Version5Rc1, testPath)
		expErr := fmt.Errorf("smooth old config to edge core config file failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func EffectToOldestVersionSmooth failed, smooth edgecore config system reserve failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(SmoothEdgeCoreConfigSystemReserve, test.ErrTest)
		defer p1.Reset()
		err := EffectToOldestVersionSmooth(constants.Version5Rc1, testPath)
		expErr := fmt.Errorf("smooth old system reserved config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func EffectToOldestVersionSmooth failed, set old kubelet root dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(setOldKubeletRootDir, test.ErrTest)
		defer p1.Reset()
		err := EffectToOldestVersionSmooth(constants.Version5Rc1, testPath)
		expErr := fmt.Errorf("smooth old config to edge core config file failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
