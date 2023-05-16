// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"context"
	"os"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"
)

const (
	retryInterval = 15 * time.Second
	startCommand  = "./nginx"
	accessLogFile = "/home/MEFCenter/logs/access.log"
	errorLogFile  = "/home/MEFCenter/logs/error.log"
	logFileMode   = 0600
)

// InitResource initial the resources needed by nginx
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
	items, err := CreateConfItems()
	if err != nil {
		return err
	}
	updater, err := NewNginxConfUpdater(items)
	if err != nil {
		return err
	}
	return updater.Update()
}

func loadCerts() error {
	err := Load(nginxcom.ServerCertKeyFile, nginxcom.PipePath)
	if err != nil {
		return err
	}
	updater, err := NewNginxConfUpdater(nil)
	if err != nil {
		return err
	}
	pipeCount, err := updater.calculatePipeCount()
	if err != nil {
		return err
	}
	err = LoadForClient(nginxcom.ClientCertKeyFile, nginxcom.ClientPipeDir, pipeCount)
	return err
}

// CreateConfItems create some items which used to replace into nginx.conf file
func CreateConfItems() ([]nginxcom.NginxConfItem, error) {
	var ret []nginxcom.NginxConfItem
	template := nginxcom.GetConfigItemTemplate()
	for _, item := range template {
		toVal, err := nginxcom.GetEnvManager().Get(item.Key)
		if err != nil {
			return nil, err
		}
		createdItem := nginxcom.NginxConfItem{
			Key:  item.Key,
			From: item.From,
			To:   toVal,
		}
		ret = append(ret, createdItem)
	}
	return ret, nil
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

	return startNginxCmd()
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

func startNginxCmd() bool {
	_, err := common.RunCommand(startCommand, true, common.DefCmdTimeoutSec)
	if err != nil {
		hwlog.RunLog.Errorf("start nginx failed:%s", err.Error())
		return false
	}
	if err = os.Chmod(accessLogFile, logFileMode); err != nil {
		hwlog.RunLog.Errorf("chmod access.log failed, cause by: {%s}", err.Error())
		return false
	}
	if err = os.Chmod(errorLogFile, logFileMode); err != nil {
		hwlog.RunLog.Errorf("chmod error.log failed, cause by: {%v}", err.Error())
		return false
	}
	hwlog.RunLog.Info("run nginx success")
	return true
}
