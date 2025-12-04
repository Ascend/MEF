// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr test for get config handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
)

var (
	getConfig           = getConfigHandler{}
	getConfigMsg        *model.Message
	getPodConfigMsg     *model.Message
	getNetMgrConfigMsg  *model.Message
	getInstallConfigMsg *model.Message
	getSoftwareCertMsg  *model.Message
	getAlarmConfigMsg   *model.Message
)

func TestGetPodConfig(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
		ApplyFuncReturn(path.GetCompConfigDir, "", nil)
	defer p.Reset()

	var err error
	getConfigMsg, err = newGetConfigMsg()
	if err != nil {
		fmt.Printf("new get pod config msg failed, error: %v\n", err)
		return
	}
	convey.Convey("test get pod config", t, testGetPodConfig)
	convey.Convey("test send response", t, testSendResponse)
	convey.Convey("test get net mgr config", t, testGetNetMgrConfig)
	convey.Convey("test func getNetConfig should be success", t, testGetNetConfig)
	convey.Convey("test func getNetConfig should be failed, get net mgr error", t, testGetNetConfigErrGetNetMgr)
	convey.Convey("test fun getNetConfig should be failed, decrypt error", t, testDecryptTokenForSdk)
	convey.Convey("test fun getInstallConfig", t, testGetInstallConfig)
	convey.Convey("test fun getSoftwareCert", t, testGetSoftwareCert)
	convey.Convey("test fun getAlarmCertConfig", t, testGetAlarmConfig)
	convey.Convey("test fun getEdgeOmCaps", t, testGetEdgeOmCaps)
	convey.Convey("test error config msg", t, testErrConfigMsg)
	convey.Convey("test error init db mgr", t, testErrInitDbMgr)
}

func testGetPodConfig() {
	getPodConfigMsg = getConfigMsg
	err := getPodConfigMsg.FillContent(constants.PodCfgResource)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getPodConfigMsg)
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p1.Reset()
	err = getConfig.Handle(getPodConfigMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testGetNetMgrConfig() {
	var p1 = gomonkey.ApplyFunc(getNetConfig,
		func(dbMgr *config.DbMgr) string {
			return ""
		})
	defer p1.Reset()

	getNetMgrConfigMsg = getConfigMsg
	err := getNetMgrConfigMsg.FillContent(constants.NetMgrConfigKey)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getNetMgrConfigMsg)
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyFuncReturn(path.GetCompConfigDir, "", testErr)
	defer p2.Reset()
	err = getConfig.Handle(getNetMgrConfigMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testSendResponse() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, testErr}},
		{Values: gomonkey.Params{getConfigMsg, nil}},
	}
	var p1 = gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, testErr).
		ApplyMethodSeq(&model.Message{}, "NewResponse", outputs)
	defer p1.Reset()

	getConfig.sendResponse(getConfigMsg, "")
	getConfig.sendResponse(getConfigMsg, "")
}

func testGetNetConfigErrGetNetMgr() {
	var p1 = gomonkey.ApplyFunc(config.GetNetManager,
		func(dbMgr *config.DbMgr) (*config.NetManager, error) {
			return nil, testErr
		})
	defer p1.Reset()

	mgr := config.NewDbMgr("./", "test.db")
	res := getNetConfig(mgr)
	convey.So(res, convey.ShouldResemble, constants.Failed)
}

func testGetInstallConfig() {
	installerConfig := &config.InstallerConfig{
		InstallDir:   "/tmp",
		SerialNumber: "Sn",
	}
	bytes, err := json.Marshal(installerConfig)
	if err != nil {
		fmt.Printf("marshal data failed: %v\n", err)
		return
	}
	outputs1 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{installerConfig, nil}, Times: 2},
		{Values: gomonkey.Params{nil, testErr}},
		{Values: gomonkey.Params{installerConfig, nil}},
	}
	outputs2 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{bytes, nil}, Times: 2},
		{Values: gomonkey.Params{nil, testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(json.Marshal, outputs2).
		ApplyFuncSeq(config.GetInstall, outputs1)
	defer p1.Reset()

	getInstallConfigMsg = getConfigMsg
	err = getInstallConfigMsg.FillContent(constants.InstallerConfigKey)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getInstallConfigMsg)
	convey.So(err, convey.ShouldBeNil)

	res := getConfig.getInstallConfig()
	convey.So(res, convey.ShouldResemble, string(bytes))
	res = getConfig.getInstallConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)
	res = getConfig.getInstallConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)
}

