// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 error_types  test file
package x509

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestError(t *testing.T) {
	testErr1 := ErrCertParseFailed
	testErr2 := ErrCrlExpired
	convey.Convey("test error info print", t, func() {
		convey.So(testErr1.Error(), convey.ShouldResemble,
			"x509 error. code: [1001], reason: [parse x509 certificate failed]")
		convey.So(testErr2.Error(), convey.ShouldNotResemble,
			"x509 error. code: [1001], reason: [parse x509 certificate failed]")
	})
	convey.Convey("test error comparison", t, func() {
		convey.So(testErr1 == ErrCertParseFailed, convey.ShouldEqual, true)
		convey.So(testErr1 == testErr2, convey.ShouldEqual, false)
	})
}
