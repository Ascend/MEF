// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

type wsSvrMessage struct {
	Msg        *wsMessage
	ClientName string
}

// WsServerProxy websocket server proxy
type WsServerProxy struct {
	ProxyCfg    *ProxyConfig
	httpServer  *http.Server
	clientMap   sync.Map
	upgrade     *websocket.Upgrader
	connMgr     *wsConnectMgr
	ClientNum   int
	CounterLock sync.Mutex
	handlerMap  sync.Map
}

// GetName get websocket server name
func (wsp *WsServerProxy) GetName() string {
	return wsp.ProxyCfg.name
}

// Start websocket server start
func (wsp *WsServerProxy) Start() error {
	httpServer := &http.Server{
		Addr:      wsp.ProxyCfg.hosts,
		TLSConfig: wsp.ProxyCfg.tlsConfig,
	}
	wsp.httpServer = httpServer
	http.HandleFunc(svcUrl, wsp.serveHTTP)
	wsp.initHandlers()
	wsp.upgrade = &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	go func() {
		err := wsp.listen()
		if err != nil {
			return
		}
	}()
	return nil
}

// Stop websocket server stop
func (wsp *WsServerProxy) Stop() error {
	wsp.ProxyCfg.cancel()
	wsp.clientMap.Range(wsp.closeOneClient)
	err := wsp.httpServer.Close()
	if err != nil {
		return fmt.Errorf("stop websocket server failed, error: %v", err)
	}
	return nil
}

// Send websocket server send message
func (wsp *WsServerProxy) Send(msg interface{}) error {
	wsMsg, ok := msg.(wsSvrMessage)
	if !ok {
		return fmt.Errorf("websocket sever send failed: the message type [%T] unsupported", msg)
	}
	clientName := wsMsg.ClientName
	cltConnMgr, ok := wsp.clientMap.Load(clientName)
	if !ok {
		return fmt.Errorf("websocket sever send failed: the client [%v] not connect", clientName)
	}
	connMgr, ok := cltConnMgr.(*wsConnectMgr)
	if !ok {
		return fmt.Errorf("websocket sever send failed: the connect manager type [%T] unsupported", cltConnMgr)
	}
	return connMgr.send(*wsMsg.Msg)
}

func (wsp *WsServerProxy) closeOneClient(name, conn interface{}) bool {
	wsp.clientMap.Delete(name)
	wsConn, ok := conn.(wsConnectMgr)
	if !ok {
		hwlog.RunLog.Errorf("close client [%v] failed: conn[%T] not a valid conn", name, conn)
		return true
	}
	err := wsConn.stop()
	if err != nil {
		hwlog.RunLog.Errorf("close client [%v] failed, error: %v", name, err)
	}
	hwlog.RunLog.Infof("client [name=%v] disconnect", name)

	return true
}

func (wsp *WsServerProxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		hwlog.RunLog.Errorf("it is not a websocket request: %v", r.RemoteAddr)
		return
	}
	if ok := CheckAndAddClientNum(wsp); !ok {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(common.ErrorMap[common.ErrorMaxEdgeClientsReached])); err != nil {
			hwlog.RunLog.Errorf("write response to mef edge error: %v", err)
		}
		return
	}
	clientName := r.Header.Get(clientNameKey)
	conn, err := wsp.upgrade.Upgrade(w, r, nil)
	if err != nil {
		RemoveClientNum(wsp)
		hwlog.RunLog.Errorf("websocket start server http failed, error: %v", err)
		return
	}
	connMgr := &wsConnectMgr{}
	connMgr.start(conn, clientName, &wsp.ProxyCfg.handlerMgr)
	wsp.clientMap.Store(clientName, connMgr)
	hwlog.RunLog.Infof("client [name=%v, addr=%v] connect", clientName, r.RemoteAddr)
	select {
	case <-connMgr.ctx.Done():
		RemoveClientNum(wsp)
	}
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
			hwlog.RunLog.Errorf("websocket listen and serve with tls failed, error: %v", err)
		}
		time.Sleep(retryTime)
	}
}

// AddHandler set customized url and handler
func (wsp *WsServerProxy) AddHandler(url string, handler func(http.ResponseWriter, *http.Request)) error {
	if url == "" || url == "/" || handler == nil {
		return fmt.Errorf("invalid http url or handler")
	}
	if _, existed := wsp.handlerMap.LoadOrStore(url, handler); existed {
		return fmt.Errorf("the url [%v] is already registered", url)
	}
	return nil
}

// initHandlers must be called after http.Server initialized
func (wsp *WsServerProxy) initHandlers() {
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
		http.HandleFunc(urlPath, handlerFunc)
		return true
	})
}
