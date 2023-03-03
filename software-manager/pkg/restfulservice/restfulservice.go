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
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
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
	// Page is the parameter of paging query in SoftwareInfo struct
	Page = "page"
	// PageSize is the parameter of paging query in SoftwareInfo struct
	PageSize = "pageSize"
	// Description is used to describe the uploaded software
	Description = "description"
)

const (
	serverCertPath = "/home/data/config/mef-certs/software-manager.crt"
	serverKeyPath  = "/home/data/config/mef-certs/software-manager.key"
	rootCaPath     = "/home/data/inner-root-ca/RootCA.crt"
)

const (
	// fileName, fileVersion and serveType are headers of download response
	fileName    = "fileName"
	fileVersion = "fileVersion"
	serveType   = "content-Description"
)

type getURLInfo struct {
	NodeId string `json:"nodeId"`
}

// UserPriInfo is the abbreviation of user private infomation
type UserPriInfo struct {
	Password []byte
	UserName string
}

// NewRestfulService init restful service
func NewRestfulService(enable bool, ip string, port int) model.Module {
	gin.SetMode(gin.ReleaseMode)
	return &restfulService{
		enable: enable,
		ip:     ip,
		httpsSvr: &httpsmgr.HttpsServer{
			Port: port,
			TlsCertPath: certutils.TlsCertInfo{
				RootCaPath: rootCaPath,
				CertPath:   serverCertPath,
				KeyPath:    serverKeyPath,
				SvrFlag:    true,
				KmcCfg:     nil,
			},
		},
	}
}

type restfulService struct {
	httpsSvr *httpsmgr.HttpsServer
	enable   bool
	ip       string
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
	err := r.httpsSvr.Init()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	err = r.httpsSvr.RegisterRoutes(setRouter)
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}
	hwlog.RunLog.Info("start http server now...")
	err = r.httpsSvr.Start()
	if err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, listen failed: %v", r.httpsSvr.Port, err)
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
