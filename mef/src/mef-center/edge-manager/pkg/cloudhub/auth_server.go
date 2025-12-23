// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cloudhub for
package cloudhub

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindx/common/xcrypto"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"

	"edge-manager/pkg/configmanager"
	"edge-manager/pkg/constants"
)

const (
	edgeAuthUrl     = "/token"
	edgeConnTestUrl = "/token-check"
	headerToken     = "token"
	dataSize        = 1024 * 1024
	maxCsrBytes     = 4096
	connSize        = 100
	limitIP         = "2/1"
)

// ClientAuthService [struct] for mef edge client auth
type ClientAuthService struct {
	httpsSvr *httpsmgr.HttpsServer
}

// NewClientAuthService new mef edge client auth service
func NewClientAuthService(serverIp string, port int, tlsCfg certutils.TlsCertInfo) *ClientAuthService {
	return &ClientAuthService{
		httpsSvr: &httpsmgr.HttpsServer{
			IP:          serverIp,
			Port:        port,
			SwitchLimit: true,
			ServerParam: httpsmgr.ServerParam{
				BurstIPReq:     connSize,
				Concurrency:    connSize,
				LimitIPReq:     limitIP,
				BodySizeLimit:  dataSize,
				LimitIPConn:    1,
				LimitTotalConn: connSize,
				CacheSize:      dataSize,
			},

			TlsCertPath: tlsCfg,
		},
	}
}

// Start ClientAuthService for mef edge auth
func (r *ClientAuthService) Start() {
	if err := r.httpsSvr.Init(); err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, init https server failed: %v", r.httpsSvr.Port, err)
		return
	}
	if err := r.httpsSvr.RegisterRoutes(setRouter); err != nil {
		hwlog.RunLog.Errorf("start restful at %d failed, set routers failed: %v", r.httpsSvr.Port, err)
		return
	}

	hwlog.RunLog.Info("start http server for edge auth now...")
	if err := r.httpsSvr.Start(); err != nil {
		hwlog.RunLog.Errorf("start restful at %d fail", r.httpsSvr.Port)
	}
}

func setRouter(engine *gin.Engine) {
	engine.POST(edgeAuthUrl, ClientAuth)
	engine.GET(edgeConnTestUrl, EdgeConnTest)
}

// ClientAuth check mef edge account then issue a cert if check passed
func ClientAuth(c *gin.Context) {
	status, err := checkEdgeToken(c)
	if err != nil {
		hwlog.RunLog.Errorf("check token request error: %v", err)
		c.String(status, "auth failed")
		return
	}

	csrData, err := io.ReadAll(io.LimitReader(c.Request.Body, maxCsrBytes))
	if err != nil {
		hwlog.RunLog.Errorf("read crs data from edge error: %v", err)
		c.String(http.StatusBadRequest, "csr data is required")
		return
	}
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
	certStr, err := reqCertParams.ReqIssueSvrCert(common.WsCltName, csrData)
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
	status, err := checkEdgeToken(c)
	if err != nil {
		hwlog.RunLog.Errorf("check token request error: %v", err)
		c.String(status, "test connect failed")
		return
	}

	c.String(http.StatusOK, "")
}

func checkEdgeToken(c *gin.Context) (int, error) {
	ip := c.ClientIP()
	lock, err := LockRepositoryInstance().isLock()
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("%s check lock status failed: %v", ip, err)
	}
	if lock {
		return http.StatusLocked, errors.New("token is locked")
	}

	token := c.GetHeader(headerToken)
	defer utils.ClearStringMemory(token)

	if match := regexp.MustCompile(common.PassWordRegex).MatchString(token); !match {
		if err := LockRepositoryInstance().recordFailed(ip); err != nil {
			return http.StatusUnauthorized, err
		}
		return http.StatusUnauthorized, errors.New("token check failed")
	}
	dbToken, salt, err := configmanager.ConfigRepositoryInstance().GetToken()
	if err != nil {
		return http.StatusBadRequest, err
	}

	encryptRawToken, err := xcrypto.Pbkdf2WithSha256([]byte(token), salt,
		common.Pbkdf2IterationCount, common.BytesOfEncryptedString)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("encrypt token from request failed: %v", err)
	}
	if !bytes.Equal(encryptRawToken, dbToken) {
		if err := LockRepositoryInstance().recordFailed(ip); err != nil {
			return http.StatusUnauthorized, err
		}
		return http.StatusUnauthorized, fmt.Errorf("edge ip %v send an incorrect token", ip)
	}
	if err := LockRepositoryInstance().deleteOneFailedRecord(ip); err != nil {
		return http.StatusBadRequest, err
	}
	return 0, nil
}
