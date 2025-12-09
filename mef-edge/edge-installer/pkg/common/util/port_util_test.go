// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"
)

func TestGetUsedPorts(t *testing.T) {
	protocol := v1.ProtocolTCP
	convey.Convey("Given a protocol", t, func() {
		usedPorts, err := GetUsedPorts(protocol)
		convey.So(err, convey.ShouldBeNil)
		convey.So(usedPorts.Len(), convey.ShouldBeGreaterThan, 0)
	})

	convey.Convey("Given an invalid protocol", t, func() {
		errProtocol := v1.Protocol("invalid")
		usedPorts, err := GetUsedPorts(errProtocol)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(usedPorts.Len(), convey.ShouldEqual, 0)
	})

	convey.Convey("test func GetUsedPorts failed, load protocol stat file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p1.Reset()
		usedPorts, err := GetUsedPorts(protocol)
		convey.So(usedPorts.Len(), convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
