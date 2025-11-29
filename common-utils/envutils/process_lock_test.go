// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package envutils
package envutils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFlock(t *testing.T) {
	convey.Convey("test flock", t, func() {
		const lockName = "my_lock"
		err := GetFlock(lockName).Lock("reason one")
		convey.So(err, convey.ShouldBeNil)
		err = GetFlock(lockName).Lock("reason two")
		convey.So(err, convey.ShouldNotBeNil)
		GetFlock(lockName).Unlock()
	})
}
