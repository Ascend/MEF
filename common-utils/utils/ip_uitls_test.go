// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils offer the some utils for certificate handling
package utils

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	localhost     = "127.0.0.1"
	localhostLoop = "0.0.0.0"
)

func TestClientIP(t *testing.T) {
	convey.Convey("test ClientIP func", t, func() {
		convey.Convey("get IP from X-Forwarded-For", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {localhost, localhostLoop}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from X-Real-Ip", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {},
				"X-Real-Ip": {localhost}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from RemoteAddr", func() {
			ip := ClientIP(mockRequest(map[string][]string{"X-Forwarded-For": {},
				"X-Real-Ip": {}}))
			convey.So(ip, convey.ShouldEqual, localhost)
		})
		convey.Convey("get IP from RemoteAddr failed", func() {
			ip := ClientIP(&http.Request{RemoteAddr: localhost})
			convey.So(ip, convey.ShouldEqual, "")
		})
		convey.Convey("get IP failed", func() {
			ip := ClientIP(&http.Request{})
			convey.So(ip, convey.ShouldEqual, "")
		})
	})
}

func mockRequest(header map[string][]string) *http.Request {
	return &http.Request{
		Method:        "GET",
		URL:           nil,
		Proto:         "HTTP",
		ProtoMajor:    0,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: 0,
		Close:         false,
		Host:          "www.test.com",
		RemoteAddr:    "127.0.0.1:8080",
	}
}
