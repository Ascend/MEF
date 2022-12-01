// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for setup router
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

func setRouter(engine *gin.Engine) {
	engine.Use(gin.Recovery())
	softwareRouter(engine)
}

func softwareRouter(engine *gin.Engine) {
	v1 := engine.Group("/softwaremanager/v1")
	{
		v1.DELETE("/", deleteSoftware)
		v1.POST("/", uploadSoftware)
		v1.GET("/repository", listRepository)
		v1.GET("/", downloadSoftware)
		v1.GET("/url", getURL)
	}
}

func deleteSoftware(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("delete software: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Delete,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func downloadSoftware(c *gin.Context) {
	info, err := downloadInfoMapping(c)
	if err != nil {
		hwlog.OpLog.Error("download software: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Get,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	if resp.Status != common.Success {
		common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
		return
	}
	common.ClearSliceByteMemory(userInfoMap[info.NodeID].Password)
	delete(userInfoMap, info.NodeID)
	c.Header(serveType, "File Transfer")
	c.Header(fileName, info.ContentType)
	c.Header(fileVersion, info.Version)
	c.File(resp.Data.(string))
}

func uploadSoftware(c *gin.Context) {
	info, err := uploadInfoMapping(c)
	if err != nil {
		hwlog.OpLog.Error("upload software: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Update,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listRepository(c *gin.Context) {
	info, err := listRepoMapping(c)
	if err != nil {
		hwlog.OpLog.Error("list repository: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Get,
		Resource:    common.Repository,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getURL(c *gin.Context) {
	info, err := getURLMapping(c)
	if err != nil {
		hwlog.OpLog.Error("get URL: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Get,
		Resource:    common.URL,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
