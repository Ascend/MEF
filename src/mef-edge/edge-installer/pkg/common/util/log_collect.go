// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util
package util

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/logmgmt/logcollect"

	"edge-installer/pkg/common/constants"
)

const (
	groupOrOtherWrite = 0022
	baseDir           = "edge"
	maxPackSize       = 200 * constants.MB
	maxFileSize       = 50 * constants.MB
)

// GetLogCollector create a log collector
func GetLogCollector(tarGzPath, logRootDir, logBackupRootDir string,
	collectPathWhiteList []string) logcollect.Collector {
	logFiles := logcollect.LogGroup{
		RootDir:   logRootDir,
		BaseDir:   baseDir,
		CheckFunc: checkLogFile,
	}
	logBackupFiles := logcollect.LogGroup{
		RootDir:   logBackupRootDir,
		BaseDir:   baseDir,
		CheckFunc: checkLogFile,
	}

	return logcollect.NewCollector(
		tarGzPath, []logcollect.LogGroup{logFiles, logBackupFiles}, maxPackSize, collectPathWhiteList)
}

func checkLogFile(filePath string) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if stat.Size() > maxFileSize {
		return errors.New("log file is too large")
	}
	uid, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return err
	}
	gid, err := envutils.GetGid(constants.EdgeUserGroup)
	if err != nil {
		return err
	}
	syscallStat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("unsupported operate system")
	}
	if !((syscallStat.Uid == 0 && syscallStat.Gid == 0) ||
		(syscallStat.Uid == uid) && (syscallStat.Gid == gid)) {
		return errors.New("bad file owner")
	}
	if (stat.Mode() & groupOrOtherWrite) != 0 {
		return errors.New("bad file permission")
	}
	realPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return err
	}
	if realPath != filePath {
		return errors.New("symlink is not allowed")
	}
	return nil
}
