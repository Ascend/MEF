// Copyright (c) 2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package downloadmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/util"
)

var downloadInfoJson = `{"softwareName":"MEFEdge","downloadInfo":{` +
	`"package":"GET https://xxx.xxx/xxx/?contentType=MEFEdge\u0026version=1.0` +
	`\u0026fileName=Ascend-mindxedge-mefedgesdk_5.0.RC3_linux-aarch64.tar.gz",` +
	`"signFile":"GET https://xxx.xxx/xxx/?contentType=MEFEdge\u0026version=1.0` +
	`\u0026fileName=Ascend-mindxedge-mefedgesdk_5.0.RC3_linux-aarch64.tar.gz.cms",` +
	`"crlFile":"GET https://xxx.xxx/xxx/?contentType=MEFEdge\u0026version=1.0` +
	`\u0026fileName=Ascend-mindxedge-mefedgesdk_5.0.RC3_linux-aarch64.tar.gz.crl",` +
	`"username":"testAccount","password":"dGVzdEZvclV0"}}`

func TestProcessDownloadSoftware(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{}, nil)
	defer p.Reset()
	convey.Convey("process method success case", t, testProcessDownloadSfwWith)
	convey.Convey("param decode error case", t, testProcessDownloadSfwWithErrorMsg)
}

func testProcessDownloadSfwWith() {
	respOk, respCert := &model.Message{}, &model.Message{}
	err := respOk.FillContent("OK")
	convey.So(err, convey.ShouldBeNil)
	err = respCert.FillContent(`{"CertName":"","CertContent":null,"ErrorMsg":""}`)
	convey.So(err, convey.ShouldBeNil)
	cells := []gomonkey.OutputCell{
		{Values: gomonkey.Params{respCert, nil}, Times: 1},
		{Values: gomonkey.Params{respOk, nil}, Times: 2},
	}
	fakeDp := downloadProcess{}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFuncSeq(modulemgr.SendSyncMessage, cells),
		gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Getuid()), nil),
		gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", nil),
		gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil),
		gomonkey.ApplyPrivateMethod(&fakeDp, "downloadSoftware", func() error { return nil }),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	d := downloadMgr{}
	msg := model.Message{}
	err = msg.FillContent(downloadInfoJson)
	convey.So(err, convey.ShouldBeNil)
	processErr := d.processDownloadSoftware(msg)
	convey.So(processErr, convey.ShouldBeNil)
}

func testProcessDownloadSfwWithErrorMsg() {
	d := downloadMgr{}
	msg := model.Message{}
	err := msg.FillContent("")
	convey.So(err, convey.ShouldBeNil)
	processErr := d.processDownloadSoftware(msg)
	convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "get download process failed")
}

func TestCheckDownloadInfo(t *testing.T) {
	convey.Convey("invalid download info case", t, func() {
		dp := downloadProcess{}
		err := json.Unmarshal(model.RawMessage(downloadInfoJson), &dp.sfwDownloadInfo)
		assertions.ShouldBeNil(err)

		dp.sfwDownloadInfo.DownloadInfo.CrlFile = ""
		processErr := dp.checkDownloadInfo()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "check software download para failed")
	})
}

func TestPrepareDownloadDir(t *testing.T) {
	convey.Convey("prepare download dir case", t, func() {
		dp := downloadProcess{}
		p1 := gomonkey.ApplyFuncReturn(util.NewInnerMsgWithFullParas, nil, errors.New("create message error"))
		processErr := dp.prepareDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "create message error")
		p1.Reset()

		p2 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, nil, errors.New("receive resp timeout"))
		processErr = dp.prepareDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring,
			"receive resp timeout")
		p2.Reset()

		fakeResp := &model.Message{}
		p4 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, fakeResp, nil)

		err := fakeResp.FillContent("prepare dir error")
		convey.So(err, convey.ShouldBeNil)
		processErr = dp.prepareDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring,
			"prepare dir error")
		p4.Reset()
	})
}

func TestCheckDownloadDir(t *testing.T) {
	convey.Convey("check download dir case", t, func() {
		dp := downloadProcess{}
		p1 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), errors.New("lookup user failed"))
		processErr := dp.checkDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "lookup user failed")
		p1.Reset()

		p2 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Getuid()), nil).
			ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", errors.New("symlinks owner may not root"))
		processErr = dp.checkDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "symlinks owner may not root")
		p2.Reset()

		p3 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Getuid()), nil).
			ApplyFuncReturn(fileutils.CheckOriginPath, "", errors.New("can't support symlinks"))
		p4 := gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", nil)
		processErr = dp.checkDownloadDir()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "can't support symlinks")
		p3.Reset()
		p4.Reset()
	})
}

func TestGetSoftwareCert(t *testing.T) {
	convey.Convey("get software cert case", t, func() {
		dp := downloadProcess{}
		p1 := gomonkey.ApplyFuncReturn(util.NewInnerMsgWithFullParas, nil, errors.New("create message error"))
		processErr := dp.getSoftwareCert()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "create message error")
		p1.Reset()

		p2 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, nil, errors.New("receive resp timeout"))
		processErr = dp.getSoftwareCert()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "receive resp timeout")
		p2.Reset()

		fakeResp := &model.Message{}
		p3 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, fakeResp, nil)

		err := fakeResp.FillContent(`{"CertName":"","CertContent":null,"ErrorMsg":"get cert error"}`)
		convey.So(err, convey.ShouldBeNil)
		processErr = dp.getSoftwareCert()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "get cert error")
		p3.Reset()
	})
}

func TestVerifyAndUnpack(t *testing.T) {
	convey.Convey("verify and unpack", t, func() {
		dp := downloadProcess{}
		p1 := gomonkey.ApplyFuncReturn(util.NewInnerMsgWithFullParas, nil, errors.New("create message error"))
		processErr := dp.verifyAndUnpack()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "create message error")
		p1.Reset()

		p2 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, nil, errors.New("receive resp timeout"))
		processErr = dp.verifyAndUnpack()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "receive resp timeout")
		p2.Reset()

		fakeResp := &model.Message{}
		p3 := gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, fakeResp, nil)

		err := fakeResp.FillContent("verify error")
		convey.So(err, convey.ShouldBeNil)
		processErr = dp.verifyAndUnpack()
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring, "verify error")
		p3.Reset()
	})
}
