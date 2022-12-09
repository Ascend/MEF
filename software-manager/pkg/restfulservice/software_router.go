// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for restful service router
package restfulservice

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// SoftwareInfo is to warp the request info into a struct object
type SoftwareInfo struct {
	ContentType string
	Version     string
	File        *multipart.FileHeader
}

// ContentType is the name of software
const ContentType = "contentType"

// Version is the version of software
const Version = "version"
const fileName = "FileName"
const fileVersion = "FileVersion"
const serveType = "Content-Description"

func deleteSoftware(c *gin.Context) {
	info := SoftwareInfo{
		ContentType: c.PostForm(ContentType),
		Version:     c.PostForm(Version),
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareManagerName,
		Option:      common.Delete,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
func downloadSoftware(c *gin.Context) {
	info := SoftwareInfo{
		ContentType: c.Query(ContentType),
		Version:     c.Query(Version),
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareManagerName,
		Option:      common.Get,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	if resp.Status == common.Success {
		c.Header(serveType, "File Transfer")
		c.Header(fileName, info.ContentType)
		c.Header(fileVersion, info.Version)
		c.File(resp.Data.(string))
	} else {
		common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
	}

}
func uploadSoftware(c *gin.Context) {
	software, err := c.FormFile("file")
	if err != nil {
		hwlog.RunLog.Error("Read file error")
		return
	}
	info := SoftwareInfo{
		ContentType: c.PostForm(ContentType),
		Version:     c.PostForm(Version),
		File:        software,
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareManagerName,
		Option:      common.Update,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func checkRepository(c *gin.Context) {
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareManagerName,
		Option:      common.Get,
		Resource:    common.Repository,
	}
	resp := common.SendSyncMessageByRestful(nil, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getURL(c *gin.Context) {
	info := SoftwareInfo{
		ContentType: c.Query(ContentType),
		Version:     c.Query(Version),
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareManagerName,
		Option:      common.Get,
		Resource:    common.URL,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
