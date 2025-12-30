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

// Package util test for env_sdk.go
package util

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

func TestDeleteImageCertFile(t *testing.T) {
	convey.Convey("test func DeleteImageCertFile success, safe deletion success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, nil)
		defer p1.Reset()
		err := DeleteImageCertFile("./")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func DeleteImageCertFile, safe deletion failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		err := DeleteImageCertFile("./")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
