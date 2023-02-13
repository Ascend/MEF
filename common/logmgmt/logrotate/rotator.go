// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logrotate provides log rotation function for third-party software
package logrotate

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

const (
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

func (l *LogRotator) checkLogs() {
	for _, logConfig := range l.configs.Logs {
		logBaseName := filepath.Base(logConfig.LogFile)
		stat, err := os.Stat(logConfig.LogFile)
		if err != nil {
			if !os.IsNotExist(err) {
				hwlog.RunLog.Errorf("can't get file size, %s", logBaseName)
			}
			return
		}
		if stat.Size() > int64(logConfig.MaxSizeMB*common.MB) {
			if err := backupAndTruncateLog(logConfig.LogFile, logConfig.BackupDir); err != nil {
				hwlog.RunLog.Errorf("failed to backup log: %s", err.Error())
			} else {
				hwlog.RunLog.Infof("backup log success")
			}
		}
		cleanBackupFiles(logBaseName, logConfig.BackupDir, logConfig.MaxBackups)
	}
}

func backupAndTruncateLog(logFile, backupDir string) error {
	dst := getBackupFileName(filepath.Base(logFile), backupDir)
	if err := copyAndCompress(dst, logFile); err != nil {
		return err
	}
	if err := os.Chmod(dst, defaultBackupPermission); err != nil {
		return err
	}
	return os.Truncate(logFile, 0)
}

func cleanBackupFiles(logBaseName, backupDir string, maxBackups int) (int, int) {
	backupFiles, err := getBackupFiles(logBaseName, backupDir)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read backup dir: %s", err.Error())
		return 0, 1
	}
	remains, fails := len(backupFiles), 0
	for _, fileName := range backupFiles {
		if remains <= maxBackups {
			break
		}
		fullPath := filepath.Join(backupDir, fileName)
		if err := os.Remove(fullPath); err != nil {
			hwlog.RunLog.Errorf("failed to remove backup file, reason:%s", err.Error())
			fails += 1
			continue
		}
		hwlog.RunLog.Infof("clean backup log %s success", fileName)
		remains -= 1
	}
	return remains, fails
}

func getBackupFiles(logBaseName, backupDir string) ([]string, error) {
	type fileAndTime struct {
		fileName   string
		backupTime time.Time
	}

	prefix, suffix := getPrefixAndExt(logBaseName)
	suffix += gZipExt
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}
	var fileAndTimeList []fileAndTime
	for _, e := range entries {
		fileName := e.Name()
		if strings.HasPrefix(fileName, prefix+"-") && strings.HasSuffix(fileName, suffix) {
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
			if err := os.Remove(dst); err != nil {
				hwlog.RunLog.Error("backup: unable to clean destination file")
			}
		}
	}()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, defaultLogPermission)
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

	if _, firstErr = compress(dstFile, srcFile); firstErr != nil {
		return firstErr
	}
	return dstFile.Sync()
}

func compress(dst io.Writer, src io.Reader) (int64, error) {
	gzipWriter := gzip.NewWriter(dst)
	nWrites, firstErr := io.Copy(gzipWriter, src)
	if err := gzipWriter.Close(); err != nil {
		if firstErr == nil {
			firstErr = err
		}
		hwlog.RunLog.Error("backup: unable to close gzip stream")
	}
	return nWrites, firstErr
}
