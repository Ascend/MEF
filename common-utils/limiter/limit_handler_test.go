//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package limiter implement a token bucket limiter
package limiter

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, context.TODO())
}

func TestReturnToken(t *testing.T) {
	const halfSecond = time.Millisecond * 500
	timer := time.NewTimer(halfSecond)
	convey.Convey("test returnToken", t, func() {
		mock := gomonkey.ApplyFunc(time.NewTimer, func(time.Duration) *time.Timer {
			return timer
		})
		defer mock.Reset()
		sc := make(chan struct{}, 1)
		go returnToken(context.Background(), sc)
		time.Sleep(time.Second)
		convey.So(len(sc), convey.ShouldEqual, 1)
	})
}

func TestNewLimitHandlerV2(t *testing.T) {
	conf := &HandlerConfig{
		PrintLog:         false,
		Method:           "",
		LimitBytes:       DefaultDataLimit,
		TotalConCurrency: defaultMaxConcurrency,
		IPConCurrency:    "2/1",
		CacheSize:        DefaultCacheSize,
	}
	convey.Convey("normal situation,no err return", t, func() {
		_, err := NewLimitHandlerV2(http.DefaultServeMux, conf)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("IPConCurrency parameter error", t, func() {
		conf.IPConCurrency = "2021/1"
		_, err := NewLimitHandlerV2(http.DefaultServeMux, conf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("cacheSize parameter error", t, func() {
		conf.CacheSize = 0
		_, err := NewLimitHandlerV2(http.DefaultServeMux, conf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("method parameter error", t, func() {
		conf.Method = "20/iajsdkjas2jhjdklsjkldjsdfasd1"
		_, err := NewLimitHandlerV2(http.DefaultServeMux, conf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("TotalConCurrency parameter error", t, func() {
		conf.TotalConCurrency = 0
		_, err := NewLimitHandlerV2(http.DefaultServeMux, conf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}
