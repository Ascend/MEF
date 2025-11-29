/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package common a series of common function
package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestMapDeepCopy for test MapDeepCopy
func TestMapDeepCopy(t *testing.T) {
	convey.Convey("test MapDeepCopy", t, func() {
		convey.Convey("input nil", func() {
			ret := MapDeepCopy(nil)
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("h.Write success", func() {
			devices := map[string]string{"100": DefaultDeviceIP}
			ret := MapDeepCopy(devices)
			convey.So(len(ret), convey.ShouldEqual, len(devices))
		})
	})
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(SocketChmod); err != nil {
		return err
	}
	return nil
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	convey.Convey("pods is nil", t, func() {
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, fmt.Errorf("err")
		})
		defer mockStat.Reset()
		_, err := GetDefaultDevices(true)
		convey.So(err, convey.ShouldNotBeNil)
	})
	if _, err := os.Stat(HiAIHDCDevice); err != nil {
		if err = createFile(HiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		if err = createFile(HiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAISVMDevice); err != nil {
		if err = createFile(HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	defaultDevices, err := GetDefaultDevices(true)
	if err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]string)
	defaultMap[HiAIHDCDevice] = ""
	defaultMap[HiAIManagerDevice] = ""
	defaultMap[HiAISVMDevice] = ""
	defaultMap[HiAi200RCEventSched] = ""
	defaultMap[HiAi200RCHiDvpp] = ""
	defaultMap[HiAi200RCLog] = ""
	defaultMap[HiAi200RCMemoryBandwidth] = ""
	defaultMap[HiAi200RCSVM0] = ""
	defaultMap[HiAi200RCTsAisle] = ""
	defaultMap[HiAi200RCUpgrade] = ""

	for _, str := range defaultDevices {
		if _, ok := defaultMap[str]; !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

// TestVerifyPath for VerifyPath
func TestVerifyPath(t *testing.T) {
	convey.Convey("TestVerifyPath", t, func() {
		convey.Convey("filepath.Abs failed", func() {
			mock := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestWatchFile for test watchFile
func TestWatchFile(t *testing.T) {
	convey.Convey("TestWatchFile", t, func() {
		convey.Convey("fsnotify.NewWatcher ok", func() {
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldBeNil)
			convey.So(watcher, convey.ShouldNotBeNil)
		})
		convey.Convey("fsnotify.NewWatcher failed", func() {
			mock := gomonkey.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
				return nil, fmt.Errorf("error")
			})
			defer mock.Reset()
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(watcher, convey.ShouldBeNil)
		})
		watcher, _ := NewFileWatch()
		convey.Convey("stat failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		defer mockStat.Reset()
		convey.Convey("Add failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return fmt.Errorf("err") })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("Add ok", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return nil })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestNewSignWatcher for test NewSignWatcher
func TestNewSignWatcher(t *testing.T) {
	convey.Convey("TestNewSignWatcher", t, func() {
		signChan := NewSignWatcher(syscall.SIGHUP)
		convey.So(signChan, convey.ShouldNotBeNil)
	})
}
