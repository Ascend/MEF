// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util
package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const (
	argInplace = "--inplace"
	argAz      = "-az"
)

type logFileInfo struct {
	component string
	suffix    string
	user      string
}

func (l logFileInfo) path() string {
	return filepath.Join(l.component, fmt.Sprintf("%s_%s", l.component, l.suffix))
}

var logFileList = []logFileInfo{
	{component: constants.EdgeCore, suffix: constants.RunLogFile, user: constants.RootUserName},
	{component: constants.DevicePlugin, suffix: constants.RunLogFile, user: constants.RootUserName},
	{component: constants.EdgeMain, suffix: constants.RunLogFile, user: constants.EdgeUserName},
	{component: constants.EdgeMain, suffix: constants.OperateLogFile, user: constants.EdgeUserName},
	{component: constants.EdgeOm, suffix: constants.RunLogFile, user: constants.RootUserName},
	{component: constants.EdgeOm, suffix: constants.OperateLogFile, user: constants.RootUserName},
	{component: constants.EdgeInstaller, suffix: constants.RunLogFile, user: constants.RootUserName},
	{component: constants.EdgeInstaller, suffix: constants.OperateLogFile, user: constants.RootUserName},
}

// LogSyncMgr sync log object
type LogSyncMgr struct {
	logRootDir  string
	syncRootDir string
}

// NewLogSyncMgr create a new LogSyncMgr instance
func NewLogSyncMgr() *LogSyncMgr {
	return &LogSyncMgr{}
}

// RecoverLogs recover the log files from disk
func (lsm *LogSyncMgr) RecoverLogs() error {
	tasks := []func() error{
		lsm.initLogConfigs,
		lsm.copyLogFiles,
	}
	for _, function := range tasks {
		if err := function(); err != nil {
			return fmt.Errorf("restore log failed, error: %v", err)
		}
	}
	return nil
}

func (lsm *LogSyncMgr) copyLogFiles() error {
	isInTmpfs, err := envutils.IsInTmpfs(lsm.logRootDir)
	if err != nil {
		return err
	}
	if !isInTmpfs {
		return nil
	}

	for _, logInfo := range logFileList {
		if err := lsm.copyOneFile(logInfo.path(), logInfo.user); err != nil {
			return fmt.Errorf("recover log failed, error: %v", err)
		}
	}
	return nil
}

func (lsm *LogSyncMgr) copyOneFile(logFileRelPath, user string) error {
	logFilePath := filepath.Join(lsm.logRootDir, logFileRelPath)
	syncFilePath := filepath.Join(lsm.syncRootDir, logFileRelPath)

	var err error
	if logFilePath, err = checkAndPrepareLogFile(logFilePath, user); err != nil {
		return err
	}

	if !fileutils.IsLexist(syncFilePath) {
		return nil
	}
	if syncFilePath, err = checkAndPrepareLogFile(syncFilePath, user); err != nil {
		return err
	}
	if err := fileutils.CopyFile(syncFilePath, logFilePath); err != nil {
		return err
	}

	uid, gid, err := getUidAndGid(user)
	if err != nil {
		return err
	}

	param := fileutils.SetOwnerParam{
		Path:       logFilePath,
		Uid:        uid,
		Gid:        gid,
		Recursive:  false,
		IgnoreFile: true,
	}
	return fileutils.SetPathOwnerGroup(param)
}

// BackupLogs backup the log files to disk
func (lsm *LogSyncMgr) BackupLogs() error {
	tasks := []func() error{
		lsm.initLogConfigs,
		lsm.syncLogFiles,
	}
	for _, function := range tasks {
		if err := function(); err != nil {
			return fmt.Errorf("backup log failed, error: %v", err)
		}
	}
	return nil
}

