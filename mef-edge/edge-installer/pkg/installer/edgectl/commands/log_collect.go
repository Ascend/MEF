// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands
package commands

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

// ErrNotSupported custom errors
var ErrNotSupported = errors.New("container log collection is not supported yet")

type logCollectCmd struct {
	tarGzPath string
	module    string
}

// LogCollectCmd edge control command log collect
func LogCollectCmd() common.Command {
	return &logCollectCmd{}
}

// Name command name
func (cmd *logCollectCmd) Name() string {
	return common.CollectLog
}

// Description command description
func (cmd *logCollectCmd) Description() string {
	return common.CollectLogDesc
}

// BindFlag command flag binding
func (cmd *logCollectCmd) BindFlag() bool {
	flag.StringVar(&cmd.tarGzPath, "log_pack_path", "", "the path of MEF Edge log")
	flag.StringVar(&cmd.module, "module", "", "the module of MEF Edge log")
	utils.MarkFlagRequired("log_pack_path")
	utils.MarkFlagRequired("module")
	return true
}

// LockFlag command lock flag
func (cmd *logCollectCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *logCollectCmd) Execute(ctx *common.Context) error {
	if cmd.module != "all" {
		hwlog.RunLog.Errorf("unsupported module parameter: %v", cmd.module)
		if cmd.module == "APP" {
			return ErrNotSupported
		}
		return errors.New("unsupported module parameter")
	}
	return cmd.logCollect()
}

func (cmd *logCollectCmd) logCollect() error {
	hwlog.RunLog.Info("start to collect log ....")
	fmt.Println("start to collect log ....")

	omLogDir, omLogBackupDir, err := path.GetCompLogDirs(constants.EdgeOm)
	if err != nil {
		fmt.Printf("collect log failed: get component log dirs failed, %v\n", err)
		hwlog.RunLog.Errorf("collect log failed: get component log dirs failed, %v", err)
		return errors.New("get log real dirs failed")
	}

	if _, err := fileutils.RealDirCheck(filepath.Dir(cmd.tarGzPath), true, false); err != nil {
		return fmt.Errorf("failed to check dir %s, %v", filepath.Dir(cmd.tarGzPath), err)
	}
	collectPathWhiteList := []string{"/run/collect_log/mef_edge.tar.gz"}
	collector := util.GetLogCollector(cmd.tarGzPath, filepath.Dir(omLogDir),
		filepath.Dir(omLogBackupDir), collectPathWhiteList)
	if _, err := collector.Collect(); err != nil {
		fmt.Printf("collect log failed: %v\n", err)
		hwlog.RunLog.Errorf("collect log failed: %v", err)
		return errors.New("collect log failed")
	}
	if err := fileutils.SetPathPermission(cmd.tarGzPath, constants.Mode400, false, false); err != nil {
		fmt.Printf("collect log failed: chmod failed, %v\n", err)
		hwlog.RunLog.Errorf("collect log failed: chmod failed %v", err)
		return errors.New("chmod failed")
	}

	fmt.Println("collect log success")
	hwlog.RunLog.Infof("collect log success")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *logCollectCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *logCollectCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
