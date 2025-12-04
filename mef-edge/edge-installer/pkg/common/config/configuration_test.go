// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for configuration
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	inputStr inputStruct
	testDB   gorm.DB
)

func TestSetNetManagerCache(t *testing.T) {
	convey.Convey("get net mgr and new db mgr should be success", t, func() {
		netCfg := NetManager{
			NetType: constants.FD,
			WithOm:  true,
		}
		err := SetNetManager(&dbMgr, &netCfg)
		convey.So(err, convey.ShouldBeNil)

		netConfig, err := GetNetManager(&dbMgr)
		if err != nil {
			fmt.Printf("get net manager config failed: %v", err)
			return
		}
		SetNetManagerCache(*netConfig)
		mgr := NewDbMgr("./", "test.db")
		convey.So(mgr.dbDir, convey.ShouldEqual, "./")
		convey.So(mgr.dbName, convey.ShouldEqual, "test.db")
	})
}

func TestInitDB(t *testing.T) {
	convey.Convey("init db should be success", t, testInitDB)
	convey.Convey("init db should be failed, make sure dir failed", t, testInitDBErrMakeSureDir)
	convey.Convey("init db should be failed, database init failed", t, testInitDBErrInitDb)
}

func testInitDB() {
	err := dbMgr.InitDB()
	convey.So(err, convey.ShouldBeNil)
}

func testInitDBErrMakeSureDir() {
	var p1 = gomonkey.ApplyFunc(fileutils.MakeSureDir,
		func(path string, _ ...fileutils.FileChecker) error {
			return testErr
		})
	defer p1.Reset()
	err := dbMgr.InitDB()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testInitDBErrInitDb() {
	var p1 = gomonkey.ApplyFuncReturn(database.InitDB, testErr)
	defer p1.Reset()
	err := dbMgr.InitDB()
	convey.So(err, convey.ShouldResemble, testErr)
}

var input = `{
   "name":"test01",
   "age":"20",
}`

func TestSetConfig(t *testing.T) {
	convey.Convey("set config should be success", t, testSetConfig)
	convey.Convey("set config should be failed, check and init db failed", t, testSetConfigErrCheck)
	convey.Convey("set config should be failed, marshal failed", t, testSetConfigErrMarshal)
	convey.Convey("set config should be failed, count failed", t, testSetConfigErrCount)
	convey.Convey("set config should be failed, update failed", t, testSetConfigErrUpdate)
	convey.Convey("set config should be failed, create failed", t, testSetConfigErrCreate)
}

func testSetConfig() {
	err := dbMgr.SetConfig("testKey", input)
	convey.So(err, convey.ShouldBeNil)
}

func testSetConfigErrCheck() {
	var c *DbMgr
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "checkAndInitDB",
		func(*DbMgr) error {
			return testErr
		})
	defer p1.Reset()
	err := dbMgr.SetConfig("testKey", input)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testSetConfigErrMarshal() {
	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p1.Reset()
	err := dbMgr.SetConfig("testKey", input)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testSetConfigErrCount() {
	testDB.Error = test.ErrTest
	var c *gorm.DB
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "Count",
		func(*gorm.DB, *int64) *gorm.DB {
			return &testDB
		})
	defer p1.Reset()
	err := dbMgr.SetConfig("testKey", input)
	convey.So(err, convey.ShouldResemble, errors.New("set config failed,count error"))
}

func testSetConfigErrUpdate() {
	testDB.Error = test.ErrTest
	var c *gorm.DB
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "Updates",
		func(*gorm.DB, interface{}) *gorm.DB {
			return &testDB
		})
	defer p1.Reset()
	err := dbMgr.SetConfig("testKey", input)
	convey.So(err.Error(), convey.ShouldContainSubstring, "set config failed")
}

func testSetConfigErrCreate() {
	testDB.Error = test.ErrTest
	var c *gorm.DB
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "Create",
		func(*gorm.DB, interface{}) *gorm.DB {
			return &testDB
		})
	defer p1.Reset()
	err := dbMgr.SetConfig("testKey1", input)
	convey.So(err, convey.ShouldResemble, errors.New("set config failed,create error"))
}

type inputStruct struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

func TestGetConfig(t *testing.T) {
	convey.Convey("get config should be success", t, testGetConfig)
	convey.Convey("get config should be failed, check and init db failed", t, testGetConfigErrCheck)
	convey.Convey("get config should be failed, unmarshal failed", t, testGetConfigErrUnmarshal)
	convey.Convey("get config should be failed, first failed", t, testGetConfigErrFirst)
	convey.Convey("get config should be failed, record not found", t, testGetConfigErrRecordNotFound)
}

func testGetConfig() {
	err := dbMgr.GetConfig("testKey", inputStr)
	convey.So(err, convey.ShouldBeNil)
}

