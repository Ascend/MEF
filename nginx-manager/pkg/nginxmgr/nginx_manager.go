// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"context"
	"os/exec"
	"time"

	"nginx-manager/pkg/checker"
	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

const (
	retryInterval = 15 * time.Second
	restyBinPath  = "/usr/bin/openresty"
	restyPrefix   = "/home/MEFCenter/"
)

// InitResource 初始化nginx需要的资源
func InitResource() error {
	err := updateConf()

	if err != nil {
		return err
	}
	err = prepareServerCert()
	if err != nil {
		return err
	}
	err = loadCerts()
	if err != nil {
		return err
	}
	return nil
}

func updateConf() error {
	items := CreateConfItems(nginxcom.Envs)
	updater := NewNginxConfUpdater(items, nginxcom.NginxDefaultConfigPath)
	return updater.Update()
}

func loadCerts() error {
	err := Load(nginxcom.ServerCertKeyFile, nginxcom.PipePath)
	if err != nil {
		return err
	}
	updater := NewNginxConfUpdater(nil, nginxcom.NginxDefaultConfigPath)
	pipeCount, err := updater.calculatePipeCount()
	if err != nil {
		return err
	}
	err = LoadForClient(nginxcom.ClientCertKeyFile, nginxcom.ClientPipeDir, pipeCount)
	return err
}

// CreateConfItems 创建nginx.conf配置文件的替换项
func CreateConfItems(envs map[string]string) []nginxcom.NginxConfItem {
	var ret []nginxcom.NginxConfItem
	template := checker.GetConfigItemTemplate()
	for _, item := range template {
		createdItem := nginxcom.NginxConfItem{
			Key:  item.Key,
			From: item.From,
			To:   item.From + " " + envs[item.Key],
		}
		ret = append(ret, createdItem)
	}
	return ret
}

// NewNginxManager create NewNginxManager module
func NewNginxManager(enable bool, ctx context.Context) model.Module {
	return &nginxManager{
		enable: enable,
		ctx:    ctx,
	}
}

type nginxManager struct {
	enable bool
	ctx    context.Context
}

// Name module name
func (n *nginxManager) Name() string {
	return nginxcom.NginxManagerName
}

// Enable module enable
func (n *nginxManager) Enable() bool {
	return n.enable
}

func doStartNginx() bool {
	err := InitResource()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}

	return startResty()
}

func startNginx() {
	count := 0
	for {
		success := doStartNginx()
		if success {
			return
		}
		count++
		hwlog.RunLog.Errorf("start nginx fail exceed %d times", count)
		time.Sleep(retryInterval)
	}
}

// Start module start
func (n *nginxManager) Start() {
	startNginx()
	registerHandlers()
	msgutil.SendVoidMsg(nginxcom.NginxManagerName, nginxcom.NginxMonitorName,
		nginxcom.ReqActiveMonitor, nginxcom.Monitor)
	for {
		select {
		case _, ok := <-n.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("nginx manager catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("nginx manager has listened stop signal")
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(nginxcom.NginxManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request failed", nginxcom.NginxManagerName)
			continue
		}
		msgutil.Handle(req)
	}
}

func registerHandlers() {
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.ReqRestartNginx, nginxcom.Nginx), reqRestartNginx)
}

func reqRestartNginx(req *model.Message) {
	startNginx()
	msgutil.SendVoidMsg(nginxcom.NginxManagerName, nginxcom.NginxMonitorName, nginxcom.RespRestartNginx, nginxcom.Nginx)
}

func startResty() bool {
	cmd := exec.Command(restyBinPath, "-p", restyPrefix)
	_, err := cmd.CombinedOutput()
	if err != nil {
		hwlog.RunLog.Errorf("run openresty failed: %s", err.Error())
		return false
	}
	hwlog.RunLog.Info("run openresty success")
	return true
}
