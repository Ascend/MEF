// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmonitor this package is for monitor the nginx
package nginxmonitor

import (
	"context"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

const (
	nginxUsedPort   = 8443
	splitCount      = 2
	tcpFilePath     = "/proc/net/tcp"
	monitorInterval = 5 * time.Second
)

type mStatus int

const (
	idle       mStatus = 0
	processing mStatus = 1
)

type monitorItem struct {
	monitorType   string
	monitorStatus mStatus
	monitorTarget int64
}

var monitorPortItem = monitorItem{monitorType: "port", monitorStatus: idle, monitorTarget: nginxUsedPort}

type nginxMonitor struct {
	enable bool
	ctx    chan struct{}
}

// NewNginxMonitor create NewNginxManager module
func NewNginxMonitor(enable bool) model.Module {
	return &nginxMonitor{
		enable: enable,
		ctx:    make(chan struct{}),
	}
}

// Name module name
func (n *nginxMonitor) Name() string {
	return nginxcom.NginxMonitorName
}

// Enable module enable
func (n *nginxMonitor) Enable() bool {
	return n.enable
}

// Start module start
func (n *nginxMonitor) Start() {
	registerHandlers()
	go startMonitor()
	for {
		select {
		case <-n.ctx:
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(nginxcom.NginxMonitorName)
		hwlog.RunLog.Infof("%s receive request from software manager", nginxcom.NginxMonitorName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from software manager failed", nginxcom.NginxMonitorName)
			continue
		}
		msgutil.Handle(req)
	}
}

func registerHandlers() {
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.RespRestartNginx, nginxcom.Nginx), respRestartNginx)
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.ReqActiveMonitor, nginxcom.Monitor), reqActiveMonitor)
}

func startMonitor() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(monitorInterval)
		if monitorPortItem.monitorStatus != processing {
			continue
		}
		isUp := isNginxUp(monitorPortItem.monitorTarget)
		if !isUp {
			msgutil.SendVoidMsg(nginxcom.NginxMonitorName, nginxcom.NginxManagerName,
				nginxcom.ReqRestartNginx, nginxcom.Nginx)
			monitorPortItem.monitorStatus = idle
		}
	}
}

func isNginxUp(targetPort int64) bool {
	data, err := ioutil.ReadFile(tcpFilePath)
	if err != nil {
		hwlog.RunLog.Error(err)
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
		port, err := strconv.ParseInt(ipPort[1], 16, 64)
		if err != nil {
			continue
		}
		if port == targetPort {
			return true
		}
	}
	hwlog.RunLog.Errorf("port %d not used, nginx is down", nginxUsedPort)
	return false
}

func respRestartNginx(req *model.Message) {
	monitorPortItem.monitorStatus = processing
}

func reqActiveMonitor(req *model.Message) {
	monitorPortItem.monitorStatus = processing
}
