// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package logrotate provides log rotation function for third-party software
package logrotate

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
)

const (
	size1            = 1
	fileCount1       = 1
	fileCount2       = 2
	randCount        = 64
	logDirPermission = 0600
)

var testRoot = "test"

func setup() (func(), error) {
	var err error
	testRoot, err = filepath.Abs(testRoot)
	if err != nil {
		return nil, fmt.Errorf("get testRoot abs path failed: %s", err.Error())
	}
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		return nil, err
	}
	if err = hwlog.InitOperateLogger(logConfig, context.Background()); err != nil {
		return nil, err
	}
	if err = os.RemoveAll(testRoot); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return func() {}, nil
}

// TestMain setup test environment
func TestMain(m *testing.M) {
	teardown, err := setup()
	if err != nil {
		fmt.Printf("failed to setup test environment, reason:%s\n", err.Error())
		return
	}
	defer teardown()
	code := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", code)
}

// TestCleanBackupFiles test cleanBackupFiles
func TestCleanBackupFiles(t *testing.T) {
	convey.Convey("TestCleanBackupFiles", t, func() {
		logBaseName := "TestCleanBackupFiles.log"
		backupDir := filepath.Join(testRoot, logBaseName+".backup")

		currentTime := time.Now()
		mockTime(currentTime, func() {
			_, err := createBackupFile(logBaseName, backupDir)
			convey.So(err, convey.ShouldBeNil)
		})
		cleanBackupFiles(logBaseName, backupDir, fileCount2)
		count, err := countFiles(backupDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(count, convey.ShouldEqual, fileCount1)

		currentTime = currentTime.Add(time.Second)
		mockTime(currentTime, func() {
			_, err := createBackupFile(logBaseName, backupDir)
			convey.So(err, convey.ShouldBeNil)
		})

		currentTime = currentTime.Add(time.Second)
		mockTime(currentTime, func() {
			_, err := createBackupFile(logBaseName, backupDir)
			convey.So(err, convey.ShouldBeNil)
		})

		cleanBackupFiles(logBaseName, backupDir, fileCount2)
		count, err = countFiles(backupDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(count, convey.ShouldEqual, fileCount2)
	})
}

// TestGetBackupFiles test getBackupFiles
func TestGetBackupFiles(t *testing.T) {
	convey.Convey("TestGetBackupFiles", t, func() {
		logBaseName := "TestGetBackupFiles.log"
		backupDir := filepath.Join(testRoot, logBaseName+".backup")

		currentTime := time.Now()
		mockTime(currentTime, func() {
			_, err := createBackupFile(logBaseName, backupDir)
			convey.So(err, convey.ShouldBeNil)
		})

		files, err := getBackupFiles(logBaseName, backupDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(files), convey.ShouldEqual, fileCount1)

		currentTime = currentTime.Add(time.Second)
		var fileName string
		mockTime(currentTime, func() {
			fileName, err = createBackupFile(logBaseName, backupDir)
			convey.So(err, convey.ShouldBeNil)
		})

		files, err = getBackupFiles(logBaseName, backupDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(files), convey.ShouldEqual, fileCount2)

		err = os.Rename(fileName, fileName+".1")
		convey.So(err, convey.ShouldBeNil)

		files, err = getBackupFiles(logBaseName, backupDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(files), convey.ShouldEqual, fileCount1)
	})
}

// TestCopyAndCompress test copyAndCompress
func TestCopyAndCompress(t *testing.T) {
	convey.Convey("TestCopyAndCompress", t, func() {
		logBaseName := "TestCopyAndCompress.log"
		backupDir := filepath.Join(testRoot, logBaseName+".backup")
		logDir := filepath.Join(testRoot, logBaseName)
		logFile := filepath.Join(logDir, logBaseName)

		fileName, err := createBackupFile(logBaseName, logDir)
		convey.So(err, convey.ShouldBeNil)
		err = os.Rename(fileName, logFile)
		convey.So(err, convey.ShouldBeNil)

		backupFile := getBackupFileName(logBaseName, backupDir)
		err = os.MkdirAll(backupDir, logDirPermission)
		convey.So(err, convey.ShouldBeNil)
		err = copyAndCompress(backupFile, logFile)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCheckLog test checkLog
func TestCheckLogs(t *testing.T) {
	convey.Convey("TestCheckLogs", t, func() {
		logBaseName := "TestCheckLogs.log"
		logDir := filepath.Join(testRoot, logBaseName)
		logFileName := filepath.Join(logDir, logBaseName)
		backupDir := filepath.Join(testRoot, logBaseName+".backup")
		configs := Configs{
			CheckIntervalSeconds: 1,
			Logs: []Config{
				{
					LogFile:    logFileName,
					BackupDir:  backupDir,
					MaxBackups: fileCount2,
					MaxSizeMB:  size1,
				},
			},
		}
		rotator := New(configs)
		currentTime := time.Now()
		err := os.MkdirAll(backupDir, logDirPermission)
		convey.So(err, convey.ShouldBeNil)

		testCheckLogsStep1(rotator, logFileName, backupDir, currentTime)
		currentTime = currentTime.Add(time.Second)
		testCheckLogsStep2(rotator, logFileName, backupDir, currentTime)
		currentTime = currentTime.Add(time.Second)
		testCheckLogsStep3(rotator, logFileName, backupDir, currentTime)

	})
}

func testCheckLogsStep1(rotator *LogRotator, logFileName, backupDir string, currentTime time.Time) {

	data, err := fillData(logFileName, size1*metaBytes+1)
	convey.So(err, convey.ShouldBeNil)
	mockTime(currentTime, func() {
		rotator.checkLogs()
	})
	count, err := countFiles(backupDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(count, convey.ShouldEqual, fileCount1)
	newestFile, err := getNewestFile(filepath.Base(logFileName), backupDir)
	convey.So(err, convey.ShouldBeNil)
	checkFileContent(logFileName, []byte{})
	checkGzipFileContent(newestFile, data)
}

func testCheckLogsStep2(rotator *LogRotator, logFileName, backupDir string, currentTime time.Time) {
	data, err := fillData(logFileName, size1*metaBytes+1)
	mockTime(currentTime, func() {
		rotator.checkLogs()
	})
	convey.So(err, convey.ShouldBeNil)
	count, err := countFiles(backupDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(count, convey.ShouldEqual, fileCount2)
	newestFile, err := getNewestFile(filepath.Base(logFileName), backupDir)
	convey.So(err, convey.ShouldBeNil)
	checkFileContent(logFileName, []byte{})
	checkGzipFileContent(newestFile, data)
}

func testCheckLogsStep3(rotator *LogRotator, logFileName, backupDir string, currentTime time.Time) {
	data, err := fillData(logFileName, size1*metaBytes+1)
	currentTime = currentTime.Add(time.Second)
	mockTime(currentTime, func() {
		rotator.checkLogs()
	})
	convey.So(err, convey.ShouldBeNil)
	count, err := countFiles(backupDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(count, convey.ShouldEqual, fileCount2)
	newestFile, err := getNewestFile(filepath.Base(logFileName), backupDir)
	convey.So(err, convey.ShouldBeNil)
	checkFileContent(logFileName, []byte{})
	checkGzipFileContent(newestFile, data)
}

func createBackupFile(logBaseName, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, logDirPermission); err != nil {
		return "", err
	}
	fileName := getBackupFileName(logBaseName, backupDir)
	if f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultLogPermission); err != nil {
		return "", err
	} else if err := f.Close(); err != nil {
		return "", err
	}
	return fileName, nil
}

func fillData(filename string, length int) ([]byte, error) {
	if err := os.MkdirAll(filepath.Dir(filename), logDirPermission); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultLogPermission)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file, %s\n", err.Error())
		}
	}()

	if length >= math.MaxInt {
		length = 1
	}
	data := make([]byte, length)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < randCount; i++ {
		index := random.Intn(length)
		if index > len(data) || index < 0 {
			continue
		}
		data[index] = 1
	}

	if _, err := f.Write(data); err != nil {
		return nil, err
	}
	return data, nil
}

func checkGzipFileContent(file string, data []byte) {
	f, err := os.Open(file)
	convey.So(err, convey.ShouldBeNil)
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file, %s\n", err.Error())
		}
	}()
	gf, err := gzip.NewReader(f)
	convey.So(err, convey.ShouldBeNil)
	realData, err := io.ReadAll(gf)
	convey.So(err, convey.ShouldBeNil)
	convey.So(bytes.Compare(data, realData), convey.ShouldBeZeroValue)
}

func checkFileContent(file string, data []byte) {
	f, err := os.Open(file)
	convey.So(err, convey.ShouldBeNil)
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file, %s\n", err.Error())
		}
	}()
	realData, err := io.ReadAll(f)
	convey.So(err, convey.ShouldBeNil)
	convey.So(bytes.Compare(data, realData), convey.ShouldBeZeroValue)
}

func getNewestFile(logBaseName, backupDir string) (string, error) {
	backupFiles, err := getBackupFiles(logBaseName, backupDir)
	if err != nil {
		return "", err
	}
	if len(backupFiles) < 1 {
		return "", errors.New("not enough files")
	}
	return filepath.Join(backupDir, backupFiles[len(backupFiles)-1]), nil
}

func countFiles(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

func mockTime(currentTime time.Time, f func()) {
	patch := gomonkey.ApplyFunc(time.Now, func() time.Time {
		return currentTime
	})
	defer patch.Reset()
	f()
}
