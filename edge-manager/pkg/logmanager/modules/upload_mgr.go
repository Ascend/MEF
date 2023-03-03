// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package modules enables collecting logs
package modules

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-manager/pkg/logmanager/constants"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
)

const (
	paramNameFileName = "filename"
	servicePort       = 10002

	certPathDir = "/home/data/config/websocket-certs"
	serviceName = "server.crt"
	keyFileName = "server.key"

	maxRetry = math.MaxInt
	waitTime = 5 * time.Second
)

// UploadMgr manages uploading
type UploadMgr interface {
	// Start starts upload manager
	Start()
}

type uploadMgr struct {
	httpsSvr *httpsmgr.HttpsServer
}

// NewUploadMgr creates new upload manager
func NewUploadMgr(context.Context) UploadMgr {
	return &uploadMgr{}
}

func (u *uploadMgr) Start() {
	go u.start()
}

func (u *uploadMgr) start() {
	if !checkWsSvcCert() {
		hwlog.RunLog.Error("check websocket service cert failed")
		return
	}
	rootCaBytes, err := getWsRootCert()
	if err != nil {
		hwlog.RunLog.Errorf("get root ca failed, %v", err)
		return
	}
	u.httpsSvr = &httpsmgr.HttpsServer{
		IP:   os.Getenv("POD_IP"),
		Port: servicePort,
		TlsCertPath: certutils.TlsCertInfo{
			KmcCfg:        common.GetDefKmcCfg(),
			RootCaContent: rootCaBytes,
			CertPath:      path.Join(certPathDir, serviceName),
			KeyPath:       path.Join(certPathDir, keyFileName),
			SvrFlag:       true,
		},
	}

	if err = u.httpsSvr.Init(); err != nil {
		hwlog.RunLog.Errorf("init http server failed, %v", err)
		return
	}

	if err = u.httpsSvr.RegisterRoutes(u.setRouter); err != nil {
		hwlog.RunLog.Errorf("register routes failed, %v", err)
		return
	}

	hwlog.RunLog.Info("start http server now...")
	if err := u.httpsSvr.Start(); err != nil {
		hwlog.RunLog.Errorf("start http server failed, %v", err)
	}
	hwlog.RunLog.Info("start http server success")
}

func (u *uploadMgr) setRouter(engine *gin.Engine) {
	engine.POST(fmt.Sprintf("%s/:%s", constants.UploadUrlPathPrefix, paramNameFileName), u.handleUpload)
}

func (u *uploadMgr) handleUpload(c *gin.Context) {
	hwlog.RunLog.Info("start to handle file uploading")
	filename := c.Param(paramNameFileName)
	if filename != filepath.Base(filename) {
		u.abort(c, http.StatusBadRequest, "illegal filename")
		return
	}

	exportPath := filepath.Join("/home/MEFCenter/log_exports", filename)
	file, err := os.OpenFile(exportPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, common.Mode600)
	if err != nil {
		u.abort(c, http.StatusInternalServerError, "can't open upload file")
		return
	}
	var success bool
	defer func() {
		if err := file.Close(); err != nil {
			hwlog.RunLog.Error("failed to close upload file")
		}
		if !success {
			if err := common.DeleteFile(exportPath); err != nil {
				hwlog.RunLog.Error("failed to delete upload file")
			}
		}
	}()
	limitedReader := io.LimitReader(c.Request.Body, logcollect.EdgeMaxPackSize)
	if _, err := io.Copy(file, limitedReader); err != nil {
		u.abort(c, http.StatusInternalServerError, "can't store upload file")
		return
	}

	success = true
	hwlog.RunLog.Info("handle file uploading successful")
	c.String(http.StatusOK, "success")
}

func (u *uploadMgr) abort(c *gin.Context, status int, reason string) {
	hwlog.RunLog.Errorf("failed to handle uploading: %s", reason)
	c.AbortWithStatus(status)
}

func getWsRootCert() ([]byte, error) {
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	var rootCaStr string
	var err error
	for i := 0; i < maxRetry; i++ {
		rootCaStr, err = reqCertParams.GetRootCa(common.WsCltName)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	if rootCaStr == "" {
		hwlog.RunLog.Errorf("get valid root ca for websocket service failed: %v", err)
		return nil, err
	}

	return []byte(rootCaStr), nil
}

func checkWsSvcCert() bool {
	keyPath := path.Join(certPathDir, keyFileName)
	certPath := path.Join(certPathDir, serviceName)
	var retry int
	for retry < maxRetry {
		if utils.IsExist(keyPath) && utils.IsExist(certPath) {
			hwlog.RunLog.Info("check websocket server certs success")
			return true
		}
		time.Sleep(waitTime)
		hwlog.RunLog.Warnf("check websocket server certs failed, retry in 15 seconds")
		retry++
	}
	return false
}
