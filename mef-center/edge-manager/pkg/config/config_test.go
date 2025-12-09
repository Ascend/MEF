// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package config

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestPodConfig(t *testing.T) {
	convey.Convey("test CheckAndModifyHostPath", t, testCheckAndModifyHostPath)
	convey.Convey("test testCheckAndModifyMaxLimitNumber", t, testCheckAndModifyMaxLimitNumber)

}

func testCheckAndModifyHostPath() {
	testPaths := []string{
		"/tmp",
		"./errorPath1",
		"/../errorPath2",
	}
	path := CheckAndModifyHostPath(testPaths)
	convey.So(path, convey.ShouldResemble, []string{"/tmp"})
}

func testCheckAndModifyMaxLimitNumber() {
	const defaultMaxLimitNumber = 20
	errorNumber := int64(0)
	number := CheckAndModifyMaxLimitNumber(errorNumber)
	convey.So(number, convey.ShouldEqual, defaultMaxLimitNumber)
}

func TestAuthConfig(t *testing.T) {
	convey.Convey("test auth config", t, testCheckAuthConfig)
}

func testCheckAuthConfig() {
	testConfig := AuthInfo{TokenExpireTime: minTokenExpireTime}
	SetConfig(testConfig)
	config := GetAuthConfig()
	convey.So(config, convey.ShouldResemble, testConfig)

	errorConfig := AuthInfo{TokenExpireTime: 0}
	err := CheckAuthConfig(errorConfig)
	convey.So(err, convey.ShouldNotBeNil)
}
