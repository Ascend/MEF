// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package hwlog test file
package hwlog

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUtilsFunc(t *testing.T) {
	convey.Convey("test utils", t, func() {
		convey.Convey("test utils func", func() {
			lg := new(logger)
			conf := &LogConfig{OnlyToStdout: true}
			userCtx := context.TODO()
			userCtx = context.WithValue(userCtx, UserID, 0)
			userCtx = context.WithValue(userCtx, ReqID, 0)
			err := lg.setLogger(conf)
			convey.So(err, convey.ShouldBeNil)
			printHelper(lg.lgInfo, "test", defaultMaxEachLineLen, true)
		})
	})
}
