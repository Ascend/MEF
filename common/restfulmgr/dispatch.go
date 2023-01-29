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
	ParseData(c *gin.Context) (interface{}, error)
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
func (g GenericDispatcher) ParseData(c *gin.Context) (interface{}, error) {
	data, err := c.GetRawData()
	if err != nil {
		return "", fmt.Errorf("get input parameter failed")
	}

	return string(data), nil
}

// sendToModule [method] send to other module
func (g GenericDispatcher) sendToModule(resource string, data interface{}) common.RespMsg {
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: g.Destination,
		Option:      g.getMethod(),
		Resource:    resource,
	}

	return common.SendSyncMessageByRestful(data, &router)
}

func (g GenericDispatcher) response(c *gin.Context, result common.RespMsg) {
	if result.Msg == "" {
		result.Msg = common.ErrorMap[result.Status]
	}

	if result.Status != common.Success {
		c.JSON(http.StatusBadRequest, result)
		return
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
	var res common.RespMsg
	defer func() {
		hwlog.RunLog.Infof("deal %s result is %v", c.FullPath(), res.Status == common.Success)

		if g.getMethod() == http.MethodGet {
			return
		}
		if res.Status != common.Success {
			hwlog.OpLog.Errorf("%s %s %s %s failed\n", c.ClientIP(), c.Request.Header["user"],
				g.getMethod(), c.FullPath())
			return
		}
		hwlog.OpLog.Infof("%s %s %s %s success\n", c.ClientIP(), c.Request.Header["user"],
			g.getMethod(), c.FullPath())
	}()

	dispatcher, ok := allModuleDispatchers[combine(c.FullPath(), c.Request.Method)]
	if !ok {
		res = common.RespMsg{Status: common.ErrorParamInvalid, Msg: "get dispatcher failed", Data: nil}
		g.response(c, res)
		return
	}
	data, err := dispatcher.ParseData(c)
	if err != nil {
		res = common.RespMsg{Status: common.ErrorParseBody, Msg: err.Error(), Data: nil}
		g.response(c, res)
		return
	}

	res = g.sendToModule(c.FullPath(), data)
	g.response(c, res)
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
