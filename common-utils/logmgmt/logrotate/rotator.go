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
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

const (
	metaBytes               = 1 << 20
	backupTimeFormat        = "2006-01-02T15-04-05.000"
	defaultBackupPermission = 0400
	defaultLogPermission    = 0600
	gZipExt                 = ".gz"
)

// LogRotator provides log rotation function for third-party software
type LogRotator struct {
	configs Configs
}

// Configs configuration for a log rotator
type Configs struct {
	CheckIntervalSeconds int
	Logs                 []Config
}

// Config log rotation configuration for single log
type Config struct {
	LogFile    string
	BackupDir  string
	MaxBackups int
	MaxSizeMB  int
}

// New creates a new log rotator
func New(configs Configs) *LogRotator {
	return &LogRotator{configs: configs}
}

// Start run log rotator
func (l *LogRotator) Start(ctx context.Context) {
	l.checkLogs()
	duration := time.Duration(l.configs.CheckIntervalSeconds) * time.Second
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		timer.Reset(duration)
		l.checkLogs()
	}
}

func (l *LogRotator) makeSureBackupSpaceAvailable(logConfig Config, backupFileSize int64) error {
	logBaseName := filepath.Base(logConfig.LogFile)
	backupFiles, err := getBackupFiles(logBaseName, logConfig.BackupDir)
	if err != nil {
		return err
	}
	for _, fileName := range backupFiles {
		if envutils.CheckDiskSpace(filepath.Dir(logConfig.BackupDir), uint64(backupFileSize)) == nil {
			return nil
		}
		fullPath := filepath.Join(logConfig.BackupDir, fileName)
		if err := fileutils.DeleteFile(fullPath); err != nil {
			hwlog.RunLog.Errorf("failed to remove backup file, reason:%s", err.Error())
			continue
		}
	}
	return envutils.CheckDiskSpace(filepath.Dir(logConfig.BackupDir), uint64(backupFileSize))
}

func (l *LogRotator) checkLogs() {
	for _, logConfig := range l.configs.Logs {
		if _, err := fileutils.CheckOriginPath(logConfig.LogFile); err != nil {
			hwlog.RunLog.Errorf("check log file %s failed", logConfig.LogFile)
			continue
		}
		if _, err := fileutils.CheckOriginPath(logConfig.BackupDir); err != nil {
			hwlog.RunLog.Errorf("check log backup dir %s failed", logConfig.BackupDir)
			continue
		}
		logBaseName := filepath.Base(logConfig.LogFile)
		stat, err := os.Stat(logConfig.LogFile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				hwlog.RunLog.Errorf("can't get file size, %s", logBaseName)
			}
			continue
		}
		if stat.Size() > int64(logConfig.MaxSizeMB*metaBytes) {
			if err := l.makeSureBackupSpaceAvailable(logConfig, stat.Size()); err != nil {
				hwlog.RunLog.Errorf("make sure backup space available failed, %s", err.Error())
				continue
			}
			if err := backupAndTruncateLog(logConfig.LogFile, logConfig.BackupDir); err != nil {
				hwlog.RunLog.Errorf("failed to backup log: %s", err.Error())
			} else {
				hwlog.RunLog.Info("backup log success")
			}
		}
		cleanBackupFiles(logBaseName, logConfig.BackupDir, logConfig.MaxBackups)
	}
}

func backupAndTruncateLog(logFile, backupDir string) error {
	if _, err := fileutils.CheckOriginPath(logFile); err != nil {
		return err
	}
	dst := getBackupFileName(filepath.Base(logFile), backupDir)
	if err := copyAndCompress(dst, logFile); err != nil {
		return err
	}
	return os.Truncate(logFile, 0)
}

func cleanBackupFiles(logBaseName, backupDir string, maxBackups int) {
	backupFiles, err := getBackupFiles(logBaseName, backupDir)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read backup dir: %s", err.Error())
		return
	}
	remains := len(backupFiles)
	for _, fileName := range backupFiles {
		if remains <= maxBackups {
			break
		}
		fullPath := filepath.Join(backupDir, fileName)
		if err := fileutils.DeleteFile(fullPath); err != nil {
			hwlog.RunLog.Errorf("failed to remove backup file, reason:%s", err.Error())
			continue
		}
		hwlog.RunLog.Infof("clean backup log %s success", fileName)
		remains -= 1
	}
}

func getBackupFiles(logBaseName, backupDir string) ([]string, error) {
	type fileAndTime struct {
		fileName   string
		backupTime time.Time
	}

	prefix, suffix := getPrefixAndExt(logBaseName)
	suffix += gZipExt
	handle, entries, err := fileutils.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	var fileAndTimeList []fileAndTime
	for _, e := range entries {
		fileName := e.Name()
		if !(strings.HasPrefix(fileName, prefix+"-") && strings.HasSuffix(fileName, suffix)) {
			continue
		}
		if len(prefix)+1 >= len(fileName) || len(fileName) < len(suffix) {
			continue
		}
		timestamp := fileName[len(prefix)+1 : len(fileName)-len(suffix)]
		backupTime, err := time.Parse(backupTimeFormat, timestamp)
		if err != nil {
			continue
		}
		fileAndTimeList = append(fileAndTimeList, fileAndTime{
			fileName:   fileName,
			backupTime: backupTime,
		})
	}
	sort.Slice(fileAndTimeList, func(i, j int) bool {
		if i >= len(fileAndTimeList) {
			return true
		}
		return fileAndTimeList[i].backupTime.Before(fileAndTimeList[j].backupTime)
	})
	var fileNames []string
	for i := range fileAndTimeList {
		fileNames = append(fileNames, fileAndTimeList[i].fileName)
	}
	return fileNames, nil
}

func getPrefixAndExt(baseName string) (string, string) {
	ext := filepath.Ext(baseName)
	prefix := baseName[:len(baseName)-len(ext)]
	return prefix, ext
}

func getBackupFileName(logBaseName, backupDir string) string {
	prefix, oldExt := getPrefixAndExt(logBaseName)
	timeStamp := time.Now().Format(backupTimeFormat)
	backupBaseName := fmt.Sprintf("%s-%s%s%s", prefix, timeStamp, oldExt, gZipExt)
	return filepath.Join(backupDir, backupBaseName)
}

func copyAndCompress(dst, src string) (firstErr error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	var dstFileCreated bool
	defer func() {
		if err := srcFile.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			hwlog.RunLog.Error("backup: unable to close source file")
		}
		if firstErr != nil && dstFileCreated {
			if err := fileutils.DeleteFile(dst); err != nil {
				hwlog.RunLog.Error("backup: unable to clean destination file")
			}
		}
	}()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, defaultBackupPermission)
	if err != nil {
		firstErr = err
		return firstErr
	}
	dstFileCreated = true
	defer func() {
		if err := dstFile.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			hwlog.RunLog.Error("backup: unable to close destination file")
		}
	}()

	if firstErr = compress(dstFile, srcFile); firstErr != nil {
		return firstErr
	}
	return dstFile.Sync()
}

func compress(dst io.Writer, src io.Reader) error {
	gzipWriter := gzip.NewWriter(dst)
	_, firstErr := io.Copy(gzipWriter, src)
	if err := gzipWriter.Close(); err != nil {
		if firstErr == nil {
			firstErr = err
		}
		hwlog.RunLog.Error("backup: unable to close gzip stream")
	}
	return firstErr
}
