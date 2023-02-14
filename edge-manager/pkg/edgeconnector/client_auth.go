// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector client_auth used to verify edge side account and issue client cert
package edgeconnector

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

const (
	edgeAuthUrl        = "/auth"
	edgeConnTestUrl    = "/account-check"
	edgeUserAccountKey = "MEFUserAccount"
	edgeUserPwdKey     = "MEFUserPwd"
	maxBodyBytes       = 2 * 1024 * 1024
)

// ClientAuthService [struct] for mef edge client auth
type ClientAuthService struct {
	httpsSvr *httpsmgr.HttpsServer
}

// NewClientAuthService new mef edge client auth service
func NewClientAuthService(port int, tlsCfg certutils.TlsCertInfo) *ClientAuthService {
	return &ClientAuthService{
		httpsSvr: &httpsmgr.HttpsServer{
			Port:        port,
			TlsCertPath: tlsCfg,
		},
	}
}

// Start ClientAuthService for mef edge auth
func (r *ClientAuthService) Start() {
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
		hwlog.RunLog.Errorf("start restful at %d fail", r.httpsSvr.Port)
	}
}

func setRouter(engine *gin.Engine) {
	engine.GET(edgeConnTestUrl, EdgeConnTest)
	engine.POST(edgeAuthUrl, ClientAuth)
}

// ClientAuth check mef edge account then issue a cert if check passed
func ClientAuth(c *gin.Context) {
	passed, err := checkEdgeAuth(c)
	if err != nil {
		hwlog.RunLog.Errorf("checkEdgeAuth error: %v", err)
		c.String(http.StatusBadRequest, "check account failed")
		return
	}
	if !passed {
		c.String(http.StatusBadRequest, "user account or password is invalid")
		return
	}
	csrReader := io.LimitReader(c.Request.Body, maxBodyBytes)
	csrData := make([]byte, maxBodyBytes)
	readBytes, err := csrReader.Read(csrData)
	if err != nil {
		hwlog.RunLog.Errorf("read crs data from edge error: %v", err)
		c.String(http.StatusBadRequest, "csr data is required")
		return
	}
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    util.RootCaPath,
			CertPath:      util.ServerCertPath,
			KeyPath:       util.ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	certStr, err := reqCertParams.ReqIssueSvrCert(common.WsCltName, csrData[:readBytes])
	if err != nil {
		hwlog.RunLog.Errorf("issue cert for edge error: %v", err)
		c.String(http.StatusBadRequest, "generate edge cert failed")
		return
	}
	c.String(http.StatusOK, certStr)
	return
}

// EdgeConnTest check mef edge account
func EdgeConnTest(c *gin.Context) {
	passed, err := checkEdgeAuth(c)
	if err != nil {
		hwlog.RunLog.Errorf("checkEdgeAuth error: %v", err)
		c.String(http.StatusBadRequest, "check account failed")
		return
	}
	if !passed {
		c.String(http.StatusBadRequest, "user account or password is invalid")
		return
	}
	c.String(http.StatusOK, "check account passed!")
}

func checkEdgeAuth(c *gin.Context) (bool, error) {
	account := c.GetHeader(edgeUserAccountKey)
	pwd := c.GetHeader(edgeUserPwdKey)
	if account == "" || pwd == "" {
		return false, fmt.Errorf("account and password are required")
	}
	defer common.ClearStringMemory(pwd)
	return checkDBAccount(account, []byte(pwd))
}

// todo
func checkDBAccount(userName string, pwd []byte) (bool, error) {
	defer common.ClearSliceByteMemory(pwd)
	testUserName := "EdgeAccount"
	testPwd := "Atlas12#$" // todo, will be deleted in future
	return userName == testUserName && string(pwd) == testPwd, nil
}
