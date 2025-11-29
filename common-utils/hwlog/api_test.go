//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog test file
package hwlog

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
)

func TestNewLogger(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test setLogger func", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			// test for log file
			mockPathCheck := gomonkey.ApplyFunc(fileutils.CheckOriginPath, func(_ string) (string, error) {
				return "", nil
			})
			mockMkdir := gomonkey.ApplyFunc(os.Chmod, func(_ string, _ fs.FileMode) error {
				return nil
			})
			defer mockPathCheck.Reset()
			defer mockMkdir.Reset()
			lgConfig = &LogConfig{
				LogFileName: path.Join(filepath.Dir(os.Args[0]), "t.log"),
				OnlyToFile:  true,
				MaxBackups:  DefaultMaxBackups,
				MaxAge:      DefaultMinSaveAge,
				CacheSize:   DefaultCacheSize,
				ExpiredTime: DefaultExpiredTime,
			}
			err = lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestLoggerPrint(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test logger print func", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
				LogLevel:     -1,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			lg.Debug("test debug")
			lg.Debugf("test debugf")
			lg.Info("test info")
			lg.Infof("test infof")
			lg.Warn("test warn")
			lg.Warnf("test warnf")
			lg.Error("test error")
			lg.Errorf("test errorf")
			lg.Critical("test critical")
			lg.Criticalf("test criticalf")
			lg.setLoggerLevel(maxLogLevel + 1)
			lg.Debug("test debug")
			lg.Debugf("test debugf")
			lg.Info("test info")
			lg.Infof("test infof")
			lg.Warn("test warn")
			lg.Warnf("test warnf")
			lg.Error("test error")
			lg.Errorf("test errorf")
			lg.Critical("test critical")
			lg.Criticalf("test criticalf")
		})
	})
}

func TestValidate(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test validate", func() {
			lg := new(logger)
			res := lg.validate()
			convey.So(res, convey.ShouldBeFalse)
			lgConfig := &LogConfig{
				OnlyToStdout: true,
			}
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			res = lg.validate()
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}
