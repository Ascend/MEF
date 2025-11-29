// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

const (
	defaultReserveRate = 0.5
	regexpSerialNumber = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
)

// WsServerProxy websocket server proxy
type WsServerProxy struct {
	ProxyCfg         *ProxyConfig
	httpServer       *http.Server
	upgrade          *websocket.Upgrader
	handlerMap       sync.Map
	clientMap        sync.Map
	clientNum        int
	clientLimitNum   int
	counterLock      sync.Mutex
	disConnCallbacks []func(WebsocketPeerInfo)
	onConnCallbacks  []func(WebsocketPeerInfo)
	bandwidthLimiter limiter.ServerBandwidthLimiterIntf
	connLimiter      limiter.ConnLimiterIntf
}

// GetName get websocket server name
func (wsp *WsServerProxy) GetName() string {
	return wsp.ProxyCfg.name
}

// Start websocket server start
func (wsp *WsServerProxy) Start() error {
	if err := wsp.checkArgs(); err != nil {
		return fmt.Errorf("invalid server address, %s, %v", wsp.ProxyCfg.hosts, err)
	}
	wsp.ProxyCfg.ctx, wsp.ProxyCfg.cancel = context.WithCancel(context.Background())
	httpServer := &http.Server{
		Addr:              wsp.ProxyCfg.hosts,
		TLSConfig:         wsp.ProxyCfg.tlsConfig,
		MaxHeaderBytes:    wsp.ProxyCfg.headerSizeLimit,
		ReadTimeout:       wsp.ProxyCfg.readTimeout,
		ReadHeaderTimeout: wsp.ProxyCfg.readHeaderTimeout,
		WriteTimeout:      wsp.ProxyCfg.writeTimeout,
	}
	httpServer.SetKeepAlivesEnabled(false)
	wsp.httpServer = httpServer
	wsp.initHandlers()
	wsp.upgrade = &websocket.Upgrader{
		ReadBufferSize:   wsReadBufferSize,
		WriteBufferSize:  wsWriteBufferSize,
		HandshakeTimeout: defaultHandshakeTimeout,
	}

	// initialize then start bandwidth limiter immediately when client proxy starts
	bandwidthLimiter := limiter.NewServerBandwidthLimiter(wsp.ProxyCfg.bandwidthLimiterCfg)
	if bandwidthLimiter != nil {
		if bwl := wsp.bandwidthLimiter; bwl != nil {
			bwl.Stop()
		}
		wsp.bandwidthLimiter = bandwidthLimiter
	}
	go func() {
		if err := wsp.listen(); err != nil {
			return
		}
	}()
	return nil
}

// Stop websocket server Stop
func (wsp *WsServerProxy) Stop() error {
	wsp.ProxyCfg.cancel()

	wsp.clientMap.Range(wsp.closeOneClient)
	if wsp.bandwidthLimiter != nil {
		wsp.bandwidthLimiter.Stop()
	}
	if err := wsp.httpServer.Close(); err != nil {
		return fmt.Errorf("stop websocket server failed, error: %v",
			utils.TrimInfoFromError(err))
	}
	return nil
}

// Send websocket server send message
func (wsp *WsServerProxy) Send(clientId string, msg *model.Message, msgType ...int) error {
	wsMsgType := websocket.TextMessage
	if len(msgType) > 0 {
		wsMsgType = msgType[0]
	}
	if !isValidWsMsgType(wsMsgType) {
		return fmt.Errorf("websocket message type [%v] is not supported", wsMsgType)
	}
	cltConnMgr, ok := wsp.clientMap.Load(clientId)
	if !ok {
		return fmt.Errorf("websocket sever send failed: the client [%v] not connect", clientId)
	}
	connMgr, ok := cltConnMgr.(*wsConnectMgr)
	if !ok {
		return fmt.Errorf("websocket sever send failed: the connect manager type [%T] unsupported", cltConnMgr)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json marshal message error: %v", err)
	}
	return connMgr.send(wsMsgType, data)
}

