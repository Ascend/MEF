// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package posts for testing print process
package posts

import (
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/installer/common"
)

func TestPrintProgress(t *testing.T) {
	convey.Convey("test print failed process", t, func() {
		testSuccessItem := &common.FlowItem{
			Description: "upgrading",
			Progress:    common.ProgressSuccess,
		}
		err := PrintProgress(testSuccessItem)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test print failed process", t, func() {
		testFailedItem := &common.FlowItem{
			Description: "check package and environment",
			Progress:    40,
			Error:       errors.New("verify package failed"),
		}
		err := PrintProgress(testFailedItem)
		convey.So(err, convey.ShouldBeNil)
	})
}
