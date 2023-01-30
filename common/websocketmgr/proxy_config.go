// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"huawei.com/mindxedge/base/common/certutils"
)

// ProxyConfig Websocket proxy config
type ProxyConfig struct {
	name       string
	tlsConfig  *tls.Config
	hosts      string
	headers    http.Header
	handlerMgr WsMsgHandler
	ctx        context.Context
	cancel     context.CancelFunc
}

// RegModInfos registers module info
func (pc *ProxyConfig) RegModInfos(regHandlers []RegisterModuleInfo) {
	for _, reg := range regHandlers {
		pc.handlerMgr.register(reg)
	}
}

// InitProxyConfig init proxy config
func InitProxyConfig(name string, ip string, port int, tlsCertInfo certutils.TlsCertInfo) (*ProxyConfig, error) {
	netConfig := &ProxyConfig{}
	netConfig.name = name
	netConfig.hosts = fmt.Sprintf("%s:%d", ip, port)
	netConfig.handlerMgr = WsMsgHandler{}
	tlsConfig, err := certutils.GetTlsCfgWithPath(tlsCertInfo)
	if err != nil {
		return nil, fmt.Errorf("init proxy config failed: %v", err)
	}
	netConfig.tlsConfig = tlsConfig
	netConfig.headers = http.Header{}
	netConfig.headers.Set(clientNameKey, name)
	netConfig.ctx, netConfig.cancel = context.WithCancel(context.Background())
	return netConfig, nil
}
