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
	"path/filepath"
	"sort"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/commands"
	"edge-installer/pkg/installer/edgectl/common"
)

const (
	ctlArgIndex         = 1
	cmdIndex            = 1
	cmdArgIndex         = 2
	defLockedCode       = 5
	defNotSupportedCode = 2
	defErrCode          = 1
	defOkCode           = 0
)

var (
	// BuildVersion the program version
	BuildVersion string

	h       bool
	help    bool
	version bool

	cmdMap         = make(map[string]common.Command)
	curCmd         common.Command
	installRootDir string
)

func handle() int {
	if err := setRootDir(); err != nil {
		fmt.Printf("current directory is invalid, error: %v\n", err)
		return defErrCode
	}
	initCmd()
	if !dealArgs() {
		return defErrCode
	}
	if err := initLog(); err != nil {
		fmt.Println(err)
		return defErrCode
	}

	curUser, ip, err := envutils.GetUserAndIP()
	if err != nil {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		hwlog.RunLog.Errorf("get current user or ip info failed, error: %s", err.Error())
		return defErrCode
	}

	if curUser != constants.RootUserName {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		hwlog.RunLog.Errorf("the current user must be root, can not be %s", curUser)
		return defErrCode
	}

	hwlog.RunLog.Infof("command [%s], start", curCmd.Name())
	hwlog.OpLog.Infof("[%s@%s] command [%s] start", curUser, ip, curCmd.Name())

	if err = checkCurPath(); err != nil {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		curCmd.PrintOpLogFail(curUser, ip)
		return defErrCode
	}

	code, err := process()
	if err != nil {
		fmt.Printf("Execute [%s] command failed.\n", curCmd.Name())
		hwlog.RunLog.Errorf("command: [%s], result: failed, error: %v", curCmd.Name(), err)
		curCmd.PrintOpLogFail(curUser, ip)
		return code
	}

	fmt.Printf("Execute [%s] command success!\n", curCmd.Name())
	hwlog.RunLog.Infof("command: [%s], result: success", curCmd.Name())
	curCmd.PrintOpLogOk(curUser, ip)
	return code
}

func main() {
	os.Exit(handle())
}

func checkCurPath() error {
	curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		hwlog.RunLog.Errorf("get current abs path failed: %s", err.Error())
		return err
	}

	curAbsPath, err := filepath.EvalSymlinks(curPath)
	if err != nil {
		hwlog.RunLog.Errorf("get current abs path failed: %s", err.Error())
		return err
	}

	workAbsDir, err := filepath.EvalSymlinks(pathmgr.NewWorkPathMgr(installRootDir).GetWorkDir())
	if err != nil {
		hwlog.RunLog.Errorf("get softlink abs path failed: %s", err.Error())
		return err
	}

	if !strings.HasPrefix(curAbsPath, workAbsDir) {
		fmt.Println("current sh path is not in the working path")
		hwlog.RunLog.Error("current sh path is not in the working path")
		return errors.New("current sh path is not in the working path")
	}

	return nil
}

func initCmd() {
	registerCmd(commands.StartCmd())
	registerCmd(commands.StopCmd())
	registerCmd(commands.RestartCmd())
	registerCmd(commands.UninstallCmd())
	registerCmd(commands.UpgradeCmd())
	registerCmd(commands.GetNetCfgCmd())
	registerCmd(commands.EffectCmd())
	registerCmd(commands.LogCollectCmd())
	registerCmd(commands.UpdateKmcCmd())
	registerCmd(commands.UpdateCrlCmd())
	initCmdExt()
}

func registerCmd(cmd common.Command) {
	if _, ok := cmdMap[cmd.Name()]; !ok {
		cmdMap[cmd.Name()] = cmd
	}
}

func dealEdgeCtlFlag() bool {
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&h, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.Parse()
	if help || h {
		printUsage()
		return false
	}
	if version {
		printVersion()
		return false
	}
	printUseHelp()
	return false
}

func dealCmdFlag() bool {
	cmd, ok := cmdMap[os.Args[cmdIndex]]
	if !ok {
		fmt.Println("the parameter is invalid")
		printUseHelp()
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
	if utils.IsRequiredFlagNotFound() {
		fmt.Println("the required parameter is missing")
		flag.PrintDefaults()
		return false
	}
	return true
}

func dealArgs() bool {
	flag.Usage = printUseHelp
	if len(os.Args) <= constants.MinArgsLen {
		printUseHelp()
		return false
	}
	if len(os.Args[ctlArgIndex]) == 0 {
		fmt.Println("the required parameter is missing")
		return false
	}
	if os.Args[ctlArgIndex][0] == '-' {
		return dealEdgeCtlFlag()
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
			return defLockedCode, err
		}
		defer func() {
			if err := common.UnlockProcessFlag(constants.FlagPath, curCmd.Name()); err != nil {
				hwlog.RunLog.Warn(err.Error())
				return
			}
		}()
	}

	if realCmd, ok := curCmd.(common.CommandWithRetCode); ok {
		return realCmd.ExecuteWithCode(ctx)
	}
	err := curCmd.Execute(ctx)
	if err != nil {
		if errors.Is(err, commands.ErrNotSupported) {
			return defNotSupportedCode, err
		} else {
			return defErrCode, err
		}
	}
	return defOkCode, nil
}

func printVersion() {
	fmt.Printf("%s\n", BuildVersion)
}

func printUseHelp() {
	fmt.Println("use '-help' for help information")
}

func printUsage() {
	descriptions := make([]string, 0, len(cmdMap))
	for _, cmd := range cmdMap {
		descriptions = append(descriptions, fmt.Sprintf("\t%-10s\t%s", cmd.Name(), cmd.Description()))
	}
	sort.Strings(descriptions)
	fmt.Printf(`Usage: [OPTIONS...] COMMAND

Options:
	-help		Print help information
	-version	Print version information

Commands:
%s
`, strings.Join(descriptions, "\n"))
}
