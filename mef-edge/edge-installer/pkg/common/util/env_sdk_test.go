// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
