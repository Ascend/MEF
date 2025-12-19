// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"io/fs"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
)

func TestCheckDir(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test check dir func", func() {
			mockStat := gomonkey.ApplyFunc(os.Stat, func(_ string) (fs.FileInfo, error) {
				return nil, os.ErrNotExist
			})
			mockMkDir := gomonkey.ApplyFunc(os.MkdirAll, func(_ string, _ fs.FileMode) error {
				return nil
			})
			defer mockStat.Reset()
			defer mockMkDir.Reset()
			err := checkDir("log")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCreateFile(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test create file", func() {
			mockExist := gomonkey.ApplyFunc(fileutils.IsExist, func(_ string) bool {
				return false
			})
			mockCreate := gomonkey.ApplyFunc(os.Create, func(_ string) (*os.File, error) {
				return nil, nil
			})
			defer mockExist.Reset()
			defer mockCreate.Reset()
			err := createFile("log")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCheckAndCreateLogFile(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test checkAndCreateLogFile func", func() {
			mockCreate := gomonkey.ApplyFunc(createFile, func(_ string) error {
				return nil
			})
			defer mockCreate.Reset()
			err := checkAndCreateLogFile("log")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestValidateLogConfigFileMaxSize(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate max size func", func() {
			conf := &LogConfig{}
			err := validateLogConfigFileMaxSize(conf)
			convey.So(err, convey.ShouldBeNil)
			convey.So(conf.FileMaxSize, convey.ShouldEqual, DefaultFileMaxSize)
			conf.FileMaxSize = -1
			err = validateLogConfigFileMaxSize(conf)
			convey.So(err, convey.ShouldBeError)
			conf.FileMaxSize = MaximumFileMaxSize + 1
			err = validateLogConfigFileMaxSize(conf)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestValidateLogConfigBackups(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate backups func", func() {
			conf := &LogConfig{MaxBackups: DefaultMaxBackups}
			err := validateLogConfigBackups(conf)
			convey.So(err, convey.ShouldBeNil)
			conf.MaxBackups = 0
			err = validateLogConfigBackups(conf)
			convey.So(err, convey.ShouldBeError)
			conf.FileMaxSize = DefaultMaxBackups + 1
			err = validateLogConfigBackups(conf)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestValidateLogConfigMaxAge(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate max age func", func() {
			conf := &LogConfig{MaxAge: DefaultMinSaveAge}
			err := validateLogConfigMaxAge(conf)
			convey.So(err, convey.ShouldBeNil)
			conf.MaxAge = 0
			err = validateLogConfigMaxAge(conf)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestValidateLogLevel(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate log level func", func() {
			conf := &LogConfig{}
			err := validateLogLevel(conf)
			convey.So(err, convey.ShouldBeNil)
			conf.LogLevel = minLogLevel - 1
			err = validateLogLevel(conf)
			convey.So(err, convey.ShouldBeError)
			conf.LogLevel = maxLogLevel + 1
			err = validateLogLevel(conf)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestValidateMaxLineLength(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate max line length func", func() {
			conf := &LogConfig{}
			err := validateMaxLineLength(conf)
			convey.So(err, convey.ShouldBeNil)
			convey.So(conf.MaxLineLength, convey.ShouldEqual, defaultMaxEachLineLen)
			conf.MaxLineLength = -1
			err = validateMaxLineLength(conf)
			convey.So(err, convey.ShouldNotBeNil)
			conf.MaxLineLength = maxEachLineLen + 1
			err = validateMaxLineLength(conf)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestValidateLogConfigFiled(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test validate config filed func", func() {
			mockCheckPath := gomonkey.ApplyFunc(fileutils.CheckOriginPath, func(_ string) (string, error) {
				return "", nil
			})
			mockCheckAndCreate := gomonkey.ApplyFunc(checkAndCreateLogFile, func(_ string) error {
				return nil
			})
			defer mockCheckPath.Reset()
			defer mockCheckAndCreate.Reset()
			conf := &LogConfig{
				MaxBackups:  DefaultMaxBackups,
				MaxAge:      DefaultMinSaveAge,
				CacheSize:   DefaultCacheSize,
				ExpiredTime: DefaultExpiredTime,
			}
			err := validateLogConfigFiled(conf)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestChangeFileMode(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test changeFileMode func", func() {
			changeFileMode(nil, fsnotify.Event{}, "log")
			mockExist := gomonkey.ApplyFunc(fileutils.IsExist, func(_ string) bool {
				return true
			})
			mockChmod := gomonkey.ApplyFunc(os.Chmod, func(_ string, _ fs.FileMode) error {
				return nil
			})
			defer mockExist.Reset()
			defer mockChmod.Reset()
			lg := new(logger)
			evt := fsnotify.Event{Name: "run-2022-01-01T00-00-00.123.log"}
			changeFileMode(lg, evt, "log")
		})
	})
}
