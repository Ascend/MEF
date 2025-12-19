// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package commands this file for edge control command net config
package commands

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	FlowCommon "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/netconfig"
)

type netConfigCmd struct {
	netType     string
	ip          string
	port        int
	authPort    int
	rootCa      string
	token       string
	testConnect bool
}

// NetConfigCmd edge control command net config
func NetConfigCmd() common.Command {
	return &netConfigCmd{}
}

// Name command name
func (cmd *netConfigCmd) Name() string {
	return common.NetConfig
}

// Description command description
func (cmd *netConfigCmd) Description() string {
	return common.NetConfigDesc
}

// BindFlag command flag binding
func (cmd *netConfigCmd) BindFlag() bool {
	flag.StringVar(&(cmd.netType), "net_type", constants.MEF, "the type of net manager")
	flag.StringVar(&(cmd.ip), "ip", "", "the ip of MEF Center")
	flag.IntVar(&(cmd.port), "port", constants.DefaultWsPort, "the port of MEF Center")
	flag.IntVar(&(cmd.authPort), "auth_port", constants.DefaultWsTestPort, "the auth port of MEF Center")
	flag.StringVar(&(cmd.rootCa), "root_ca", "", "the root ca of MEF Center")
	flag.BoolVar(&(cmd.testConnect), "test_connect", true, "whether to test the connection between MEF Edge and Center")
	return true
}

// LockFlag command lock flag
func (cmd *netConfigCmd) LockFlag() bool {
	return true
}

func (cmd *netConfigCmd) getConfigFlow(ctx *common.Context) FlowCommon.Flow {
	switch cmd.netType {
	case constants.MEF:
		param := netconfig.Param{
			NetType:       cmd.netType,
			Ip:            cmd.ip,
			Port:          cmd.port,
			AuthPort:      cmd.authPort,
			RootCa:        cmd.rootCa,
			TestConnect:   cmd.testConnect,
			ConfigPathMgr: ctx.ConfigPathMgr,
		}
		return netconfig.NewMefConfigFlow(param)
	default:
		return nil
	}
}

// Execute execute command
func (cmd *netConfigCmd) Execute(ctx *common.Context) error {
	hwlog.RunLog.Info("start to net config")
	fmt.Println("start to net config...")

	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	if err := cmd.necessaryParamCheck(); err != nil {
		fmt.Printf("check net config param failed: %s\n", err.Error())
		flag.PrintDefaults()
		return fmt.Errorf("net config failed, error: %v", err)
	}

	if err := common.InitEdgeOmResource(); err != nil {
		return fmt.Errorf("init resource failed, error: %v", err)
	}

	flow := cmd.getConfigFlow(ctx)
	if flow == nil {
		hwlog.RunLog.Errorf("get %s config flow failed", cmd.netType)
		return fmt.Errorf("get %s config flow failed", cmd.netType)
	}
	if err := flow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("net config %s failed, error: %v", cmd.netType, err)
		return fmt.Errorf("net config %s failed, error: %v", cmd.netType, err)
	}
	hwlog.RunLog.Info("net config success")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *netConfigCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] config net %s success", user, ip, cmd.netType)
}

// PrintOpLogFail print operation fail log
func (cmd *netConfigCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] config net %s failed", user, ip, cmd.netType)
}

func (cmd *netConfigCmd) necessaryParamCheck() error {
	netType := flag.Lookup("net_type").Value.String()

	if netType != constants.MEF {
		hwlog.RunLog.Errorf("type of net manager error, only support [%s]", constants.MEF)
		return fmt.Errorf("type of net manager only support [%s]", constants.MEF)
	}

	if flag.Lookup("ip").Value.String() == "" {
		hwlog.RunLog.Error("param ip is necessary, can not be empty")
		return errors.New("param ip is necessary, can not be empty")
	}

	if flag.Lookup("root_ca").Value.String() == "" {
		hwlog.RunLog.Error("param root_ca is necessary, can not be empty")
		return errors.New("param root_ca is necessary, can not be empty")
	}

	return nil
}
