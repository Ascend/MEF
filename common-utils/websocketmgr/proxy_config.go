// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package websocketmgr

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"
)

// ProxyConfig Websocket proxy config
type ProxyConfig struct {
	name                string
	tlsConfig           *tls.Config
	isServer            bool
	hosts               string
	headers             http.Header
	handlerMgr          modulemgr.HandleMessageIntf
	ctx                 context.Context
	cancel              context.CancelFunc
	headerSizeLimit     int
	readTimeout         time.Duration
	readHeaderTimeout   time.Duration
	writeTimeout        time.Duration
	rpsLimiterCfg       *limiter.RpsLimiterCfg          // rps limiter for each WebSocket connection
	bandwidthLimiterCfg *limiter.BandwidthLimiterConfig // bandwidth limiter shared by all WebSocket connections
}

// RegModInfos registers module info
func (pc *ProxyConfig) RegModInfos(regHandlers []modulemgr.MessageHandlerIntf) {
	if pc.handlerMgr == nil {
		hwlog.RunLog.Errorf("message handler is not initialized")
		return
	}
	for _, reg := range regHandlers {
		pc.handlerMgr.Register(reg)
	}
}

// UpdateTlsCa UpdateTls to update websocket tls config
func (pc *ProxyConfig) UpdateTlsCa(caCertBytes []byte) error {
	if len(caCertBytes) == 0 {
		return fmt.Errorf("invalid ca cert data")
	}
	if err := x509.CheckPemCertChain(caCertBytes); err != nil {
		return fmt.Errorf("root ca cert check failed: %v", err)
	}
	if pc.isServer {
		if ok := pc.tlsConfig.ClientCAs.AppendCertsFromPEM(caCertBytes); !ok {
			return fmt.Errorf("append ca cert to server ca pool failed")
		}
		return nil
	}
	if ok := pc.tlsConfig.RootCAs.AppendCertsFromPEM(caCertBytes); !ok {
		return fmt.Errorf("append ca cert to client ca pool failed")
	}
	return nil
}

// SetRpsLimiterCfg - conn-level rps limiter configuration
// rps represents how many requests are allowed in ONE SECOND
// burst represents maximum requests are allowed at the same time
func (pc *ProxyConfig) SetRpsLimiterCfg(rps float64, burst int) error {
	if rps <= 0 || burst <= 0 {
		return fmt.Errorf("invalid message limiter config. rps: %v, burst: %v", rps, burst)
	}
	pc.rpsLimiterCfg = &limiter.RpsLimiterCfg{
		Rps:   rps,
		Burst: burst,
	}
	return nil
}

// SetBandwidthLimiterCfg set bandwidth limiter configuration
func (pc *ProxyConfig) SetBandwidthLimiterCfg(maxThroughput int, period time.Duration, reserveRate ...float64) error {
	if maxThroughput <= 0 || period <= 0 {
		return fmt.Errorf("invalid bandwidth limiter config. maxThroughput: %v, period: %v", maxThroughput, period)
	}

	resRate := defaultReserveRate
	if len(reserveRate) > 0 && reserveRate[0] >= 0 {
		resRate = reserveRate[0]
	}
	pc.bandwidthLimiterCfg = &limiter.BandwidthLimiterConfig{
		MaxThroughput: maxThroughput,
		Period:        period,
		ReserveRate:   resRate,
	}
	return nil
}

// SetTimeout set 0 to use default timeout
func (pc *ProxyConfig) SetTimeout(rTimeout, wTimeout, rHeaderTime time.Duration) {
	pc.readTimeout = defaultReadTimeout
	if rTimeout > 0 {
		pc.readTimeout = rTimeout
	}
	pc.writeTimeout = defaultWriteTimeout
	if wTimeout > 0 {
		pc.writeTimeout = wTimeout
	}
	pc.readHeaderTimeout = defaultReadTimeout
	if rHeaderTime > 0 {
		pc.readHeaderTimeout = rHeaderTime
	}
}

// SetSizeLimit set 0 to use default timeout
func (pc *ProxyConfig) SetSizeLimit(headerLimit int) {
	pc.headerSizeLimit = defaultHeaderSizeLimit
	if headerLimit > 0 {
		pc.headerSizeLimit = headerLimit
	}
}

// InitProxyConfig init proxy config
func InitProxyConfig(name string, host string, port int, tlsInfo certutils.TlsCertInfo, urlPath ...string,
) (*ProxyConfig, error) {
	netConfig := &ProxyConfig{}
	netConfig.name = name
	netConfig.hosts = net.JoinHostPort(host, strconv.Itoa(port))
	// use in client side
	if len(urlPath) > 0 {
		netConfig.hosts = fmt.Sprintf("%s:%d%s", host, port, urlPath[0])
	}
	netConfig.handlerMgr = &modulemgr.MsgHandler{}
	tlsConfig, err := certutils.GetTlsCfgWithPath(tlsInfo)
	if err != nil {
		return nil, fmt.Errorf("init proxy config failed, error: %v", err)
	}
	netConfig.tlsConfig = tlsConfig
	netConfig.headers = http.Header{}
	netConfig.headers.Set(clientNameKey, name)
	return netConfig, nil
}
