// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"edge-manager/pkg/kubeclient"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/common"
	"edge-manager/pkg/util"
)

// CreateApp Create application
func CreateApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create app")
	var req util.CreateAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	total, err := GetTableCount(AppInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get app table num failed")
		return common.RespMsg{Status: "", Msg: "get app table num failed", Data: nil}
	}
	if total >= MaxApp {
		hwlog.RunLog.Errorf("app number is enough, can not create")
		return common.RespMsg{Status: "", Msg: "app number is enough, can not create", Data: nil}
	}
	app := getAppInfo(req)
	container := getAppContainer(req, app.CreatedAt, app.ModifiedAt)

	if err = AppRepositoryInstance().createApp(app, container); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Errorf("app name is duplicate")
			return common.RespMsg{Status: "", Msg: "app name is duplicate", Data: nil}
		}
		hwlog.RunLog.Errorf("app db create failed")
		return common.RespMsg{Status: "", Msg: "app db create failed", Data: nil}
	}
	hwlog.RunLog.Infof("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getAppInfo(req util.CreateAppReq) *AppInfo {
	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		CreatedAt:   time.Now().Format(common.TimeFormat),
		ModifiedAt:  time.Now().Format(common.TimeFormat),
	}
}

func getAppContainer(req util.CreateAppReq, createdAt, modifiedAt string) *AppContainer {
	return &AppContainer{
		AppName:       req.AppName,
		CreatedAt:     createdAt,
		ModifiedAt:    modifiedAt,
		ContainerName: req.ContainerName,
		CpuRequest:    req.CpuRequest,
		CpuLimit:      req.CpuLimit,
		MemoryRequest: req.MemRequest,
		MemoryLimit:   req.MemLimit,
		Npu:           req.Npu,
		ImageName:     req.ImageName,
		ImageVersion:  req.ImageVersion,
		ContainerPort: req.ContainerPort,
		Command:       getCommand(req.Command),
		Env:           req.Env,
		UserID:        req.UserId,
		GroupID:       req.GroupId,
		HostIp:        req.HostIp,
		HostPort:      req.HostPort,
	}
}

func getCommand(reqCommand []string) string {
	res := ""
	for _, str := range reqCommand {
		res += str + ";"
	}
	return res
}

// ListAppDeployed get appInstances list
func ListAppDeployed(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start list deployed apps")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	apps, err := AppRepositoryInstance().listAppsDeployed(req.PageNum, req.PageSize)
	if err == nil {
		hwlog.RunLog.Infof("list deployed apps success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: apps}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Infof("dont have any deployed apps")
		return common.RespMsg{Status: common.Success, Msg: "dont have any deployed apps", Data: nil}
	}
	hwlog.RunLog.Errorf("list deployed apps failed")
	return common.RespMsg{Status: "", Msg: "list deployed apps failed", Data: nil}
}

// DeployApp deploy application on node group
func DeployApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Infof("start deploy app")
	var req util.DeployAppReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	appInstanceInfo, err := AppRepositoryInstance().getAppAndNodeGroupInfo(req.AppName, req.NodeGroupName)
	if err != nil {
		hwlog.RunLog.Error("get app and node group information failed")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	daemonset, err := kubeclient.GetKubeClient().InitDaemonSet(appInstanceInfo)
	if err != nil {
		hwlog.RunLog.Error("app daemonset init failed")
		return common.RespMsg{Status: "", Msg: "app daemonset init failed", Data: nil}
	}
	daemonset, err = kubeclient.GetKubeClient().CreateDaemonSet(daemonset)
	if err != nil {
		hwlog.RunLog.Error("app daemonset create failed")
		return common.RespMsg{Status: "", Msg: "app daemonset create failed", Data: nil}
	}

	hwlog.RunLog.Infof("app daemonset create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
