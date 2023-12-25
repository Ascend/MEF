// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configmanager for inner message test
package configmanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetAllNodeInfo(t *testing.T) {
	convey.Convey("get all node info functional test", t, func() {
		convey.Convey("get all node info success", func() {
			_, err := getAllNodeInfo()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
