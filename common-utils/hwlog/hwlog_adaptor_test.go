//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog test file
package hwlog

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestInitRunLogger(t *testing.T) {
	convey.Convey("test hwlog adaptor", t, func() {
		convey.Convey("test init run log", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			err := InitRunLogger(nil, ctx)
			convey.So(err, convey.ShouldBeError, errors.New("run logger config is nil"))
			lgConfig := &LogConfig{OnlyToStdout: true}
			err = InitRunLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			// repeat initialize
			err = InitRunLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			cancel()
		})
	})
}

func TestInitOperateLogger(t *testing.T) {
	convey.Convey("test hwlog adaptor", t, func() {
		convey.Convey("test init operate log", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			err := InitOperateLogger(nil, ctx)
			convey.So(err, convey.ShouldBeError, errors.New("operate logger config is nil"))
			lgConfig := &LogConfig{OnlyToStdout: true}
			err = InitOperateLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			// repeat initialize
			err = InitOperateLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			cancel()
		})
	})
}

func TestInitSecurityLogger(t *testing.T) {
	convey.Convey("test hwlog adaptor", t, func() {
		convey.Convey("test init security log", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			err := InitSecurityLogger(nil, ctx)
			convey.So(err, convey.ShouldBeError, errors.New("security logger config is nil"))
			lgConfig := &LogConfig{OnlyToStdout: true}
			err = InitSecurityLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			// repeat initialize
			err = InitSecurityLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			cancel()
		})
	})
}

func TestInitUserLogger(t *testing.T) {
	convey.Convey("test hwlog adaptor", t, func() {
		convey.Convey("test init user log", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			err := InitUserLogger(nil, ctx)
			convey.So(err, convey.ShouldBeError, errors.New("user logger config is nil"))
			lgConfig := &LogConfig{OnlyToStdout: true}
			err = InitUserLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			// repeat initialize
			err = InitUserLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			cancel()
		})
	})
}

func TestInitDebugLogger(t *testing.T) {
	convey.Convey("test hwlog adaptor", t, func() {
		convey.Convey("test init debug log", func() {
			ctx, cancel := context.WithCancel(context.TODO())
			err := InitDebugLogger(nil, ctx)
			convey.So(err, convey.ShouldBeError, errors.New("debug logger config is nil"))
			lgConfig := &LogConfig{OnlyToStdout: true}
			err = InitDebugLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			// repeat initialize
			err = InitDebugLogger(lgConfig, ctx)
			convey.So(err, convey.ShouldBeNil)
			cancel()
		})
	})
}
