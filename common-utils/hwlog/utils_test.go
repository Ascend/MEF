//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

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
