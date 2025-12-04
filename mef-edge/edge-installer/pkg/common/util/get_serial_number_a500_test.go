// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package util test for get_serial_number_a500.go
package util

import (
	"testing"

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
