// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr test for save cert handler
package handlermgr

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/imageconfig"
)

var (
	saveCertUpdateMsg *model.Message
	saveCertDeleteMsg *model.Message
	saveCertSdk       = saveCertHandlerSdk{}
)

func TestSaveCertHandler(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", nil).
		ApplyFuncReturn(fileutils.MakeSureDir, nil).
		ApplyFuncReturn(fileutils.WriteData, nil).
		ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(fileutils.IsSoftLink, nil).
		ApplyFuncReturn(fileutils.CopyFile, nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyMethodReturn(imageconfig.ImageCfgFlow{}, "RunTasks", nil).
		ApplyFuncReturn(fileutils.DeleteFile, nil).
		ApplyFuncReturn(util.DeleteImageCertFile, nil)
	defer p.Reset()

	var err error
	saveCertUpdateMsg, err = newSaveCertUpdateMsg()
	if err != nil {
		fmt.Printf("new save cert msg failed, error: %v\n", err)
		return
	}
	saveCertDeleteMsg, err = newSaveCertDeleteMsg()
	if err != nil {
		fmt.Printf("new save cert msg failed, error: %v\n", err)
		return
	}
	convey.Convey("save cert should be success", t, testSaveCertSdkHandler)
	convey.Convey("save cert should be failed, param convert failed", t, testSaveCertSdkHandlerErrParamConv)
	convey.Convey("save cert should be failed, param convert failed", t, testSaveCertSdkHandlerErrOpt)
	convey.Convey("update cert should be failed, get path error", t, testUpdateErrGetPath)
	convey.Convey("update cert should be failed, make sure error", t, testUpdateErrMakeSure)
	convey.Convey("update cert should be failed, write data error", t, testUpdateErrWriteData)
	convey.Convey("update cert should be failed, copy error", t, testUpdateErrCopy)
	convey.Convey("update cert should be failed, image config error", t, testUpdateErrImageConfig)
	convey.Convey("delete cert should be failed, get path error", t, testDeleteErrGetPath)
	convey.Convey("delete cert should be failed, delete ca cert error", t, testDeleteErrDeleteCaCert)
	convey.Convey("delete cert should be failed, delete image cert error", t, testDeleteErrDeleteImageCert)
	convey.Convey("check cert info should be failed, name or opt error", t, testCheckErrNameAndOpt)
	convey.Convey("test func checkCaContent should be failed, content error", t, testCheckCaContent)
	convey.Convey("test func checkImageAddress should be failed, port error", t, testCheckImageAddressErrPort)
	convey.Convey("test func checkImageAddress should be failed, ip error", t, testCheckImageAddressErrIp)
}

func testSaveCertSdkHandler() {
	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)

	err = saveCertSdk.Parse(saveCertDeleteMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertDeleteMsg)
	convey.So(err, convey.ShouldBeNil)

	saveCertSdk.PrintOpLogOk()
	saveCertSdk.PrintOpLogFail()
}

func testSaveCertSdkHandlerErrParamConv() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	err = msg.FillContent("error content")
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Parse(msg)
	convey.So(err, convey.ShouldNotBeNil)
}

func testSaveCertSdkHandlerErrOpt() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	certInfo := &util.ClientCertResp{
		CertOpt: "error opt",
	}
	err = msg.FillContent(certInfo, true)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Parse(msg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("cert operation not support"))
}

func testUpdateErrGetPath() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", testErr}},
		{Values: gomonkey.Params{"", nil}},
		{Values: gomonkey.Params{"", testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(path.GetCompSpecificDir, outputs)
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldResemble, errors.New("get cert dir failed"))
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get docker certs dir failed, error: %v", testErr))
}

func testUpdateErrMakeSure() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}},
		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(fileutils.MakeSureDir, outputs)
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldResemble, errors.New("create ca folder failed"))
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create docker certs dir failed, error: %v", testErr))
}

func testUpdateErrWriteData() {
	var p1 = gomonkey.ApplyFunc(fileutils.WriteData,
		func(filePath string, fileData []byte, checkerParam ...fileutils.FileChecker) error {
			return testErr
		})
	defer p1.Reset()
	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldResemble, errors.New("update ca file failed"))
}

func testUpdateErrCopy() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, testErr)
	defer p1.Reset()
	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldNotBeNil)
}

func testUpdateErrImageConfig() {
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(imageconfig.ImageCfgFlow{}), "RunTasks",
		func(imageconfig.ImageCfgFlow) error {
			return testErr
		})
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertUpdateMsg)
	convey.So(err, convey.ShouldNotBeNil)
}

func testDeleteErrGetPath() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", testErr)
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertDeleteMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertDeleteMsg)
	convey.So(err, convey.ShouldResemble, errors.New("get cert dir failed"))
}

