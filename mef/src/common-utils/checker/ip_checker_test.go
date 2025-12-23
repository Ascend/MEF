// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestIpv4Checker(t *testing.T) {
	checker := GetIpV4Checker("", true)
	convey.Convey("test normal case", t, func() {
		ip := "51.38.66.39"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})
	convey.Convey("test none-ip", t, func() {
		ip := "256.257.258.0"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP parse failed")
	})
	convey.Convey("test ipv6", t, func() {
		ip := "fe80::6f7b:10d5:3cbd:5c3d"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP is not a valid IPv4 address")
	})
	convey.Convey("test multicast", t, func() {
		ip := "224.0.0.0"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP can't be a multicast address")
	})
	convey.Convey("test link local unicast", t, func() {
		ip := "169.254.0.0"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP can't be a link-local unicast address")
	})
	convey.Convey("test broadcast ip", t, func() {
		ip := "255.255.255.255"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP can't be a broadcast address")
	})
	convey.Convey("test zero ip", t, func() {
		ip := "0.0.0.0"
		checkResult := checker.Check(ip)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "IP can't be an all zeros address")
	})
}