func testGetSoftwareCert() {
	patchConfigPathMgr := pathmgr.NewConfigPathMgr("/tmp")
	patchImageCertPath := filepath.Join("/tmp", constants.Config, constants.EdgeOm,
		constants.ImageCertPathName, constants.ImageCertFileName)
	outputs1 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{patchConfigPathMgr, nil}, Times: 2},
		{Values: gomonkey.Params{patchConfigPathMgr, testErr}},
		{Values: gomonkey.Params{patchConfigPathMgr, nil}},
	}
	outputs2 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, nil}, Times: 2},
		{Values: gomonkey.Params{nil, testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(path.GetConfigPathMgr, outputs1).
		ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetImageCertPath", patchImageCertPath).
		ApplyFuncSeq(certutils.GetCertContentWithBackup, outputs2)
	defer p1.Reset()

	getSoftwareCertMsg = getConfigMsg
	err := getSoftwareCertMsg.FillContent(constants.SoftwareCert)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getSoftwareCertMsg)
	convey.So(err, convey.ShouldBeNil)

	res := getConfig.getSoftwareCert()
	convey.So(res, convey.ShouldResemble, "")
	res = getConfig.getSoftwareCert()
	convey.So(res, convey.ShouldResemble, constants.Failed)
	res = getConfig.getSoftwareCert()
	convey.So(res, convey.ShouldResemble, constants.Failed)
}

func testGetAlarmConfig() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{10, nil}, Times: 4},
		{Values: gomonkey.Params{0, testErr}, Times: 2},
		{Values: gomonkey.Params{10, nil}},
		{Values: gomonkey.Params{0, testErr}},
		{Values: gomonkey.Params{10, nil}, Times: 2},
	}
	var p1 = gomonkey.ApplyMethodSeq(&config.DbMgr{}, "GetAlarmConfig", outputs)
	defer p1.Reset()
	getAlarmConfigMsg = getConfigMsg
	err := getAlarmConfigMsg.FillContent(constants.AlarmCertConfig)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getAlarmConfigMsg)
	convey.So(err, convey.ShouldBeNil)

	alarmCertCfg := config.AlarmCertCfg{
		CheckPeriod:      10,
		OverdueThreshold: 10,
	}
	bytes, err := json.Marshal(alarmCertCfg)
	if err != nil {
		fmt.Printf("marshal alarm cert config failed: %v\n", err)
		return
	}
	res := getConfig.getAlarmCertConfig()
	convey.So(res, convey.ShouldResemble, string(bytes))
	res = getConfig.getAlarmCertConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)
	res = getConfig.getAlarmCertConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)
	res = getConfig.getAlarmCertConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)

	var p2 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p2.Reset()
	res = getConfig.getAlarmCertConfig()
	convey.So(res, convey.ShouldResemble, constants.Failed)
}

func testGetEdgeOmCaps() {
	testCaps := []string{"npu_sharing_config"}
	var p1 = gomonkey.ApplyFuncReturn(config.GetCapabilityMgr, &config.CapabilityMgr{}).
		ApplyMethodReturn(&config.CapabilityMgr{}, "GetCaps", testCaps)
	defer p1.Reset()

	getEdgeOmCapsMsg := getConfigMsg
	err := getEdgeOmCapsMsg.FillContent(constants.EdgeOmCapabilities)
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("getEdgeOmCaps success", func() {
		err = getConfig.Handle(getEdgeOmCapsMsg)
		convey.So(err, convey.ShouldBeNil)

		res := getConfig.getEdgeOmCaps()
		convey.So(res, convey.ShouldResemble, `{"product_capability_edge":["npu_sharing_config"]}`)
	})

	convey.Convey("getEdgeOmCaps failed", func() {
		p2 := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, test.ErrTest)
		defer p2.Reset()
		res := getConfig.getEdgeOmCaps()
		convey.So(res, convey.ShouldResemble, constants.Failed)
	})
}

func testErrConfigMsg() {
	errConfigMsg := getConfigMsg
	err := errConfigMsg.FillContent("error content")
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(errConfigMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testErrInitDbMgr() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetCompConfigDir, "", testErr)
	defer p1.Reset()

	_, err := getConfig.initDbMgr()
	convey.So(err, convey.ShouldResemble, errors.New("get config dir failed"))

	err = getConfig.Handle(getPodConfigMsg)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getNetMgrConfigMsg)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getInstallConfigMsg)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getSoftwareCertMsg)
	convey.So(err, convey.ShouldBeNil)
	err = getConfig.Handle(getAlarmConfigMsg)
	convey.So(err, convey.ShouldBeNil)
}

func newGetConfigMsg() (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return nil, errors.New("new message failed")
	}
	msg.SetRouter(constants.InnerClient, constants.ConfigMgr, constants.OptGet, constants.ResConfig)
	return msg, nil
}
