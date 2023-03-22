// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package httpsmgr for https manager
package httpsmgr

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
)

const (
	readTimeout    = 60 * time.Second
	writeTimeout   = 90 * time.Second
	maxHeaderBytes = 1 * 1024 * 1024
)

// HttpsServer [struct] for HttpsServer init parameters
type HttpsServer struct {
	IP          string
	Port        int
	TlsCertPath certutils.TlsCertInfo
	server      *http.Server
	engine      *gin.Engine
}

// Start [method] for start http server
func (ghs *HttpsServer) Start() error {
	if ghs.server == nil {
		return errors.New("gin server is not init, please init first")
	}
	err := ghs.server.ListenAndServeTLS("", "")
	if err != nil {
		return common.TrimInfoFromError(err, ghs.IP)
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
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	ghs.setMidUse(engine)
	ghs.TlsCertPath.SvrFlag = true
	tlsCfg, err := certutils.GetTlsCfgWithPath(ghs.TlsCertPath)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", ghs.IP, ghs.Port),
		Handler:        engine,
		TLSConfig:      tlsCfg,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	ghs.engine = engine
	ghs.server = server
	return nil
}

func (ghs *HttpsServer) setMidUse(engine *gin.Engine) {
	engine.Use(common.LoggerAdapter())
	engine.Use(gin.Recovery())
}
