// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package main this file for edge control main
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/innercommands"
)

const (
	cmdIndex    = 1
	cmdArgIndex = 2

	okCode            = 0
	errorExitCode     = 1
	wrongArgsCode     = 2
	getRootDirCode    = 3
	initLogFailCode   = 4
	processLockedCode = 5
)

var (
	cmdMap         = make(map[string]common.Command)
	curCmd         common.Command
	installRootDir string
)

func main() {
	if err := setRootDir(); err != nil {
		fmt.Printf("current directory is invalid, error: %v\n", err)
		os.Exit(getRootDirCode)
	}
	initCmd()
	if !dealArgs() {
		os.Exit(wrongArgsCode)
	}
	if err := initLog(); err != nil {
		fmt.Println(err)
		os.Exit(initLogFailCode)
	}

	curUser, ip, err := envutils.GetUserAndIP()
	if err != nil {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		hwlog.RunLog.Errorf("get current user or ip info failed, error: %s", err.Error())
		os.Exit(errorExitCode)
	}

	hwlog.RunLog.Infof("command [%s],start", curCmd.Name())
	hwlog.OpLog.Infof("[%s@%s] command [%s] start", curUser, ip, curCmd.Name())
	code, err := process()
	if err != nil {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		hwlog.RunLog.Errorf("command: [%s], result: failed, error: %v", curCmd.Name(), err)
		curCmd.PrintOpLogFail(curUser, ip)
		os.Exit(code)
	}

	fmt.Printf("Execute [%s] command success!\n", curCmd.Name())
	hwlog.RunLog.Infof("command: [%s], result: success", curCmd.Name())
	curCmd.PrintOpLogOk(curUser, ip)
}

func initCmd() {
	registerCmd(innercommands.PrepareEdgecoreCmd())
	registerCmd(innercommands.NewRecoverLogCmd())
	initCmdExt()
}

func registerCmd(cmd common.Command) {
	if _, ok := cmdMap[cmd.Name()]; !ok {
		cmdMap[cmd.Name()] = cmd
	}
}

func dealCmdFlag() bool {
	cmd, ok := cmdMap[os.Args[cmdIndex]]
	if !ok {
		fmt.Println("the parameter is invalid")
		printParamErr()
		return false
	}
	curCmd = cmd
	if !curCmd.BindFlag() {
		return true
	}
	flag.Usage = flag.PrintDefaults
	if err := flag.CommandLine.Parse(os.Args[cmdArgIndex:]); err != nil {
		fmt.Printf("parse cmd args failed,error:%v", err)
		return false
	}
	return true
}

func dealArgs() bool {
	flag.Usage = printParamErr
	if len(os.Args) <= constants.MinArgsLen {
		printParamErr()
		return false
	}
	return dealCmdFlag()
}

func setRootDir() error {
	dir, err := path.GetInstallRootDir()
	if err != nil {
		return errors.New("get install root dir failed")
	}
	installRootDir = dir
	return nil
}

func initLog() error {
	if err := util.InitComponentLog(constants.EdgeInstaller); err != nil {
		return fmt.Errorf("initialize log failed, error: %v", err)
	}
	return nil
}

func process() (int, error) {
	ctx := &common.Context{
		WorkPathMgr:   pathmgr.NewWorkPathMgr(installRootDir),
		ConfigPathMgr: pathmgr.NewConfigPathMgr(installRootDir),
		Args:          os.Args[cmdArgIndex:],
	}

	if curCmd.LockFlag() {
		if err := common.LockProcessFlag(constants.FlagPath, curCmd.Name()); err != nil {
			return processLockedCode, err
		}
		defer func() {
			if err := common.UnlockProcessFlag(constants.FlagPath, curCmd.Name()); err != nil {
				hwlog.RunLog.Warn(err.Error())
				return
			}
		}()
	}
	if err := curCmd.Execute(ctx); err != nil {
		return errorExitCode, err
	}
	return okCode, nil
}

func printParamErr() {
	fmt.Println("invalid param count")
}
