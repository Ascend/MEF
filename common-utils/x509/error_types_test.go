//  Copyright(c) 2024. Huawei Technologies Co.,Ltd.  All rights reserved.

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
