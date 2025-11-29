// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common methods for install or upgrade
package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const (
	dfCmd           = "df"
	filesystemP7    = "/dev/mmcblk0p7"
	firmwareDir     = "firmware"
	columnCount     = 6
	filesystemIndex = 0
	mountedIndex    = 5
)

// CheckLogDirs check the validity of log dir and log backup dir
func CheckLogDirs(logRootDir, logRootBackupDir string, allowTmpfs bool) error {
	if err := CheckDir(logRootDir, constants.LogDirName); err != nil {
		return err
	}
	if err := CheckDir(logRootBackupDir, constants.LogBackupDirName); err != nil {
		return err
	}
	if err := CheckInTmpfs(logRootBackupDir, allowTmpfs); err != nil {
		return err
	}
	return nil
}

// CheckDir Check the validity of directories during the installation and upgrade
func CheckDir(dir, dirName string) error {
	if !utils.IsFlagSet(dirName) {
		if err := fileutils.CreateDir(dir, constants.Mode755); err != nil {
			return fmt.Errorf("create dir [%s] failed, error: %v", dirName, err)
		}
	}

	if !fileutils.IsExist(dir) {
		return fmt.Errorf("dir [%s] does not exist", dirName)
	}

	if !filepath.IsAbs(dir) {
		return fmt.Errorf("dir [%s] is not absolute path", dirName)
	}

	if strings.HasPrefix(dir, constants.UnpackPath) {
		return fmt.Errorf("dir [%s] cannot be in the decompression path", dirName)
	}

	if _, err := fileutils.RealDirCheck(dir, true, false); err != nil {
		return fmt.Errorf("check dir [%s] failed, error: %v", dirName, err)
	}
	return nil
}

// CheckInTmpfs check whether the path is allowed in the temporary file system
func CheckInTmpfs(dir string, allowTmpfs bool) error {
	isInTmpfs, err := envutils.IsInTmpfs(dir)
	if err != nil {
		return err
	}
	if !isInTmpfs {
		return nil
	}
	if allowTmpfs {
		fmt.Println("the dir is allowed in the temporary file system")
		return nil
	}
	return errors.New("the dir cannot be in the tmpfs filesystem")
}

// CopyResetScriptToP7 copy reset script to filesystem p7
func CopyResetScriptToP7() error {
	out, err := envutils.RunCommand(dfCmd, envutils.DefCmdTimeoutSec, "-h")
	if err != nil {
		return fmt.Errorf("execute [%s] command failed, error: %v", dfCmd, err)
	}

	lines := strings.Split(out, "\n")
	iterationCount := 1
	for _, line := range lines {
		if iterationCount > constants.MaxIterationCount {
			break
		}
		iterationCount++
		parts := strings.Fields(line)
		if len(parts) < columnCount {
			return errors.New("command output format invalid")
		}
		if parts[filesystemIndex] != filesystemP7 {
			continue
		}
		softwareDir, err := path.GetCompWorkDir()
		if err != nil {
			return fmt.Errorf("get component work dir failed, error: %v", err)
		}
		firmwarePath := filepath.Join(parts[mountedIndex], firmwareDir)
		if !fileutils.IsExist(firmwarePath) {
			return fmt.Errorf("firmware path [%s] does not exist", firmwarePath)
		}

		checker := fileutils.NewFileLinkChecker(true)
		checker.SetNext(fileutils.NewFileModeChecker(true, constants.ModeUmask022, false, false))
		checker.SetNext(fileutils.NewFileOwnerChecker(true, true, constants.RootUserUid, constants.RootUserGid))

		resetScriptSrc := filepath.Join(softwareDir, constants.Script, constants.ResetMiddlewareScript)
		resetScriptDst := filepath.Join(firmwarePath, constants.ResetMiddlewareScript)
		if err = fileutils.DeleteFile(resetScriptDst, checker); err != nil {
			return fmt.Errorf("delete existed [%s] failed, error: %v", resetScriptDst, err)
		}
		if err = fileutils.CopyFile(resetScriptSrc, resetScriptDst, checker); err != nil {
			return fmt.Errorf("copy [%s] to [%s] failed, error: %v", resetScriptSrc, resetScriptDst, err)
		}
		hwlog.RunLog.Info("copy reset script to filesystem p7 success")
		return nil
	}
	return errors.New("the path of filesystem p7 is not found")
}
