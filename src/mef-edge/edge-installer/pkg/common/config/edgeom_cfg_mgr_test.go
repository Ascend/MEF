// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for edgeom config manager
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

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	testOmCfg = `
{
  "hostPath": [
  ],
  "maxContainerNumber": 20,
  "containerModelFileNumber": 48,
  "totalModelFileNumber": 512,
  "systemReservedCPUQuota": 1,
  "systemReservedMemoryQuota": 1024
}`
)

func prepareOmCfg() {
	omCfgDir := filepath.Join(testPath, constants.MEFEdgeName, constants.Config, constants.EdgeOm)
	if err := os.MkdirAll(omCfgDir, constants.Mode700); err != nil {
		fmt.Printf("make dir [%s] failed, error: %v", omCfgDir, err)
		return
	}
	omCfgPath := filepath.Join(omCfgDir, constants.ContainerCfgFile)
	if err := os.Remove(omCfgPath); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup om config file failed, error: %v", err)
		return
	}
	if err := os.WriteFile(omCfgPath, []byte(testOmCfg), constants.Mode400); err != nil {
		fmt.Printf("write file failed, error: %v\n", err)
		return
	}
	return
}

func TestSmoothEdgeOmContainerConfig(t *testing.T) {
	prepareOmCfg()
	configPath := filepath.Join(testPath, constants.MEFEdgeName,
		constants.Config, constants.EdgeOm, constants.ContainerCfgFile)
	var patches = gomonkey.ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetContainerConfigPath", configPath)
	defer patches.Reset()
	convey.Convey("smooth edgeom container cfg should be success", t, testSmoothEdgeOmContainerConfig)
	convey.Convey("smooth edgeom container cfg should be failed, load json file failed", t, testSmoothOmErrLoadJson)
	convey.Convey("smooth edgeom container cfg should be failed, set json value failed", t, testSmoothOmErrSetJson)
	convey.Convey("smooth edgeom container cfg should be failed, marshal failed", t, testSmoothOmErrMarshal)
	convey.Convey("smooth edgeom container cfg should be failed, chmod failed", t, testSmoothOmErrChmod)
	convey.Convey("smooth edgeom container cfg should be failed, save json failed", t, testSmoothOmErrSaveJson)
}

func testSmoothEdgeOmContainerConfig() {
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldBeNil)
}

func testSmoothOmErrLoadJson() {
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("get edgeom container config failed"))
}

func testSmoothOmErrSetJson() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},
	}

	var p1 = gomonkey.ApplyFuncSeq(util.SetJsonValue, outputs)
	defer p1.Reset()
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("set value for maxContainerNumber failed"))
	err = SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("set value for hostPath failed"))
}

func testSmoothOmErrMarshal() {
	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p1.Reset()
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("marshal edgeOmContainerConfig failed"))
}

func testSmoothOmErrChmod() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},
	}

	var p1 = gomonkey.ApplyFuncSeq(os.Chmod, outputs).ApplyFuncSeq(fileutils.SetPathPermission, outputs)
	defer p1.Reset()
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("change edgeom container config path failed"))
	err = SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("post change edgeom container config path failed"))
}

func testSmoothOmErrSaveJson() {
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SmoothEdgeOmContainerConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("save edgeom container config failed"))
}

// The following ut belongs to the edgecore_cfg_mgr.go.
// The number of lines in the file must be less than 300. Therefore, the following ut is placed here.
func prepareCfgDir() {
	cfgDir := filepath.Join(testPath, constants.MEFEdgeName, constants.Config, constants.EdgeCore)
	if err := os.MkdirAll(cfgDir, constants.Mode700); err != nil {
		fmt.Printf("make dir [%s] failed, error: %v", cfgDir, err)
		return
	}
	cfgPath := filepath.Join(cfgDir, constants.EdgeCoreJsonName)
	if err := os.Remove(cfgPath); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup edgecore json file failed, error: %v", err)
		return
	}
	if err := os.WriteFile(cfgPath, []byte(testJson), constants.Mode640); err != nil {
		fmt.Printf("write file failed, error: %v\n", err)
		return
	}
}

func TestSmoothEdgeCoreConfigPipePath(t *testing.T) {
	prepareCfgDir()
	convey.Convey("smooth edgecore config pipe should be success", t, testSmoothEdgeCoreConfigPipePath)
	convey.Convey("smooth edgecore config pipe should be failed, load json file failed", t, testSmoothErrLoadJson)
	convey.Convey("smooth edgecore config pipe should be failed, set json value failed", t, testSmoothErrSetJson)
	convey.Convey("smooth edgecore config pipe should be failed, save json value failed", t, testSmoothErrSaveJson)
}

func testSmoothEdgeCoreConfigPipePath() {
	err := SmoothEdgeCoreConfigPipePath(testPath, "")
	convey.So(err, convey.ShouldBeNil)
}

func testSmoothErrLoadJson() {
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SmoothEdgeCoreConfigPipePath(testPath, "")
	convey.So(err, convey.ShouldResemble, errors.New("get edgecore config failed"))
}

