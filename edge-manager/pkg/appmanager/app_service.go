package appmanager

import (
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/common"
	"edge-manager/pkg/util"
)

// CreateApp Create application
func CreateApp(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create app")
	req, ok := input.(util.CreateAppReq)
	if !ok {
		hwlog.RunLog.Error("create app convert request error")
		return common.RespMsg{Status: "", Msg: "convert request error", Data: nil}
	}
	total, err := GetTableCount(AppInfo{})
	if err != nil {
		hwlog.RunLog.Error("get app table num failed")
		return common.RespMsg{Status: "", Msg: "get app table num failed", Data: nil}
	}
	if total >= MaxApp {
		hwlog.RunLog.Error("app number is enough, can not create")
		return common.RespMsg{Status: "", Msg: "app number is enough, can not create", Data: nil}
	}
	app := getAppInfo(req)
	container := getAppContainer(req, app.CreatedAt, app.ModifiedAt)

	if err = AppRepositoryInstance().CreateApp(app, container); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("app name is duplicate")
			return common.RespMsg{Status: "", Msg: "app name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("app db create failed")
		return common.RespMsg{Status: "", Msg: "db create failed", Data: nil}
	}
	hwlog.RunLog.Info("app db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func getAppInfo(req util.CreateAppReq) *AppInfo {
	return &AppInfo{
		AppName:     req.AppName,
		Description: req.Description,
		CreatedAt:   time.Now().Format(TimeFormat),
		ModifiedAt:  time.Now().Format(TimeFormat),
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
		Command:       req.Command,
		Env:           req.Env,
		ContainerHost: HostAddr{req.HostIp, req.HostPort},
		ContainerUser: UserInfo{req.UserId, req.GroupId},
	}
}

//func DeployApp(input interface{}) common.RespMsg {
//
//}
//
//func UndeployApp() {
//
//}
//
//func GetApp() {
//
//}
//
//func DeleteApp() {
//
//}
