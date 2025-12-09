// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package httpsmgr for
package httpsmgr

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
)

const (
	defaultReadTimeout       = 60 * time.Second
	defaultReadHeaderTimeout = time.Duration(0)
	defaultWriteTimeout      = 90 * time.Second
	maxHeaderBytes           = 1 * 1024
)

// HttpsServer [struct] for HttpsServer init parameters
type HttpsServer struct {
	ServerParam
	TlsCertPath       certutils.TlsCertInfo
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	server            *http.Server
	engine            *gin.Engine
	SwitchLimit       bool
	IP                string
	Port              int
}

// ServerParam limit server parameter
type ServerParam struct {
	Concurrency    int
	BodySizeLimit  int64
	LimitIPReq     string
	LimitIPConn    int
	LimitTotalConn int
	CacheSize      int
	BurstIPReq     int
}

// Start [method] for start http server
func (ghs *HttpsServer) Start() error {
	if ghs.server == nil {
		return errors.New("gin server is not init, please init first")
	}
	if !ghs.SwitchLimit {
		err := ghs.server.ListenAndServeTLS("", "")
		if err != nil {
			return utils.TrimInfoFromError(err)
		}
		return nil
	}
	handler, err := limiter.NewLimitHandlerV3(ghs.engine, initConfig(ghs.ServerParam))
	if err != nil {
		return err
	}
	ghs.server.Handler = handler
	ln, err := net.Listen("tcp", ghs.server.Addr)
	if err != nil {
		return err
	}
	limitLs, err := limiter.LimitListener(ln, ghs.LimitTotalConn, ghs.LimitIPConn, ghs.CacheSize)
	if err != nil {
		return err
	}
	if err := ghs.server.ServeTLS(limitLs, "", ""); err != nil {
		return utils.TrimInfoFromError(err)
	}
	return nil
}

// RegisterRoutes [method] for Register gin routers
func (ghs *HttpsServer) RegisterRoutes(setRouterFunc func(*gin.Engine)) error {
	if ghs.engine == nil {
		return errors.New("gin engine is not init, please init first")
	}
	setRouterFunc(ghs.engine)
	return nil
}

// Init [method] for Init https sever
func (ghs *HttpsServer) Init() error {
	if err := ghs.checkArgs(); err != nil {
		return fmt.Errorf("invalid server address, %s:%d", ghs.IP, ghs.Port)
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	ghs.setMidUse(engine)
	ghs.TlsCertPath.SvrFlag = true
	tlsCfg, err := certutils.GetTlsCfgWithPath(ghs.TlsCertPath)
	if err != nil {
		return err
	}
	var (
		readTimeout       = defaultReadTimeout
		writeTimeout      = defaultWriteTimeout
		readHeaderTimeout = defaultReadHeaderTimeout
	)
	if ghs.ReadTimeout > 0 {
		readTimeout = ghs.ReadTimeout
	}
	if ghs.WriteTimeout > 0 {
		writeTimeout = ghs.WriteTimeout
	}
	if ghs.ReadHeaderTimeout > 0 {
		readHeaderTimeout = ghs.ReadHeaderTimeout
	}
	server := &http.Server{
		Addr:              net.JoinHostPort(ghs.IP, strconv.Itoa(ghs.Port)),
		Handler:           engine,
		TLSConfig:         tlsCfg,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
	ghs.engine = engine
	ghs.server = server
	return nil
}

func (ghs *HttpsServer) checkArgs() error {
	if checkResult := checker.GetIpV4Checker("", true).Check(ghs.IP); !checkResult.Result {
		return fmt.Errorf("ip [%s] is not supported, %s", ghs.IP, checkResult.Reason)
	}
	if ghs.Port == 0 {
		return errors.New("random port is not supported")
	}
	return nil
}

func (ghs *HttpsServer) setMidUse(engine *gin.Engine) {
	engine.Use(LoggerAdapter())
	engine.Use(gin.Recovery())
	engine.Use(serializable())
}

func initConfig(param ServerParam) *limiter.HandlerConfigV3 {
	conf := limiter.HandlerConfig{
		PrintLog:         false,
		Method:           "",
		LimitBytes:       param.BodySizeLimit,
		TotalConCurrency: param.Concurrency,
		IPConCurrency:    param.LimitIPReq,
		CacheSize:        param.CacheSize,
	}
	configV3 := limiter.HandlerConfigV3{
		HandlerConfig: conf,
		IPBurst:       param.BurstIPReq,
	}
	return &configV3
}