func testSmoothErrSetJson() {
	var p1 = gomonkey.ApplyFunc(util.SetJsonValue,
		func(object map[string]interface{}, value interface{}, names ...string) error {
			return testErr
		})
	defer p1.Reset()
	err := SmoothEdgeCoreConfigPipePath(testPath, "")
	convey.So(err, convey.ShouldResemble, errors.New("set value for modules.edgeHub.tlsPrivateKeyFile failed"))
}

func testSmoothErrSaveJson() {
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SmoothEdgeCoreConfigPipePath(testPath, "")
	convey.So(err, convey.ShouldResemble, errors.New("save edgecore config failed"))
}

func TestSmoothEdgeCoreSafeConfig(t *testing.T) {
	prepareCfgDir()
	convey.Convey("smooth edgecore safe config should be success", t, testSmoothEdgeCoreSafeConfig)
	convey.Convey("smooth edgecore safe config should be failed, load json failed", t, testSmoothSafeCfgErrLoadJson)
	convey.Convey("smooth edgecore safe config should be failed, set json failed", t, testSmoothSafeCfgErrSetJson)
	convey.Convey("smooth edgecore safe config should be failed, save json failed", t, testSmoothSafeCfgErrSaveJson)
}

func testSmoothEdgeCoreSafeConfig() {
	err := SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldBeNil)
}

func testSmoothSafeCfgErrLoadJson() {
	var p1 = gomonkey.ApplyFunc(util.LoadJsonFile,
		func(jsonFilePath string) (map[string]interface{}, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("get edge core config failed"))
}

func testSmoothSafeCfgErrSetJson() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},
	}

	var p1 = gomonkey.ApplyFuncSeq(util.SetJsonValue, outputs)
	defer p1.Reset()
	err := SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)
	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)
	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)
	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)

	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)

	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)

	err = SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldNotEqual, nil)
}

func testSmoothSafeCfgErrSaveJson() {
	var p1 = gomonkey.ApplyFunc(util.SaveJsonValue,
		func(jsonFilePath string, jsonValue map[string]interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := SmoothEdgeCoreSafeConfig(testPath)
	convey.So(err, convey.ShouldResemble, errors.New("save edge core config failed"))
}

func TestSmoothAlarmConfigDB(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), nil).
		ApplyFuncReturn(database.CreateTableIfNotExist, nil).
		ApplyFuncReturn(SetDefaultAlarmCfg, nil)
	defer patches.Reset()

	convey.Convey("test func SmoothAlarmConfigDB success", t, func() {
		// table exists
		var p1 = gomonkey.ApplyFuncReturn(database.GetDb().Migrator().HasTable, true)
		defer p1.Reset()
		err := SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldBeNil)

		// table doesn't exist
		var p2 = gomonkey.ApplyFuncReturn(database.GetDb().Migrator().HasTable, false)
		defer p2.Reset()
		err = SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SmoothAlarmConfigDB failed, get component db manager failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), test.ErrTest)
		defer p1.Reset()
		err := SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldResemble, errors.New("get config path manager failed"))
	})

	convey.Convey("test func SmoothAlarmConfigDB failed, init db failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&DbMgr{}, "InitDB", test.ErrTest)
		defer p1.Reset()
		err := SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldResemble, errors.New("init alarm manager database failed"))
	})

	convey.Convey("test func SmoothAlarmConfigDB failed, CreateTableIfNotExist failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(database.GetDb().Migrator().HasTable, true).
			ApplyFuncReturn(database.CreateTableIfNotExist, test.ErrTest)
		defer p1.Reset()
		// table exists
		err := SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldResemble, errors.New("create alarm config table failed"))

		// table doesn't exist
		var p2 = gomonkey.ApplyFuncReturn(database.GetDb().Migrator().HasTable, false)
		defer p2.Reset()
		err = SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldResemble, errors.New("create alarm config table failed"))
	})

	convey.Convey("test func SmoothAlarmConfigDB failed, SetDefaultAlarmCfg failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(database.GetDb().Migrator().HasTable, false).
			ApplyFuncReturn(SetDefaultAlarmCfg, test.ErrTest)
		defer p1.Reset()
		// table doesn't exist
		err := SmoothAlarmConfigDB()
		convey.So(err, convey.ShouldResemble, errors.New("set default alarm config to table failed"))
	})
}

func TestSetDefaultAlarmCfg(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), nil).
		ApplyMethodReturn(&DbMgr{}, "SetAlarmConfig", nil)
	defer patches.Reset()

	convey.Convey("test func SetDefaultAlarmCfg success", t, func() {
		err := SetDefaultAlarmCfg()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func SetDefaultAlarmCfg failed, get component db manager failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), test.ErrTest)
		defer p1.Reset()
		err := SetDefaultAlarmCfg()
		convey.So(err, convey.ShouldResemble, errors.New("get config path manager failed"))
	})

	convey.Convey("test func SetDefaultAlarmCfg failed, set alarm config to db failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&DbMgr{}, "SetAlarmConfig", test.ErrTest)
		defer p1.Reset()
		err := SetDefaultAlarmCfg()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set alarm config %s failed", constants.CertCheckPeriodDB))
	})
}
