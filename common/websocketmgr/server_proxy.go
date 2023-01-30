// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
)

type wsSvrMessage struct {
	Msg        *wsMessage
	ClientName string
}

// WsServerProxy websocket server proxy
type WsServerProxy struct {
	ProxyCfg   *ProxyConfig
	httpServer *http.Server
	clientMap  sync.Map
	upgrade    *websocket.Upgrader
	connMgr    *wsConnectMgr
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
	http.HandleFunc(serverPattern, wsp.serveHTTP)
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
		return fmt.Errorf("stop websocket server failed: %v", err)
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
		hwlog.RunLog.Errorf("close client [%v] failed: %v", name, err)
	}
	return true
}

func (wsp *WsServerProxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		hwlog.RunLog.Errorf("it is not a websocket request: %v", r.RemoteAddr)
		return
	}
	clientName := r.Header.Get(clientNameKey)
	conn, err := wsp.upgrade.Upgrade(w, r, nil)
	if err != nil {
		hwlog.RunLog.Errorf("websocket start server http failed: %v", err)
		return
	}
	connMgr := &wsConnectMgr{}
	connMgr.start(conn, clientName, &wsp.ProxyCfg.handlerMgr)
	wsp.clientMap.Store(clientName, connMgr)
	hwlog.RunLog.Infof("client [name=%v, addr=%v] connect", clientName, r.RemoteAddr)
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
		// todo 这里需要修改为tls的接口 ListenAndServeTLS
		err := wsp.httpServer.ListenAndServe()
		if err != nil {
			hwlog.RunLog.Errorf("websocket listen and serve with tls failed: %v", err)
		}
		time.Sleep(retryTime)
	}
}