func (lsm *LogSyncMgr) syncLogFiles() error {
	isInTmpfs, err := envutils.IsInTmpfs(lsm.logRootDir)
	if err != nil {
		return err
	}
	if !isInTmpfs {
		return nil
	}

	for _, logInfo := range logFileList {
		if err := lsm.syncOneFile(logInfo.path(), logInfo.user); err != nil {
			return fmt.Errorf("recover log failed, error: %v", err)
		}
	}
	return nil
}

func (lsm *LogSyncMgr) syncOneFile(logFileRelPath, user string) error {
	logFilePath := filepath.Join(lsm.logRootDir, logFileRelPath)
	syncFilePath := filepath.Join(lsm.syncRootDir, logFileRelPath)

	if !fileutils.IsLexist(logFilePath) {
		return nil
	}

	var err error
	if logFilePath, err = checkAndPrepareLogFile(logFilePath, user); err != nil {
		return err
	}
	if syncFilePath, err = checkAndPrepareLogFile(syncFilePath, user); err != nil {
		return err
	}

	uid, gid, err := getUidAndGid(user)
	if err != nil {
		return err
	}

	output, err := envutils.RunCommandWithUser(
		constants.RsyncCmd, constants.RsyncTimeWaitTime, uid, gid, argInplace, argAz, logFilePath, syncFilePath)
	if err != nil {
		return fmt.Errorf("execute rsync failed, error: %v, output: %s", err, output)
	}
	return nil
}

func (lsm *LogSyncMgr) initLogConfigs() error {
	installerLogDir, installerLogBackupDir, err := path.GetCompLogDirs(constants.EdgeInstaller)
	if err != nil {
		hwlog.RunLog.Errorf("get component log dirs failed, %v", err)
		return errors.New("get component log dirs failed")
	}
	lsm.logRootDir = filepath.Dir(installerLogDir)
	lsm.syncRootDir = filepath.Join(filepath.Dir(filepath.Dir(installerLogBackupDir)), constants.MEFEdgeLogSyncName)
	return nil
}

func checkAndPrepareLogFile(logFilePath, user string) (string, error) {
	var err error

	uid, gid, err := getUidAndGid(user)
	if err != nil {
		return "", err
	}

	if logFilePath, err = fileutils.CheckOriginPath(logFilePath); err != nil {
		return "", err
	}
	logDirPath := filepath.Dir(logFilePath)
	logDirRootPath := filepath.Dir(logDirPath)
	if !fileutils.IsLexist(logDirRootPath) {
		if err := mkdir(logDirRootPath, constants.Mode755); err != nil {
			return "", err
		}
	}
	if _, err := fileutils.RealDirCheck(logDirRootPath, true, false); err != nil {
		return "", err
	}

	if !fileutils.IsLexist(logDirPath) {
		if err = mkdir(logDirPath, constants.Mode750); err != nil {
			return "", err
		}
		param := fileutils.SetOwnerParam{
			Path:       logDirPath,
			Uid:        uid,
			Gid:        gid,
			Recursive:  false,
			IgnoreFile: true,
		}
		if err = fileutils.SetPathOwnerGroup(param); err != nil {
			return "", err
		}
	}
	if _, err := fileutils.CheckOwnerAndPermission(logDirPath, constants.ModeUmask027, uid); err != nil {
		return "", err
	}

	if fileutils.IsLexist(logFilePath) {
		if _, err := fileutils.CheckOwnerAndPermission(logFilePath, constants.ModeUmask137, uid); err != nil {
			return "", err
		}
	}

	return logFilePath, nil
}

func getUidAndGid(user string) (uint32, uint32, error) {
	uid, err := envutils.GetUid(user)
	if err != nil {
		return 0, 0, err
	}
	gid, err := envutils.GetGid(user)
	if err != nil {
		return 0, 0, err
	}
	return uid, gid, err
}

func mkdir(dirPath string, perm os.FileMode) error {
	if err := fileutils.CreateDir(dirPath, perm); err != nil {
		return err
	}
	return fileutils.SetPathPermission(dirPath, perm, false, true)
}
