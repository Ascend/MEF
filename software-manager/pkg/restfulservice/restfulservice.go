// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice this file is for restful service router
package restfulservice

import (
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// SoftwareInfo is to warp the request info into a struct object
type SoftwareInfo struct {
	File        *multipart.FileHeader
	ContentType string
	Version     string
	UserName    string
	Password    []byte
	Page        int
	PageSize    int
	NodeID      string
	Description string
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
	// Description is used to describe the uploaded software
	Description = "description"
)

const (
	// fileName, fileVersion and serveType are headers of download response
	fileName    = "fileName"
	fileVersion = "fileVersion"
	serveType   = "content-Description"
)

type downloadInfo struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
	NodeID   string `json:"nodeId"`
}

type getURLInfo struct {
	NodeId string `json:"nodeId"`
}

type UserPriInfo struct {
	Password []byte
	UserName string
}

type restfulService struct {
	engine *gin.Engine
	enable bool
	ip     string
	port   int
}

// Name module name
func (r *restfulService) Name() string {
	return common.RestfulServiceName
}

// Enable module enable
func (r *restfulService) Enable() bool {
	return r.enable
}

// Start module start
func (r *restfulService) Start() {
	r.engine.Use(common.LoggerAdapter())
	setRouter(r.engine)
	hwlog.RunLog.Info("start http server now...")
	if err := r.engine.Run(fmt.Sprintf(":%d", r.port)); err != nil {
		hwlog.RunLog.Errorf("start restful at %d fail", r.port)
	}
}

// UserInfoMap contains current generated username and password
var userInfoMap = make(map[string]UserPriInfo)

// QueryUserInfo is to get the user information of specific nodeID
func QueryUserInfo(nodeID string) (*UserPriInfo, bool) {
	userInfo, ok := userInfoMap[nodeID]
	return &userInfo, ok

}

// AddUserInfo is to add the user information of specific nodeID
func AddUserInfo(nodeID string, info *UserPriInfo) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID is error")
	}
	userInfoMap[nodeID] = *info
	return nil
}

// NewRestfulService init restful service
func NewRestfulService(enable bool, ip string, port int) model.Module {
	gin.SetMode(gin.ReleaseMode)
	return &restfulService{
		enable: enable,
		engine: gin.New(),
		ip:     ip,
		port:   port,
	}
}

func downloadInfoMapping(c *gin.Context) (SoftwareInfo, error) {
	info := SoftwareInfo{
		ContentType: c.Query(ContentType),
		Version:     c.Query(Version),
		UserName:    c.Request.Header.Get("userName"),
		Password:    []byte(c.Request.Header.Get("password")),
		NodeID:      c.Request.Header.Get("nodeId"),
	}
	return info, nil
}

func uploadInfoMapping(c *gin.Context) (SoftwareInfo, error) {
	software, err := c.FormFile("file")
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return SoftwareInfo{}, err
	}
	info := SoftwareInfo{
		ContentType: c.PostForm(ContentType),
		Version:     c.PostForm(Version),
		File:        software,
		Description: c.PostForm(Description),
	}
	return info, nil
}

func listRepoMapping(c *gin.Context) (SoftwareInfo, error) {
	page, err := strconv.Atoi(c.Query(Page))
	if err != nil {
		hwlog.RunLog.Error("class convert error")
		return SoftwareInfo{}, err
	}
	pageSize, err := strconv.Atoi(c.Query(PageSize))
	if err != nil {
		hwlog.RunLog.Error("class convert error")
		return SoftwareInfo{}, err
	}
	info := SoftwareInfo{
		Page:     page,
		PageSize: pageSize,
	}
	return info, nil
}

func getURLMapping(c *gin.Context) (SoftwareInfo, error) {
	var subInfo getURLInfo
	err := c.ShouldBindJSON(&subInfo)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return SoftwareInfo{}, err
	}
	info := SoftwareInfo{
		ContentType: c.Query(ContentType),
		Version:     c.Query(Version),
		NodeID:      subInfo.NodeId,
	}
	return info, nil
}
