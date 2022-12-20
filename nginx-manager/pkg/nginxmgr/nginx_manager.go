// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"fmt"
	"os"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
	"nginx-manager/pkg/checker"
	"nginx-manager/pkg/nginxcom"
)

const (
	maxRetryTime  = 10
	retryInterval = 10 * time.Second
)

// InitResource 初始化nginx需要的资源
func InitResource() error {
	err := checkEnvs()
	if err != nil {
		return err
	}
	err = updateConf()
	if err != nil {
		return err
	}
	err = loadCerts()
	if err != nil {
		return err
	}
	return nil
}

func loadEnvs() map[string]string {
	envs := make(map[string]string)
	envs[nginxcom.EdgeUrlKey] = os.Getenv(nginxcom.EdgeUrlKey)
	envs[nginxcom.EdgePortKey] = os.Getenv(nginxcom.EdgePortKey)
	envs[nginxcom.SoftUrlKey] = os.Getenv(nginxcom.SoftUrlKey)
	envs[nginxcom.SoftPortKey] = os.Getenv(nginxcom.SoftPortKey)
	return envs
}

func checkEnvs() error {
	envs := loadEnvs()
	return checker.Check(checker.Env, envs)
}

func updateConf() error {
	envs := loadEnvs()
	items := CreateConfItems(envs)
	updater, err := NewNginxConfUpdater(items, nginxcom.NginxDefaultConfigPath)
	if err != nil {
		return err
	}
	return updater.Update()
}

func loadCerts() error {
	loader := NewNginxCertLoader(nginxcom.CertKeyFile, nginxcom.PipePath)
	return loader.Load()
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

type handlerFunc func(req *model.Message)

// NewNginxManager create NewNginxManager module
func NewNginxManager(enable bool) model.Module {
	return &nginxManager{
		enable: enable,
		ctx:    make(chan struct{}),
	}
}

type nginxManager struct {
	enable bool
	ctx    chan struct{}
}

// Name module name
func (n *nginxManager) Name() string {
	return "NginxManager"
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

	err = cmdStart()
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}
	return true
}

func startNginx() bool {
	count := 0
	for {
		success := doStartNginx()
		if success {
			return true
		}
		count++
		if count >= maxRetryTime {
			hwlog.RunLog.Errorf("try start nginx fail exceed %d times, exit program", maxRetryTime)
		}
		time.Sleep(retryInterval)
	}
}

// Start module start
func (n *nginxManager) Start() {
	hwlog.RunLog.Error("try start nginx ")
	if !startNginx() {
		return
	}
	for {
		select {
		case <-n.ctx:
			return
		default:
		}
		req, err := modulemanager.ReceiveMessage(nginxcom.NginxManagerName)
		hwlog.RunLog.Debugf("%s receive request from software manager", nginxcom.NginxManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from software manager failed", nginxcom.NginxManagerName)
			continue
		}
		dispatch(req)
	}
}

func dispatch(req *model.Message) {
	method, exit := nodeMethodList()[combine(req.GetOption(), req.GetResource())]
	if !exit {
		return
	}
	method(req)
}

func nodeMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
