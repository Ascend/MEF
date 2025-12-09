// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"os"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// TestLimit for test the function of write with limit
func TestOpenLimit(t *testing.T) {
	convey.Convey("test Limiter", t, func() {
		convey.Convey("test limited write func", func() {
			l := &LogLimiter{
				Logs: &Logs{
					FileName: "testOpenLimit.log",
				},
				CacheSize:   DefaultCacheSize,
				ExpiredTime: DefaultExpiredTime,
			}
			defer removeFile("testOpenLimit.log")
			defer closeLog(l)
			input := []byte("[INFO]     2023/01/01 01:00:00.111111 1        foobarfoobarfoobarfoobar!")
			unlimitedWrite(input, l)
			for i := 0; i < 1000; i++ {
				limitedWrite(input, l)
			}
			time.Sleep(time.Second)
			unlimitedWrite(input, l)
		})
	})
}

// TestCloseLimit for test the function of write without limit
func TestCloseLimit(t *testing.T) {
	convey.Convey("test Limiter", t, func() {
		convey.Convey("test unlimited write func", func() {
			l := &LogLimiter{
				Logs: &Logs{
					FileName: "testCloseLimit.log",
				},
			}
			defer removeFile("testCloseLimit.log")
			defer closeLog(l)
			input := []byte("[INFO]     2023/01/01 01:00:00.111111 1        foobarfoobarfoobarfoobar!")
			unlimitedWrite(input, l)
			for i := 0; i < 1000; i++ {
				unlimitedWrite(input, l)
			}
		})
	})
}

// TestValidateLimiterConf for test the function of validate limiter conf
func TestValidateLimiterConf(t *testing.T) {
	convey.Convey("test Limiter", t, func() {
		convey.Convey("test validate limiter config func", func() {
			l := &LogLimiter{
				Logs:        &Logs{},
				CacheSize:   -1,
				ExpiredTime: -1,
			}
			l.validateLimiterConf()
			convey.So(l.CacheSize, convey.ShouldEqual, DefaultCacheSize)
			convey.So(l.ExpiredTime, convey.ShouldEqual, DefaultExpiredTime)
		})
	})
}

// TestNullPoint for test the null point exception
func TestNullPoint(t *testing.T) {
	convey.Convey("test Limiter", t, func() {
		convey.Convey("test null point exception", func() {
			l := &LogLimiter{
				Logs: &Logs{
					FileName: "testNullPoint.log",
				},
			}
			tmpPoint := l
			defer removeFile("testNullPoint.log")
			defer closeLog(tmpPoint)
			l.logCache = nil
			input := []byte("[INFO]     2023/01/01 01:00:00.111111 1        foobarfoobarfoobarfoobar!")
			unlimitedWrite(input, l)
			l = nil
			forbidWrite(input, l)
			err := l.Close()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func closeLog(l *LogLimiter) {
	err := l.Close()
	convey.So(err, convey.ShouldBeNil)
}

func removeFile(name string) {
	err := os.Remove(name)
	convey.So(err, convey.ShouldBeNil)
}

func unlimitedWrite(b []byte, l *LogLimiter) {
	n, err := l.Write(b)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(b), convey.ShouldEqual, n)
}

func limitedWrite(b []byte, l *LogLimiter) {
	n, err := l.Write(b)
	convey.So(err, convey.ShouldBeNil)
	convey.So(n, convey.ShouldEqual, 0)
}

func forbidWrite(b []byte, l *LogLimiter) {
	n, err := l.Write(b)
	convey.So(err, convey.ShouldNotBeNil)
	convey.So(n, convey.ShouldEqual, 0)
}