func (wsp *WsServerProxy) closeOneClient(name, conn interface{}) bool {
	wsConn, ok := conn.(*wsConnectMgr)
	if !ok {
		hwlog.RunLog.Errorf("close client [%v] failed: conn[%T] not a valid conn", name, conn)
		return true
	}
	if err := wsConn.stop(); err != nil {
		hwlog.RunLog.Errorf("close client [%v] failed, error: %v", name, err)
		return true
	}
	wsp.clientMap.Delete(name)
	hwlog.RunLog.Infof("client [name=%v] disconnect", name)

	return true
}

func (wsp *WsServerProxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		hwlog.RunLog.Errorf("request is not a websocket request")
		return
	}
	if wsp.connLimiter != nil && !wsp.connLimiter.ConnAdd() {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("max websocket client connection reached, please try again later")); err != nil {
			hwlog.RunLog.Errorf("write response error: %v", utils.TrimInfoFromError(err))
		}
		return
	}
	if wsp.connLimiter != nil {
		defer wsp.connLimiter.ConnDone()
	}
	clientName := r.Header.Get(clientNameKey)
	if matchFlag := regexp.MustCompile(regexpSerialNumber).MatchString(clientName); !matchFlag {
		hwlog.RunLog.Error("client name is invalid")
		return
	}
	ip := r.Header.Get(realIpKey)
	if ip != "" && net.ParseIP(ip) == nil {
		hwlog.RunLog.Error("ip is invalid")
		return
	}
	conn, err := wsp.upgrade.Upgrade(w, r, nil)
	if err != nil {
		hwlog.RunLog.Errorf("websocket start server http failed, error: %v", utils.TrimInfoFromError(err))
		return
	}
	connMgr := &wsConnectMgr{
		conn:         conn,
		peerInfo:     WebsocketPeerInfo{Ip: ip, Sn: clientName},
		currentProxy: wsp,
	}
	_, loaded := wsp.clientMap.LoadOrStore(clientName, connMgr)
	if loaded {
		hwlog.RunLog.Errorf("client %s has already connected", clientName)
		if err := conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close client %s failed, error: %v", clientName, err)
		}
		return
	}

	hwlog.RunLog.Infof("client [name=%v] [ip=%v] is connected", clientName, ip)
	hwlog.OpLog.Infof("build websocket connection with [name=%v] [ip=%v] success", clientName, ip)
	connMgr.start()
	<-connMgr.ctx.Done()
	wsp.clientMap.Delete(clientName)
}

func (wsp *WsServerProxy) listen() error {
	if wsp.httpServer == nil {
		return fmt.Errorf("https server not init, can not listen")
	}
	for {
		select {
		case <-wsp.ProxyCfg.ctx.Done():
			return nil
		default:
		}
		if err := wsp.httpServer.ListenAndServeTLS("", ""); err != nil {
			hwlog.RunLog.Errorf("websocket listen and serve with tls failed, error: %v",
				utils.TrimInfoFromError(err))
		}
		time.Sleep(retryInterval)
	}
}

