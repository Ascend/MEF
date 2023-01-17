// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulmgr to restful deal
package restfulmgr

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

var allModuleDispatchers map[string]DispatcherItf
var dispatcherLock sync.RWMutex

// DispatcherItf [interface] for Dispatch massage
type DispatcherItf interface {
	ParseData(c *gin.Context) common.Result
	dispatch(c *gin.Context)
	getMethod() string
	getRelativePath() string
}

// GenericDispatcher [struct] to deal message Dispatch
type GenericDispatcher struct {
	RelativePath string
	Method       string
	Destination  string
}

// ParseData [method] parse url data
func (g GenericDispatcher) ParseData(c *gin.Context) common.Result {
	data, err := c.GetRawData()
	if err != nil {
		return common.Result{ResultFlag: false, ErrorMsg: "get input parameter failed"}
	}

	return common.Result{ResultFlag: true, Data: string(data)}
}

// sendToModule [method] send to other module
func (g GenericDispatcher) sendToModule(resource string, data interface{}) common.Result {
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: g.Destination,
		Option:      g.getMethod(),
		Resource:    resource,
	}

	resp := common.SendSyncMessageByRestful(data, &router)
	return common.Result{ResultFlag: resp.Status == common.Success, Data: resp.Data, ErrorMsg: resp.Msg}
}

func (g GenericDispatcher) response(c *gin.Context, errorCode string, msg string, Data interface{}) {
	if msg == "" {
		msg = common.ErrorMap[errorCode]
	}
	result := common.RespMsg{
		Status: errorCode,
		Msg:    msg,
		Data:   Data,
	}
	c.JSON(http.StatusOK, result)
}

func (g GenericDispatcher) getMethod() string {
	return g.Method
}

func (g GenericDispatcher) getRelativePath() string {
	return g.RelativePath
}

func (g GenericDispatcher) dispatch(c *gin.Context) {
	var res common.Result
	defer func() {
		hwlog.RunLog.Infof("deal %s result is %v", c.FullPath(), res.ResultFlag)

		if g.getMethod() == http.MethodGet {
			return
		}
		if !res.ResultFlag {
			hwlog.OpLog.Errorf("%s %s %s %s failed\n", c.ClientIP(), c.Request.Header["user"],
				g.getMethod(), c.FullPath())
			return
		}
		hwlog.OpLog.Infof("%s %s %s %s success\n", c.ClientIP(), c.Request.Header["user"],
			g.getMethod(), c.FullPath())
	}()

	dispatcher, ok := allModuleDispatchers[combine(c.FullPath(), c.Request.Method)]
	if !ok {
		res = common.Result{ResultFlag: false, ErrorMsg: "get dispatcher failed"}
		g.response(c, common.ErrorParamInvalid, res.ErrorMsg, nil)
		return
	}
	res = dispatcher.ParseData(c)
	if !res.ResultFlag {
		g.response(c, common.ErrorParseBody, res.ErrorMsg, nil)
		return
	}

	res = g.sendToModule(c.FullPath(), res.Data)
	if !res.ResultFlag {
		g.response(c, common.ErrorsSendSyncMessageByRestful, res.ErrorMsg, nil)
		return
	}

	g.response(c, common.Success, "message deal success", res.Data)
	return
}

func combine(fulPath, method string) string {
	return fmt.Sprintf("%s%s", fulPath, method)
}

// InitRouter [method] for init router handler
func InitRouter(engine *gin.Engine, urlDispatchers map[string][]DispatcherItf) {
	dispatcherLock.Lock()
	defer dispatcherLock.Unlock()

	if allModuleDispatchers == nil {
		allModuleDispatchers = make(map[string]DispatcherItf)
	}
	for k, dispatchers := range urlDispatchers {
		group := engine.Group(k)
		for _, dispatcher := range dispatchers {
			switch dispatcher.getMethod() {
			case http.MethodPost:
				group.POST(dispatcher.getRelativePath(), dispatcher.dispatch)
			case http.MethodGet:
				group.GET(dispatcher.getRelativePath(), dispatcher.dispatch)
			case http.MethodPatch:
				group.PATCH(dispatcher.getRelativePath(), dispatcher.dispatch)
			case http.MethodDelete:
				group.DELETE(dispatcher.getRelativePath(), dispatcher.dispatch)
			default:
				hwlog.RunLog.Errorf("url method is not supported")
				continue
			}
			allModuleDispatchers[combine(
				filepath.Join(k, dispatcher.getRelativePath()),
				dispatcher.getMethod())] = dispatcher
		}
	}
}