func testGetConfigErrCheck() {
	var c *DbMgr
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "checkAndInitDB",
		func(*DbMgr) error {
			return testErr
		})
	defer p1.Reset()
	err := dbMgr.GetConfig("testKey", inputStr)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetConfigErrUnmarshal() {
	var p1 = gomonkey.ApplyFunc(json.Unmarshal,
		func(data []byte, v interface{}) error {
			return testErr
		})
	defer p1.Reset()
	err := dbMgr.GetConfig("testKey", inputStr)
	convey.So(err, convey.ShouldResemble, errors.New("unmarshal configuration value failed"))
}

func testGetConfigErrFirst() {
	testDB.Error = test.ErrTest
	var c *gorm.DB
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "First",
		func(*gorm.DB, interface{}, ...interface{}) *gorm.DB {
			return &testDB
		})
	defer p1.Reset()
	err := dbMgr.GetConfig("testKey", inputStr)
	convey.So(err, convey.ShouldResemble, errors.New("get configuration failed"))
}

func testGetConfigErrRecordNotFound() {
	testDB.Error = gorm.ErrRecordNotFound
	var c *gorm.DB
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "First",
		func(*gorm.DB, interface{}, ...interface{}) *gorm.DB {
			return &testDB
		})
	defer p1.Reset()
	err := dbMgr.GetConfig("testKey", inputStr)
	convey.So(err, convey.ShouldResemble, testDB.Error)
}

func TestSetAlarmConfig(t *testing.T) {
	alarmConfig := &AlarmConfig{
		ConfigName:  constants.CertCheckPeriodDB,
		ConfigValue: constants.DefaultCheckPeriod,
		HasModified: util.GetBoolPointer(false),
	}

	convey.Convey("test DbMgr method SetAlarmConfig success", t, func() {
		testDB.Error = nil
		var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Count", &testDB).
			ApplyMethodReturn(&gorm.DB{}, "Create", &testDB)
		defer p1.Reset()
		err := dbMgr.SetAlarmConfig(alarmConfig)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test DbMgr method SetAlarmConfig failed, check db failed failed", t, func() {
		var p1 = gomonkey.ApplyPrivateMethod(&DbMgr{}, "checkAndInitDB", func(*DbMgr) error { return testErr })
		defer p1.Reset()
		err := dbMgr.SetAlarmConfig(alarmConfig)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test DbMgr method SetAlarmConfig failed, count failed", t, func() {
		testDB.Error = test.ErrTest
		var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Count", &testDB)
		defer p1.Reset()
		err := dbMgr.SetAlarmConfig(alarmConfig)
		convey.So(err, convey.ShouldResemble, errors.New("get alarm config count failed"))
	})

	convey.Convey("test DbMgr method SetAlarmConfig failed, create failed", t, func() {
		testDB.Error = test.ErrTest
		testDB2 := testDB
		testDB2.Error = nil
		var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Count", &testDB2).
			ApplyMethodReturn(&gorm.DB{}, "Create", &testDB)
		defer p1.Reset()
		err := dbMgr.SetAlarmConfig(alarmConfig)
		convey.So(err, convey.ShouldResemble, errors.New("create alarm config failed"))
	})
}

func TestGetAlarmConfig(t *testing.T) {
	convey.Convey("test DbMgr method GetAlarmConfig success", t, func() {
		testDB.Error = nil
		var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "First", &testDB)
		defer p1.Reset()
		_, err := dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test DbMgr method GetAlarmConfig failed, check db failed failed", t, func() {
		var p1 = gomonkey.ApplyPrivateMethod(&DbMgr{}, "checkAndInitDB", func(*DbMgr) error { return testErr })
		defer p1.Reset()
		_, err := dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test DbMgr method GetAlarmConfig failed, find config in db failed", t, func() {
		testDB.Error = test.ErrTest
		var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "First", &testDB)
		defer p1.Reset()
		_, err := dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
		convey.So(err, convey.ShouldResemble, errors.New("get alarm config failed"))

		testDB.Error = gorm.ErrRecordNotFound
		var p2 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "First", &testDB)
		defer p2.Reset()
		_, err = dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("alarm config %s does not exist", constants.CertCheckPeriodDB))
	})
}

func TestGetComponentDbMgr(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, pathmgr.NewConfigPathMgr("/tmp"), nil)
	defer patches.Reset()

	convey.Convey("test func GetComponentDbMgr success", t, func() {
		dbMgr, err := GetComponentDbMgr(constants.EdgeMain)
		convey.So(dbMgr, convey.ShouldResemble, NewDbMgr("/tmp/MEFEdge/config/edge_main", constants.DbEdgeMainPath))
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func GetComponentDbMgr failed, component name error", t, func() {
		_, err := GetComponentDbMgr("error component")
		convey.So(err, convey.ShouldResemble, errors.New("get component db name failed"))
	})

	convey.Convey("test func GetComponentDbMgr failed, get config path manager failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
		defer p1.Reset()
		_, err := GetComponentDbMgr(constants.EdgeMain)
		convey.So(err, convey.ShouldResemble, errors.New("get config path manager failed"))
	})
}
