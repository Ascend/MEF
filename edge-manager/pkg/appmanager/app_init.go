package appmanager

import (
	"context"
	"fmt"

	"edge-manager/module_manager"
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"edge-manager/pkg/database"

	"huawei.com/mindx/common/hwlog"
)

type handlerFunc func(req interface{}) common.RespMsg

type appManager struct {
	enable bool
	ctx    context.Context
}

// NewAppManager create app manager
func NewAppManager(enable bool) *appManager {
	am := &appManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return am
}

func (app *appManager) Name() string {
	return common.AppManagerName
}

func (app *appManager) Enable() bool {
	return app.enable
}

func (app *appManager) Start() {
	if err := database.CreateTableIfNotExists(AppInfo{}); err != nil {
		hwlog.RunLog.Error("create app database table failed")
		return
	}
	if err := database.CreateTableIfNotExists(AppContainer{}); err != nil {
		hwlog.RunLog.Error("create app group database table failed")
		return
	}
	for {
		select {
		case _, ok := <-app.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}
		req, err := module_manager.ReceiveMessage(common.NodeManagerName)
		hwlog.RunLog.Debugf("%s revice requst from restful service", common.NodeManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s revice requst from restful service failed", common.NodeManagerName)
			continue
		}
		msg := methodSelect(req)
		if msg == nil {
			hwlog.RunLog.Error("%s get method by option and resource failed", common.NodeManagerName)
			continue
		}
		resp, err := req.NewResponse()
		if err != nil {
			hwlog.RunLog.Error("%s new response failed", common.NodeManagerName)
			continue
		}
		resp.FillContent(msg)
		if err = module_manager.SendMessage(resp); err != nil {
			hwlog.RunLog.Error("%s send response failed", common.NodeManagerName)
			continue
		}
	}
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := appMethodList()[combine(req.GetOption(), req.GetResource())]
	if !exit {
		return nil
	}
	res = method(req.GetContent())
	return &res
}

func appMethodList() map[string]handlerFunc {
	return map[string]handlerFunc{
		combine(common.Create, common.App): CreateApp,
		//combine(common.Delete, common.App):   DeleteApp,
		//combine(common.Get, common.App):      GetApp,
		//combine(common.Deploy, common.App):   DeployApp,
		//combine(common.Undeploy, common.App): UndeployApp,
	}
}

func combine(option, resource string) string {
	return fmt.Sprintf("%s%s", option, resource)
}
