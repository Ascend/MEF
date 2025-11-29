// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common test for client info
package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/constants"
)

func TestSetClientAddr(t *testing.T) {
	convey.Convey("set fd ip should be success", t, func() {
		err := SetFdIp(constants.ModDeviceOm, "127.0.0.1:10000")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set fd ip should be failed, addr error", t, func() {
		err := SetFdIp(constants.ModDeviceOm, "127.0.0.1")
		convey.So(err, convey.ShouldResemble, errors.New("invalid addr: 127.0.0.1"))
	})

	convey.Convey("set fd ip should be failed, mod name error", t, func() {
		err := SetFdIp(constants.ModEdgeOm, "127.0.0.1:10000")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("set fd ip failed. "+
			"invalid module name: %v", constants.ModEdgeOm))
	})
}

func TestGetClientAddr(t *testing.T) {
	convey.Convey("get fd ip should be success", t, func() {
		fdIp = "127.0.0.1"
		ip, err := GetFdIp()
		convey.So(ip, convey.ShouldEqual, "127.0.0.1")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("get fd ip should be failed, ip is nil", t, func() {
		fdIp = ""
		ip, err := GetFdIp()
		convey.So(ip, convey.ShouldEqual, "")
		convey.So(err, convey.ShouldResemble, errors.New("fd ip not found"))
	})
}
