// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package commands

import (
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/domainconfig"
)

type domainConfigCmd struct {
	domain string
	ip     string
}

// DomainConfigCmd  edge control command domain mapping config
func DomainConfigCmd() common.Command {
	return &domainConfigCmd{}
}

func (cmd *domainConfigCmd) Name() string {
	return common.DomainConfig
}

func (cmd *domainConfigCmd) Description() string {
	return common.DomainConfigDesc
}

// BindFlag command flag binding
func (cmd *domainConfigCmd) BindFlag() bool {
	flag.StringVar(&cmd.domain, "domain", "", "the domain(server_name) of image registry")
	flag.StringVar(&cmd.ip, "ip", "", "the ip of image registry")
	utils.MarkFlagRequired("domain")
	utils.MarkFlagRequired("ip")
	return true
}

func (cmd *domainConfigCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *domainConfigCmd) Execute(ctx *common.Context) error {

	if err := common.InitEdgeOmResource(); err != nil {
		return fmt.Errorf("init resource failed, error: %v", err)
	}

	flow := domainconfig.NewDomainCfgFlow(cmd.domain, cmd.ip)
	if err := flow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("domain mapping config %s failed, error: %v", cmd.domain, err)
		return fmt.Errorf("domain mapping config %s failed, error: %v", cmd.domain, err)
	}
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *domainConfigCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] config domain %s mapping success", user, ip, cmd.domain)
}

// PrintOpLogFail print operation fail log
func (cmd *domainConfigCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] config domain %s mapping failed", user, ip, cmd.domain)
}
