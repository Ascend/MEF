// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package util test for get_serial_number_a500.go
package util

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"
)

func TestGetSerialNumber(t *testing.T) {
	convey.Convey("test func GetSerialNumber success, get a500 sn success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(getA500Sn, "", nil)
		defer p1.Reset()
		_, err := GetSerialNumber("")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func GetSerialNumber success, get a500 sn failed, get uuid success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(getA500Sn, "", test.ErrTest).
			ApplyFuncReturn(envutils.RunCommand, "", nil)
		defer p1.Reset()
		_, err := GetSerialNumber("")
		convey.So(err, convey.ShouldBeNil)
	})
}
