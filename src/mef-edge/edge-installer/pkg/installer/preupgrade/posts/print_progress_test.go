// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
