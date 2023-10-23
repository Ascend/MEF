// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmonitor this package is for monitor the nginx
package nginxmonitor

import (
	"context"
	"strconv"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"
)

type mStatus int

const (
	splitCount              = 2
	tcpFilePath             = "/proc/net/tcp"
	monitorInterval         = 5 * time.Second
	idle            mStatus = 0
	processing      mStatus = 1
	int64base               = 16
)

var nginxSslPort int

type monitorItem struct {
	monitorType   string
	monitorStatus mStatus
}

var monitorPortItem = monitorItem{monitorType: "port", monitorStatus: idle}

type nginxMonitor struct {
	enable bool
	ctx    context.Context
}

// NewNginxMonitor create NewNginxManager module
func NewNginxMonitor(enable bool, ctx context.Context) model.Module {
	return &nginxMonitor{
		enable: enable,
		ctx:    ctx,
	}
}

// Name module name
func (n *nginxMonitor) Name() string {
	return nginxcom.NginxMonitorName
}

// Enable module enable
func (n *nginxMonitor) Enable() bool {
	sslPort, err := nginxcom.GetEnvManager().GetInt(nginxcom.NginxSslPortKey)
	if err != nil {
		return false
	}
	nginxSslPort = sslPort
	hwlog.RunLog.Infof("%s: %d", nginxcom.NginxSslPortKey, nginxSslPort)
	return n.enable
}

// Start module start
func (n *nginxMonitor) Start() {
	registerHandlers()
	go startMonitor(n.ctx)
	for {
		select {
		case _, ok := <-n.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("nginx monitor service catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("nginx monitor service has listened stop signal")
			return
		default:
		}
		req, err := modulemgr.ReceiveMessage(nginxcom.NginxMonitorName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request failed", nginxcom.NginxMonitorName)
			continue
		}
		msgutil.Handle(req)
	}
}

func registerHandlers() {
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.RespRestartNginx, nginxcom.Nginx), respRestartNginx)
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.ReqActiveMonitor, nginxcom.Monitor), reqActiveMonitor)
}

func startMonitor(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("nginx monitor catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("nginx monitor has listened stop signal")
			return
		default:
		}
		time.Sleep(monitorInterval)
		if monitorPortItem.monitorStatus != processing {
			continue
		}
		isUp := isNginxUp(nginxSslPort)
		if !isUp {
			msgutil.SendVoidMsg(nginxcom.NginxMonitorName, nginxcom.NginxManagerName,
				nginxcom.ReqRestartNginx, nginxcom.Nginx)
			monitorPortItem.monitorStatus = idle
		}
	}
}

func isNginxUp(targetPort int) bool {
	realTcpFilePath, err := fileutils.EvalSymlinks(tcpFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("get real path of %s failed, %v", tcpFilePath, err)
		return false
	}
	data, err := fileutils.LoadFile(realTcpFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("load file %s failed, %v", tcpFilePath, err)
		return false
	}
	lines := strings.Split(string(data), "\n")

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < splitCount {
			continue
		}
		ipPort := strings.Split(fields[1], ":")
		if len(ipPort) != splitCount {
			continue
		}
		port, err := strconv.ParseInt(ipPort[1], int64base, common.BitSize64)
		if err != nil {
			continue
		}
		if int(port) == targetPort {
			return true
		}
	}
	hwlog.RunLog.Errorf("port %d not used, nginx is down", targetPort)
	return false
}

func respRestartNginx(req *model.Message) {
	monitorPortItem.monitorStatus = processing
}

func reqActiveMonitor(req *model.Message) {
	monitorPortItem.monitorStatus = processing
}