func (wsp *WsServerProxy) checkArgs() error {
	host, portStr, err := net.SplitHostPort(wsp.ProxyCfg.hosts)
	if err != nil {
		return err
	}
	if checkResult := checker.GetIpV4Checker("", true).Check(host); !checkResult.Result {
		return fmt.Errorf("ip [%s] is not supported, %v", host, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("unable to parse port %s, %v", portStr, err)
	}
	if port == 0 {
		return errors.New("random port is not supported")
	}
	return nil
}

// AddHandler set customized url and handler
func (wsp *WsServerProxy) AddHandler(url string, handler func(http.ResponseWriter, *http.Request)) error {
	if handler == nil {
		return fmt.Errorf("invalid handler")
	}
	if _, existed := wsp.handlerMap.LoadOrStore(url, handler); existed {
		return fmt.Errorf("url is already registered")
	}
	return nil
}

// AddDefaultHandler set default / handler
func (wsp *WsServerProxy) AddDefaultHandler() {
	wsp.handlerMap.Store(svcUrl, wsp.serveHTTP)
}

// initHandlers must be called after http.Server initialized
func (wsp *WsServerProxy) initHandlers() {
	handlerMux := http.NewServeMux()
	wsp.handlerMap.Range(func(key, value interface{}) bool {
		urlPath, ok := key.(string)
		if !ok {
			hwlog.RunLog.Error("initHandlers url path type error")
			return false
		}
		handlerFunc, ok := value.(func(http.ResponseWriter, *http.Request))
		if !ok {
			hwlog.RunLog.Error("initHandlers handler function type error")
			return false
		}
		handlerMux.HandleFunc(urlPath, handlerFunc)
		return true
	})
	wsp.httpServer.Handler = handlerMux
}

// SetOnConnCallback is the func to set the callback func when client disconnect
// set nil means clear the callback
func (wsp *WsServerProxy) SetOnConnCallback(cf ...func(WebsocketPeerInfo)) {
	if len(cf) == 0 {
		wsp.onConnCallbacks = []func(WebsocketPeerInfo){}
		return
	}
	wsp.onConnCallbacks = cf
}

// SetDisconnCallback is the func to set the callback func when client disconnect
// set nil means clear the callback
func (wsp *WsServerProxy) SetDisconnCallback(cf ...func(WebsocketPeerInfo)) {
	if len(cf) == 0 {
		wsp.disConnCallbacks = []func(WebsocketPeerInfo){}
		return
	}
	wsp.disConnCallbacks = cf
}

// GetAllPeers gets all WebsocketPeerInfo
func (wsp *WsServerProxy) GetAllPeers() ([]WebsocketPeerInfo, error) {
	var (
		peerInfos []WebsocketPeerInfo
		err       error
	)
	wsp.clientMap.Range(func(key, value interface{}) bool {
		connMgr, ok := value.(*wsConnectMgr)
		if !ok {
			err = fmt.Errorf("unexpected type of client conn for %v", key)
			return false
		}
		peerInfos = append(peerInfos, connMgr.peerInfo)
		return true
	})
	return peerInfos, err
}

// GetPeer gets a single peer
func (wsp *WsServerProxy) GetPeer(clientName string) (*WebsocketPeerInfo, error) {
	value, ok := wsp.clientMap.Load(clientName)
	if !ok {
		return nil, fmt.Errorf("client %s not found", clientName)
	}
	connMgr, ok := value.(*wsConnectMgr)
	if !ok {
		return nil, fmt.Errorf("unexpected type of client conn for %s", clientName)
	}
	return &connMgr.peerInfo, nil
}

// GetReconnectCallbacks get all registered reconnect callback functions, return empty function slice on server side
func (wsp *WsServerProxy) GetReconnectCallbacks() []func() {
	return []func(){}
}

// GetDisconnectCallbacks get all registered disconnect callback functions
func (wsp *WsServerProxy) GetDisconnectCallbacks() []func(WebsocketPeerInfo) {
	return wsp.disConnCallbacks
}

// GetOnConnectCallbacks get all registered on connection callback functions
func (wsp *WsServerProxy) GetOnConnectCallbacks() []func(WebsocketPeerInfo) {
	return wsp.onConnCallbacks
}

// GetBandwidthLimiter Get bandwidth limiter instance, it not set, return nil
func (wsp *WsServerProxy) GetBandwidthLimiter() limiter.ServerBandwidthLimiterIntf {
	return wsp.bandwidthLimiter
}

// GetProxyConfig Get proxy config for server proxy
func (wsp *WsServerProxy) GetProxyConfig() *ProxyConfig {
	return wsp.ProxyCfg
}

// SetConnLimiter set WebSocket client conn limit for server
func (wsp *WsServerProxy) SetConnLimiter(maxConnLimit int) error {
	connLimiter, err := limiter.NewConnLimiter(maxConnLimit)
	if err != nil {
		return err
	}
	wsp.connLimiter = connLimiter
	return nil
}
