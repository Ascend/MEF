// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"context"
	"os"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"

	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"
)

const (
	retryInterval        = 3 * time.Second
	maxGetNorthCertTimes = 15
	startCommand         = "./nginx"
	accessLogFile        = "/home/MEFCenter/logs/access.log"
	errorLogFile         = "/home/MEFCenter/logs/error.log"
	logFileMode          = 0600
)

// initResource initial the resources needed by nginx
func initResource() error {
	if err := updateConf(); err != nil {
		return err
	}

	if err := prepareServerCert(); err != nil {
		return err
	}

	if err := loadCerts(); err != nil {
		return err
	}

	// remove old 3rd north ca
	if err := utils.DeleteFile(nginxcom.NorthernCertFile); err != nil {
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

func startNginx() {
	if err := initResource(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	for {
		for !utils.IsExist(nginxcom.NorthernCertFile) {
			hwlog.RunLog.Info("start to get north ca from cert manager")
			if err := getNorthCert(); err != nil {
				hwlog.RunLog.Errorf("get north ca from cert manager failed: %s", err.Error())
				continue
			}
			hwlog.RunLog.Info("get north ca from cert manager success")
			break
		}
		if startNginxCmd() {
			return
		}
		hwlog.RunLog.Error("start nginx failed")
		time.Sleep(retryInterval)
	}
}

func getNorthCert() error {
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: nginxcom.RootCaPath,
			CertPath:   nginxcom.ClientCertFile,
			KeyPath:    nginxcom.ClientCertKeyFile,
			SvrFlag:    false,
		},
	}
	var caStr string
	var err error
	for i := 0; i < maxGetNorthCertTimes; i++ {
		caStr, err = reqCertParams.GetRootCa(common.NorthernCertName)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}
		if err = utils.WriteData(nginxcom.NorthernCertFile, []byte(caStr)); err != nil {
			continue
		}
		return nil
	}
	return err
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
		req, err := modulemgr.ReceiveMessage(nginxcom.NginxManagerName)
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
