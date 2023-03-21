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
func InitProxyConfig(nm string, ip string, port int, tls certutils.TlsCertInfo, url ...string) (*ProxyConfig, error) {
	netConfig := &ProxyConfig{}
	netConfig.name = nm
	netConfig.hosts = fmt.Sprintf("%s:%d", ip, port)
	// use in client side
	if len(url) > 0 {
		netConfig.hosts = fmt.Sprintf("%s:%d%s", ip, port, url[0])
	}
	netConfig.handlerMgr = WsMsgHandler{}
	tlsConfig, err := certutils.GetTlsCfgWithPath(tls)
	if err != nil {
		return nil, fmt.Errorf("init proxy config failed, error: %v", err)
	}
	netConfig.tlsConfig = tlsConfig
	netConfig.headers = http.Header{}
	netConfig.headers.Set(clientNameKey, nm)
	netConfig.ctx, netConfig.cancel = context.WithCancel(context.Background())
	return netConfig, nil
}
