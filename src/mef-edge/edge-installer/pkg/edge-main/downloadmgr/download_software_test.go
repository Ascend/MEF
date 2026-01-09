// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package downloadmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/util"
)

func TestDownloadSoftware(t *testing.T) {
	dp := downloadProcess{}
	err := json.Unmarshal(model.RawMessage(downloadInfoJson), &dp.sfwDownloadInfo)
	assertions.ShouldBeNil(err)

	convey.Convey("success case", t, func() {
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFuncReturn(envutils.CheckDiskSpace, nil),
			gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil),
			gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil),
			gomonkey.ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", nil),
			gomonkey.ApplyPrivateMethod(&httpsmgr.HttpsRequest{}, "GetRespToFileWithLimit",
				func(writer io.Writer, limit int64) error {
					return nil
				}),
		}
		defer func() {
			for _, patch := range patches {
				patch.Reset()
			}
		}()
		processErr := dp.downloadSoftware()
		convey.So(processErr, convey.ShouldBeNil)
	})
}

func TestParseUrlInfo(t *testing.T) {
	convey.Convey("nil url case", t, func() {
		_, err := parseUrlInfo("")
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "url is nil")
	})
	convey.Convey("error url segments case", t, func() {
		_, err := parseUrlInfo(`https://xxx.xxx/xxx/"`)
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "url filed num invalid")
	})
	convey.Convey("error url method case", t, func() {
		_, err := parseUrlInfo(`PATCH https://xxx.xxx/xxx/"`)
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "url method invalid")
	})

}

func TestGetValidUrls(t *testing.T) {
	var sfwDownloadInfo util.SoftwareDownloadInfo
	err := json.Unmarshal(model.RawMessage(downloadInfoJson), &sfwDownloadInfo)
	assertions.ShouldBeNil(err)

	convey.Convey("invalid package url", t, func() {
		temp := sfwDownloadInfo.DownloadInfo.Package
		sfwDownloadInfo.DownloadInfo.Package = ""
		_, processErr := getValidUrls(sfwDownloadInfo)
		convey.So(fmt.Sprintf("%v", processErr), convey.ShouldContainSubstring,
			"package url is invalid")
		sfwDownloadInfo.DownloadInfo.Package = temp
	})

}

func TestCreateHttpsReqAndSaveToFile(t *testing.T) {
	convey.Convey("nil url case", t, func() {
		params := downloadParams{}
		params.savePath = ".."
		err := createHttpsReqAndSaveToFile(params)
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "path is no a file")
	})
	convey.Convey("error url segments case", t, errorUrlSegmentCase)
}

func errorUrlSegmentCase() {
	params := downloadParams{}
	fakeFile := &os.File{}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFuncReturn(fileutils.CheckOriginPath, "", nil),
		gomonkey.ApplyFuncReturn(os.OpenFile, fakeFile, errors.New("open file failed")),
		gomonkey.ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", nil),
		gomonkey.ApplyPrivateMethod(fakeFile, "Close", func() error { return nil }),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	err := createHttpsReqAndSaveToFile(params)
	convey.So(err, convey.ShouldNotBeNil)
}

func TestGetTargetFilePath(t *testing.T) {
	convey.Convey("invalid package type case", t, func() {
		_, err := getTargetFilePath("MEFEdge", "invalidFileType")
		convey.So(fmt.Sprintf("%v", err), convey.ShouldContainSubstring, "invalid package type")
	})
}
