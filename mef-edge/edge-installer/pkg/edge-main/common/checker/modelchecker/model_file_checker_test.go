// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package modelchecker

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/types"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestCheckModelFileMsg(t *testing.T) {
	convey.Convey("test check update model file msg", t, testUpdateModelFile)
	convey.Convey("test check delete model file all msg", t, testDeleteAllModelFile)
	convey.Convey("test check delete model file temp msg", t, testDeleteTempModelFile)
	convey.Convey("test check delete model file empty msg", t, testDeleteEmptyModelFile)
}

func createUpdateTemplate() types.ModelFileInfo {
	fileServer := types.FileServerInfo{
		Protocol: "https",
		Path:     "GET https://FDAddr/models",
		UserName: "huawei",
		PassWord: "12345678",
	}
	modelFile := types.ModelFile{
		Name:       "module.om",
		Version:    "1.0",
		CheckType:  "sha256",
		CheckCode:  "7e642fa557533508b589d7fcaef922b443ff4b5df16b2ad91581bd2b3d7c1fc3",
		Size:       "1024",
		FileServer: fileServer,
	}
	info := types.ModelFileInfo{
		Operation:  "update",
		Target:     "all",
		Uuid:       "bdf3242b-aec1-4100-af91-afa2b8fde88a",
		ModelFiles: []types.ModelFile{modelFile},
	}
	return info
}

func createDeleteTemplate() types.ModelFileInfo {
	fileServer := types.FileServerInfo{
		Protocol: "https",
		Path:     "GET https://FDAddr:port/models",
		UserName: "huawei",
		PassWord: "12345678",
	}
	modelFile := types.ModelFile{
		Name:       "module.om",
		Version:    "1.0",
		CheckType:  "sha256",
		CheckCode:  "7e642fa557533508b589d7fcaef922b443ff4b5df16b2ad91581bd2b3d7c1fc3",
		Size:       "1024",
		FileServer: fileServer,
	}
	info := types.ModelFileInfo{
		Operation:  "delete",
		Target:     "all",
		Uuid:       "bdf3242b-aec1-4100-af91-afa2b8fde88a",
		ModelFiles: []types.ModelFile{modelFile},
	}
	return info
}

func createDeleteAllFailCases() []downloadCase {
	var deleteFailCases []downloadCase

	info1 := createDeleteTemplate()
	info1.ModelFiles = []types.ModelFile{}
	info1.Uuid = "1bdf3242b-aec1-4100-af91-afa2b8fde88a"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info1, expect: "regex checker Check [Uuid] failed"})

	return deleteFailCases
}

func createDeleteTempFailCases() []downloadCase {
	var deleteFailCases []downloadCase

	info1 := createDeleteTemplate()
	info1.Target = "temp"
	info1.ModelFiles = []types.ModelFile{}
	info1.Uuid = "1bdf3242b-aec1-4100-af91-afa2b8fde88a"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info1, expect: "regex checker Check [Uuid] failed"})

	info2 := createDeleteTemplate()
	info2.Target = "temp"
	info2.ModelFiles[0].Name = "module.jpg"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info2, expect: "name suffix not valid"})

	info3 := createDeleteTemplate()
	info3.Target = "temp"
	info3.ModelFiles[0].Version = "_1.0"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info3, expect: "regex checker Check [Version] failed"})

	info4 := createDeleteTemplate()
	info4.Target = "temp"
	info4.ModelFiles[0].Name = "model..tar.gz"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info4, expect: "string excludeWords words checker Check [Name] failed"})
	return deleteFailCases
}

func createDeleteEmptyFailCases() []downloadCase {
	var deleteFailCases []downloadCase
	info1 := createDeleteTemplate()
	info1.Target = ""
	info1.ModelFiles = []types.ModelFile{}
	info1.Uuid = "1bdf3242b-aec1-4100-af91-afa2b8fde88a"
	deleteFailCases = append(deleteFailCases,
		downloadCase{modelFileInfo: info1, expect: "regex checker Check [Uuid] failed"})
	return deleteFailCases
}

