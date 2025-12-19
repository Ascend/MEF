// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgeproxy forward msg from device-om, edgecore or edge-om
package edgeproxy

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/msgconv"
)

// StartEdgeProxy start new edge proxy(websocket server)
func StartEdgeProxy() error {
	if err := initEdgeProxyServer(); err != nil {
		hwlog.RunLog.Errorf("edge-proxy server init error: %v", err)
		return err
	}
	hwlog.RunLog.Info("edge-proxy server success")

	RegistryMsgRouters()
	return msgconv.Init(forwardingRegisterInfoList...)
}

func getCertInfo() (*certutils.TlsCertInfo, error) {
	certContents, err := common.GetWsCertContent()
	if err != nil {
		return nil, err
	}
	innerCertDir, err := path.GetCompSpecificDir(constants.InnerCertPathName)
	if err != nil {
		return nil, err
	}
	certPath := filepath.Join(innerCertDir, constants.ServerCertName)
	keyPath := filepath.Join(innerCertDir, constants.ServerKeyName)
	kmcCfg, err := util.GetKmcConfig("")
	if err != nil {
		return nil, err
	}
	return &certutils.TlsCertInfo{
		RootCaContent: certContents.RootCaContent,
		KeyPath:       keyPath,
		CertPath:      certPath,
		KmcCfg:        kmcCfg,
		SvrFlag:       true,
		WithBackup:    true,
	}, nil
}

func getProxyConfig() (*websocketmgr.ProxyConfig, error) {
	certInfo, err := getCertInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get server cert content error: %v", err)
		return nil, err
	}

	// if modify this param [LocalIp], operation log ip for edge-om proc need to modify together
	return websocketmgr.InitProxyConfig(constants.ServerIdName, constants.LocalIp,
		constants.InnerServerPort, *certInfo)
}

func startWebsocketServer(proxyConfig *websocketmgr.ProxyConfig) error {
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}

	if err := setHandlers(proxy); err != nil {
		hwlog.RunLog.Errorf("set http handler error: %v", err)
		return err
	}

	if err := proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("server proxy start error: %v", err)
		return err
	}

	return nil
}

// initEdgeProxyServer init mef edge proxy websocket server
func initEdgeProxyServer() error {
	proxyConfig, err := getProxyConfig()
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config error: %v", err)
		return err
	}

	if err := startWebsocketServer(proxyConfig); err != nil {
		hwlog.RunLog.Errorf("server proxy start error: %v", err)
		return err
	}

	return nil
}

func getConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	if !websocket.IsWebSocketUpgrade(r) {
		w.WriteHeader(http.StatusBadRequest)
		hwlog.RunLog.Errorf("request from %v is not a websocket request", r.URL.String())
		return nil, fmt.Errorf("request from %v is not a websocket request", r.URL.String())
	}

	upgrade := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			remoteAddr := strings.Split(r.RemoteAddr, constants.IpPortSeparator)
			if len(remoteAddr) != constants.IpPortSliceLen {
				hwlog.RunLog.Errorf("remote address is invalid %s", r.RemoteAddr)
				return false
			}
			if remoteAddr[0] != constants.LocalIp {
				hwlog.RunLog.Errorf("remote address is not local ip %s", r.RemoteAddr)
				return false
			}
			hwlog.RunLog.Info("remote address is valid")
			return true
		},
	}

	conn, err := upgrade.Upgrade(w, r, http.Header{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		hwlog.RunLog.Errorf("websocket upgrade error: %v", err)
		return nil, fmt.Errorf("websocket upgrade error: %v", err)
	}
	return conn, nil
}

func newProxy(url string) proxyInterface {
	if strings.HasPrefix(url, constants.DeviceOmSvcUrl) {
		return &DeviceOmProxy{}
	}
	if strings.HasPrefix(url, constants.EdgeCoreSvcUrl) {
		return &EdgecoreProxy{}
	}
	if strings.HasPrefix(url, constants.EdgeOmSvcUrl) {
		return &EdgeOmProxy{}
	}
	return nil
}

func handleConnectReq(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		hwlog.RunLog.Error("handle connect request input is invalid")
		return
	}

	conn, err := getConnection(w, r)
	if err != nil {
		return
	}
	proxy := newProxy(r.URL.String())
	if proxy == nil {
		hwlog.RunLog.Errorf("can not get proxy by url %s", r.URL.String())
		return
	}
	if err := proxy.Start(conn); err != nil {
		hwlog.RunLog.Errorf("start proxy fail by url %s", r.URL.String())
	}
}

func setHandlers(proxy *websocketmgr.WsServerProxy) error {
	if proxy == nil {
		return fmt.Errorf("invalid server proxy")
	}
	netType, err := configpara.GetNetType()
	if err != nil {
		return err
	}

	if netType == constants.FDWithOM {
		if err := proxy.AddHandler(constants.DeviceOmSvcUrl, handleConnectReq); err != nil {
			return err
		}
	}
	if err := proxy.AddHandler(constants.EdgeCoreSvcUrl, handleConnectReq); err != nil {
		return err
	}
	if err := proxy.AddHandler(constants.EdgeOmSvcUrl, handleConnectReq); err != nil {
		return err
	}
	return nil
}