func testDeleteErrDeleteCaCert() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, testErr)
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertDeleteMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertDeleteMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("delete %s ca file failed", saveCertSdk.res.CertName))
}

func testDeleteErrDeleteImageCert() {
	var p1 = gomonkey.ApplyFuncReturn(util.DeleteImageCertFile, testErr)
	defer p1.Reset()

	err := saveCertSdk.Parse(saveCertDeleteMsg)
	convey.So(err, convey.ShouldBeNil)
	err = saveCertSdk.Handle(saveCertDeleteMsg)
	convey.So(err, convey.ShouldResemble, errors.New("delete image cert failed"))
}

func testCheckErrNameAndOpt() {
	validResp := util.ClientCertResp{
		CertName:     constants.ImageCertName,
		CertContent:  testCaContent,
		CertOpt:      constants.OptUpdate,
		ImageAddress: "",
	}
	saveCertSdk.res = validResp
	err := saveCertSdk.Check(saveCertUpdateMsg)
	convey.So(err, convey.ShouldBeNil)

	invalidResps := []util.ClientCertResp{
		{
			"error cert name",
			testCaContent,
			"",
			constants.OptUpdate,
			"",
		},
		{
			constants.ImageCertName,
			testCaContent,
			"",
			"error operation",
			"",
		},
	}

	for _, invalidResp := range invalidResps {
		saveCertSdk.res = invalidResp
		err = saveCertSdk.Check(saveCertUpdateMsg)
		convey.So(err, convey.ShouldResemble, errors.New("check cert info failed"))
	}
}

func testCheckCaContent() {
	bigBytes := []byte("test-data ")
	for {
		bigBytes = append(bigBytes, bigBytes...)
		if len(bigBytes) > constants.CertSizeLimited {
			break
		}
	}
	err := checkCaContent(string(bigBytes))
	convey.So(err, convey.ShouldResemble, errors.New("verify ca file size failed"))

	err = checkCaContent(testWrongCaContent)
	convey.So(err, convey.ShouldResemble, errors.New("verify ca file failed"))
}

func testCheckImageAddressErrPort() {
	err := checkImageAddress("127.0.0.1:0")
	convey.So(err, convey.ShouldNotBeNil)

	var p1 = gomonkey.ApplyFunc(strconv.Atoi,
		func(s string) (int, error) {
			return 0, testErr
		})
	defer p1.Reset()
	err = checkImageAddress("127.0.0.1:0")
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckImageAddressErrIp() {
	err := checkImageAddress("ab:10000")
	convey.So(err, convey.ShouldNotBeNil)
	err = checkImageAddress(".a.bc:10000")
	convey.So(err, convey.ShouldNotBeNil)
	err = checkImageAddress("a.bc-:10000")
	convey.So(err, convey.ShouldNotBeNil)
	err = checkImageAddress("a.bc~:10000")
	convey.So(err, convey.ShouldNotBeNil)
	err = checkImageAddress("127.0.0.1:10000")
	convey.So(err, convey.ShouldNotBeNil)

	ip := net.ParseIP("127.0.0.10")
	var p1 = gomonkey.ApplyFuncReturn(net.ParseIP, ip).
		ApplyMethodReturn(net.IP{}, "IsMulticast", true)
	defer p1.Reset()
	err = checkImageAddress("127:10000")
	convey.So(err, convey.ShouldNotBeNil)

	var p2 = gomonkey.ApplyFunc(net.ParseIP,
		func(s string) net.IP {
			return nil
		})
	defer p2.Reset()
	err = checkImageAddress("127.0.0.1:10000")
	convey.So(err, convey.ShouldBeNil)
}

func newSaveCertUpdateMsg() (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return nil, errors.New("new message failed")
	}

	certInfo := &util.ClientCertResp{
		CertName:     constants.ImageCertName,
		CertContent:  testCaContent,
		CrlContent:   testUnmatchedCrlContent,
		CertOpt:      constants.OptUpdate,
		ImageAddress: "test update image addr",
	}
	err = msg.FillContent(certInfo, true)
	if err != nil {
		return nil, errors.New("fill content failed")
	}
	return msg, nil
}

func newSaveCertDeleteMsg() (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return nil, errors.New("new message failed")
	}

	certInfo := &util.ClientCertResp{
		CertName:     constants.ImageCertName,
		CertContent:  testCaContent,
		CertOpt:      constants.OptDelete,
		ImageAddress: "test delete image addr",
	}
	err = msg.FillContent(certInfo, true)
	if err != nil {
		fmt.Printf("fill content failed: %v", err)
		return nil, errors.New("fill content failed")
	}
	return msg, nil
}
