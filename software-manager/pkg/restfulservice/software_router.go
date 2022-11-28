// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for restful service router
package restfulservice

import (
	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"mime/multipart"
)

// SoftwareInfo is to warp the request info into a struct object
type SoftwareInfo struct {
	ContentType string
	Version     string
	UserName    []byte
	Password    []byte
	File        *multipart.FileHeader
	Page        string
	PageSize    string
	NodeID      string
}

const (
	// ContentType is the name of software in SoftwareInfo struct
	ContentType = "contentType"
	// Version is the version of software in SoftwareInfo struct
	Version = "version"
	// UserName is used to check the right of downloading in SoftwareInfo struct
	UserName = "user_name"
	// Password is used to check the right of downloading in SoftwareInfo struct
	Password = "password"
	// Page is the parameter of paging query in SoftwareInfo struct
	Page = "page"
	// PageSize is the parameter of paging query in SoftwareInfo struct
	PageSize = "pageSize"
	// nodeID is used to check the right of downloading in SoftwareInfo struct
	nodeID = "node_id"
)

const (
	// fileName, fileVersion and serveType are headers of download response
	fileName    = "FileName"
	fileVersion = "FileVersion"
	serveType   = "Content-Description"
)

var UserInfoMap = make(map[string]map[string][]byte)

func deleteSoftware(c *gin.Context) {
	info := SoftwareInfo{
		ContentType: c.PostForm(ContentType),
		Version:     c.PostForm(Version),
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
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
		UserName:    []byte(c.PostForm(UserName)),
		Password:    []byte(c.PostForm(Password)),
		NodeID:      c.PostForm(nodeID),
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.SoftwareRepositoryName,
		Option:      common.Get,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	if resp.Status == common.Success {
		common.ClearSliceByteMemory(UserInfoMap[info.NodeID][UserName])
		common.ClearSliceByteMemory(UserInfoMap[info.NodeID][Password])
		delete(UserInfoMap, info.NodeID)
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
		Destination: common.SoftwareRepositoryName,
		Option:      common.Update,
		Resource:    common.Software,
	}
	resp := common.SendSyncMessageByRestful(info, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func checkRepository(c *gin.Context) {
	info := SoftwareInfo{
		Page:     c.Query(Page),
		PageSize: c.Query(PageSize),
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
	info := SoftwareInfo{
		ContentType: c.Query(ContentType),
		Version:     c.Query(Version),
		NodeID:      c.PostForm(nodeID),
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
