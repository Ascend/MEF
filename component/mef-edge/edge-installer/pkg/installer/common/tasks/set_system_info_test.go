// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common test for tasks when installing and upgrading
package tasks

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	testSetSystemInfoDir = "/tmp/test_set_system_info"
	softwarePathMgr      = pathmgr.NewSoftwarePathMgr(testSetSystemInfoDir, "")
	setSystemInfo        = SetSystemInfoTask{
		ConfigDir:     softwarePathMgr.ConfigPathMgr.GetConfigDir(),
		ConfigPathMgr: softwarePathMgr.ConfigPathMgr,
		LogPathMgr:    pathmgr.NewLogPathMgr(testSetSystemInfoDir, testSetSystemInfoDir),
	}
)

func TestSetSystemInfoTask(t *testing.T) {
	prepareEdgeCoreJson()
	defer func() {
		if err := fileutils.DeleteAllFileWithConfusion(testSetSystemInfoDir); err != nil {
			hwlog.RunLog.Errorf("cleanup test dir failed, error: %v", err)
			return
		}
	}()
	patch := gomonkey.ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil).
		ApplyMethodReturn(backuputils.NewBackupDirMgr(""), "BackUp", nil)
	defer patch.Reset()

	convey.Convey("set system info should be success", t, testSetSystemInfoTask)
	convey.Convey("set system info should be failed, init db failed", t, testSetSystemInfoTaskErrInitDB)
	convey.Convey("set system info should be failed, set config failed", t, testSetSystemInfoTaskErrSetConfig)
	convey.Convey("set system info should be failed, get sn failed", t, testSetSystemInfoTaskErrGetSn)
	convey.Convey("set system info should be failed, get config failed", t, testSetSystemInfoTaskErrGetConfig)
	convey.Convey("set system info should be failed, set edgecore cfg failed", t, testSetSystemInfoTaskErrSetEdgeCoreCfg)
	convey.Convey("set system info should be failed, backup config failed", t, testSetSystemInfoTaskErrBackUpConfig)
}

var testEdgeCoreJson = `
{
  "database": {
    "dataSource": "/var/lib/kubeedge/edgecore.db"
  },
  "kind": "EdgeCore",
  "modules": {
    "edgeHub": {
	  "tlsCaFile": "/etc/kubeedge/ca/rootCA.crt",
	  "tlsCertFile": "/etc/kubeedge/certs/server.crt",
	  "tlsPrivateKeyFile": "/run/edgecore.pipe"
    },
    "edged": {
      "hostnameOverride": "",
	  "nodeLabels": {"serialNumber": ""},
	  "tailoredKubeletConfig": {
		"readOnlyPort": 0,
		"serverTLSBootstrap": true,
		"evictionHard": {
		  "imagefs.available": "0%",
		  "memory.available": "0%",
		  "nodefs.available": "10%",
		  "nodefs.inodesFree": "5%"
		}
      }
    }
  }
}`

func prepareEdgeCoreJson() {
	edgeCoreConfigDir := softwarePathMgr.ConfigPathMgr.GetCompConfigDir(constants.EdgeCore)
	jsonFile := softwarePathMgr.ConfigPathMgr.GetEdgeCoreConfigPath()
	if err := fileutils.DeleteFile(jsonFile); err != nil {
		hwlog.RunLog.Errorf("cleanup edgecore json file failed, error: %v", err)
		return
	}

	if err := fileutils.CreateDir(edgeCoreConfigDir, constants.Mode755); err != nil {
		hwlog.RunLog.Errorf("create edgecore config dir [%s] failed, error: %v", edgeCoreConfigDir, err)
		return
	}
	if err := os.WriteFile(jsonFile, []byte(testEdgeCoreJson), constants.Mode600); err != nil {
		hwlog.RunLog.Errorf("write file failed, error: %v", err)
		return
	}
}

func testSetSystemInfoTask() {
	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldBeNil)
}

func testSetSystemInfoTaskErrInitDB() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 2},
	}
	var p1 = gomonkey.ApplyMethodSeq(&config.DbMgr{}, "InitDB", outputs)
	defer p1.Reset()

	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("prepare edgecore database failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("init edge om database failed"))

	var p2 = gomonkey.ApplyFuncReturn(database.CreateTableIfNotExist, test.ErrTest)
	defer p2.Reset()

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("create table failed"))
}

func testSetSystemInfoTaskErrSetConfig() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},
	}
	var p1 = gomonkey.ApplyMethodSeq(&config.DbMgr{}, "SetConfig", outputs)
	defer p1.Reset()

	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("set install config failed, error: %v", test.ErrTest))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func testSetSystemInfoTaskErrGetSn() {
	var p1 = gomonkey.ApplyFuncReturn(util.GetSerialNumber, "", test.ErrTest)
	defer p1.Reset()
	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get serial number failed, error: %v", test.ErrTest))
}

func testSetSystemInfoTaskErrGetConfig() {
	var p1 = gomonkey.ApplyFuncReturn(config.GetInstall, nil, test.ErrTest)
	defer p1.Reset()
	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get install config failed, error: %v", test.ErrTest))
}

func testSetSystemInfoTaskErrSetEdgeCoreCfg() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 2},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 3},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 4},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 5},
		{Values: gomonkey.Params{test.ErrTest}},
	}
	var p1 = gomonkey.ApplyFuncSeq(util.SetJsonValue, outputs)
	defer p1.Reset()

	err := setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save database to edge core config file failed, error: "+
		"set database value failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save certPath to edge core config file failed, error: "+
		"set value for modules.edgeHub.tlsCaFile failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save certPath to edge core config file failed, error: "+
		"set value for modules.edgeHub.tlsCertFile failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save hostnameOverride to edge core config file failed, error: "+
		"set value for modules.edged.hostnameOverride failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save serialNumber to edge core config file failed, error: "+
		"set value for modules.edged.nodeLabels.serialNumber failed"))

	err = setSystemInfo.Run()
	convey.So(err, convey.ShouldResemble, errors.New("save cgroupDriver to edge core config file failed, error: "+
		"set value for modules.edged.tailoredKubeletConfig.cgroupDriver failed"))
}

func testSetSystemInfoTaskErrBackUpConfig() {
	convey.Convey("warning: create backup files failed", func() {
		p := gomonkey.ApplyMethodReturn(backuputils.NewBackupDirMgr(""), "BackUp", test.ErrTest)
		defer p.Reset()
		err := setSystemInfo.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set edge-main config dir owner failed", func() {
		p := gomonkey.ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, test.ErrTest)
		defer p.Reset()
		err := setSystemInfo.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