func createUpdateFailCases1() []downloadCase {
	var downloadTestCases []downloadCase
	info1 := createUpdateTemplate()
	info1.Operation = "kill"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info1, expect: "no check func for model file"})
	info2 := createUpdateTemplate()
	info2.Target = ""
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info2, expect: "the value not in"})
	info3 := createUpdateTemplate()
	info3.Uuid = "bdf3242b-aec1-4100-af91-afa2b8fde88a1"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info3, expect: "the string value not match requirement"})
	info4 := createUpdateTemplate()
	info4.ModelFiles = []types.ModelFile{}
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info4, expect: "list checker Check len"})
	info5 := createUpdateTemplate()
	info5.ModelFiles[0].Name = "module.jpg"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info5, expect: "name suffix not valid"})
	info6 := createUpdateTemplate()
	info6.ModelFiles[0].Version = "_10"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info6, expect: "the string value not match requirement"})
	info7 := createUpdateTemplate()
	info7.ModelFiles[0].CheckType = "MD5"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info7, expect: "string choice checker Check [CheckType] failed: the value not in"})
	return downloadTestCases
}

func createUpdateFailCases2() []downloadCase {
	var downloadTestCases []downloadCase
	info8 := createUpdateTemplate()
	info8.ModelFiles[0].Size = "0"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info8, expect: "model file size is invalid"})
	info9 := createUpdateTemplate()
	info9.ModelFiles[0].Size = "abc"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info9, expect: "model size check failed"})
	info10 := createUpdateTemplate()
	info10.ModelFiles[0].FileServer.Protocol = "http"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info10, expect: "string choice checker Check [Protocol] failed"})
	info11 := createUpdateTemplate()
	info11.ModelFiles[0].FileServer.UserName = ""
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info11, expect: "regex checker Check [UserName] failed"})
	info12 := createUpdateTemplate()
	info12.ModelFiles[0].FileServer.PassWord = ""
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info12, expect: "regex checker Check [PassWord] failed"})
	info13 := createUpdateTemplate()
	info13.ModelFiles[0].FileServer.Path = "GET http://www.huawei.com"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info13, expect: "not https url"})
	info14 := createUpdateTemplate()
	info14.ModelFiles[0].FileServer.Path = "GET https://www.$huawei.com"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info14, expect: "contain invalid char"})
	info15 := createUpdateTemplate()
	info15.ModelFiles[0].Size = "4294967297"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info15, expect: "model file size is invalid"})
	info16 := createUpdateTemplate()
	info16.ModelFiles[0].CheckCode = "7e642fa557533508b589d7fcaef922b443ff4b5df16b2ad91581bd2b3d7c1fc34sdf"
	downloadTestCases = append(downloadTestCases,
		downloadCase{modelFileInfo: info16, expect: "regex checker Check [CheckCode] failed"})
	return downloadTestCases
}

type downloadCase struct {
	modelFileInfo types.ModelFileInfo
	expect        string
}

func testUpdateModelFile() {
	patch := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", nil, nil)
	defer patch.Reset()

	info1 := createUpdateTemplate()
	content1, err := json.Marshal(info1)
	convey.So(err, convey.ShouldBeNil)
	err = CheckModelFileMsg(content1)
	convey.So(err, convey.ShouldBeNil)

	var updateTestCases []downloadCase
	updateTestCases1 := createUpdateFailCases1()
	updateTestCases = append(updateTestCases, updateTestCases1...)
	updateTestCases2 := createUpdateFailCases2()
	updateTestCases = append(updateTestCases, updateTestCases2...)
	for _, failCase := range updateTestCases {
		content, err := json.Marshal(failCase.modelFileInfo)
		convey.So(err, convey.ShouldBeNil)
		err = CheckModelFileMsg(content)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, failCase.expect)
	}
}

func testDeleteAllModelFile() {
	deleteFailCases := createDeleteAllFailCases()
	for _, failCase := range deleteFailCases {
		content, err := json.Marshal(failCase.modelFileInfo)
		convey.So(err, convey.ShouldBeNil)
		err = CheckModelFileMsg(content)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, failCase.expect)
	}
}

func testDeleteTempModelFile() {
	deleteFailCases := createDeleteTempFailCases()
	for _, failCase := range deleteFailCases {
		content, err := json.Marshal(failCase.modelFileInfo)
		convey.So(err, convey.ShouldBeNil)
		err = CheckModelFileMsg(content)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, failCase.expect)
	}
}

func testDeleteEmptyModelFile() {
	deleteFailCases := createDeleteEmptyFailCases()
	for _, failCase := range deleteFailCases {
		content, err := json.Marshal(failCase.modelFileInfo)
		convey.So(err, convey.ShouldBeNil)
		err = CheckModelFileMsg(content)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, failCase.expect)
	}
}
