// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"bytes"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindxedge/base/common"
)

func TestDiskPressureProtection(t *testing.T) {
	convey.Convey("test disk pressure protection", t, func() {
		type testcase struct {
			stat   syscall.Statfs_t
			result bool
		}
		runner := func(tc testcase) {
			patch := gomonkey.ApplyFunc(syscall.Statfs, func(path string, statfs *syscall.Statfs_t) error {
				*statfs = tc.stat
				return nil
			})
			defer patch.Reset()

			var buffer bytes.Buffer
			protectedWriter := WithDiskPressureProtect(&buffer, "")
			_, err := protectedWriter.Write([]byte("hello"))
			assertion := convey.ShouldNotBeNil
			if tc.result {
				assertion = convey.ShouldBeNil
			}
			convey.So(err, assertion)
		}

		const (
			size10  = 10
			size20  = 20
			size21  = 21
			size100 = 100
		)
		testcases := []testcase{
			{stat: syscall.Statfs_t{Bsize: size20 * common.MB, Blocks: size100, Bavail: size20}},
			{stat: syscall.Statfs_t{Bsize: size20 * common.MB, Blocks: size100, Bavail: size21}, result: true},
			{stat: syscall.Statfs_t{Bsize: size10 * common.MB, Blocks: size20, Bavail: size20}},
		}
		for _, tc := range testcases {
			runner(tc)
		}
	})
}
