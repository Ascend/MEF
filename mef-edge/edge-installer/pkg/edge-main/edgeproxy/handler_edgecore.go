// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/msgconv"
)

// EdgecoreProxy msg do
type EdgecoreProxy struct {
	proxyBase

	conn               *websocket.Conn
	rateLimit          *limiter.RpsLimiter
	modName            string
	jobs               []ConnLoopJobFunc
	ctx                context.Context
	cf                 context.CancelFunc
	msgAdaptationProxy *msgconv.Proxy
}

// Start edge core proxy
func (ecp *EdgecoreProxy) Start(conn *websocket.Conn) error {
	if conn == nil {
		return fmt.Errorf("the input of edge core proxy is invalid")
	}
	ecp.modName = constants.ModEdgeCore
	if err := RegistryConn(constants.ModEdgeCore, conn); err != nil {
		hwlog.RunLog.Errorf("register %s websocket connection error: %v", ecp.modName, err)
		if err = conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close [%s] websocket error: %v", ecp.modName, err)
		}
		return fmt.Errorf("register %s websocket connection error: %v", ecp.modName, err)
	}
	hwlog.RunLog.Infof("client [%s] is connected", ecp.modName)
	ecp.conn = conn
	rpsLimiter := limiter.NewRpsLimiter(constants.MsgRate, constants.BurstSize)
	ecp.rateLimit = rpsLimiter

	ecp.ctx, ecp.cf = context.WithCancel(context.Background())
	ecp.msgAdaptationProxy = &msgconv.Proxy{MessageSource: msgconv.Edge, DispatchFunc: dispatchMsg}
	ecp.lastAlive = time.Now()
	ecp.registryJobs()
	ProcessJob(ecp)
	select {
	case _, _ = <-ecp.ctx.Done():
		ecp.close()
	}
	return nil
}

func (ecp *EdgecoreProxy) processMsg() error {
	if ecp.rateLimit == nil || !ecp.rateLimit.Allow() {
		return nil
	}
	msgData, err := ecp.readMsg(ecp.conn)
	if err != nil {
		return err
	}
	// both err and msgData are nil means message is not TextMessage, ignore it
	if msgData == nil {
		return nil
	}
	ecp.handleMsg(msgData)
	return nil
}

// do business operation for message
func (ecp *EdgecoreProxy) handleMsg(msg []byte) {
	var fdMsg model.Message
	if err := json.Unmarshal(msg, &fdMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal %v message failed, error: %v", ecp.modName, err)
		return
	}

	newMsg, err := common.MsgInProcess(&fdMsg)
	if err != nil {
		hwlog.RunLog.Errorf("MsgInProcess %v error: %v", ecp.modName, err)
		return
	}

	netType, err := configpara.GetNetType()
	if err != nil {
		hwlog.RunLog.Errorf("get net type failed: %v", err)
		return
	}

	if netType == constants.MEF {
		fdMsg.Router.Destination = constants.ModCloudCore
		if err = modulemgr.SendAsyncMessage(&fdMsg); err != nil {
			hwlog.RunLog.Errorf("send message to module %v error: %v", constants.ModCloudCore, err)
		}
		return
	}

	if IsInternalKubeedgeResp(newMsg) {
		if err = modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("response to %v error: %v", newMsg.GetParentId(), err)
		}
		return
	}
	if err := ecp.msgAdaptationProxy.DispatchMessage(newMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed: %v", err)
	}
}

func dispatchMsg(newMsg *model.Message) error {
	dest, err := GetMsgDest(constants.ModEdgeCore, newMsg)
	if err != nil {
		// not all messages from edgecore are needed, ignore not routed messages.
		return nil
	}
	switch dest.DestType {
	case MsgDestTypeModule:
		newMsg.Router.Destination = dest.DestName
		if err := modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("send message to module %v error: %v", dest.DestName, err)
			return err
		}
	case MsgDestTypeWs:
		if err := SendMsgToWs(newMsg, dest.DestName); err != nil {
			hwlog.RunLog.Errorf("send msg to websocket conn error: %v", err)
			return err
		}
	default:
		hwlog.RunLog.Errorf("Unknown dest type: %v", dest.DestType)
		return errors.New("unknown dest type")
	}
	return nil
}

func (ecp *EdgecoreProxy) heartbeatInit() error {
	ecp.conn.SetPingHandler(func(appData string) error {
		ecp.lastAlive = time.Now()
		if err := SendHeartbeatToPeer(websocket.PongMessage, appData, ecp.modName); err != nil {
			hwlog.RunLog.Errorf("response heartbeat pong to %v error: %v", ecp.modName, err)
		}
		return nil
	})
	ecp.conn.SetPongHandler(func(appData string) error {
		ecp.lastAlive = time.Now()
		return nil
	})
	ecp.conn.SetCloseHandler(func(code int, text string) error {
		hwlog.RunLog.Infof("client %s close connection: %v", ecp.modName, text)
		ecp.cf()
		return nil
	})
	return nil
}

func (ecp *EdgecoreProxy) close() {
	defer hwlog.RunLog.Infof("client [%v] is disconnected", ecp.modName)
	if err := UnRegistryConn(ecp.modName); err != nil {
		hwlog.RunLog.Errorf("unregister %v websocket connection error: %v", ecp.modName, err)
	}
	if ecp.conn == nil {
		return
	}
	if err := ecp.conn.Close(); err != nil {
		hwlog.RunLog.Errorf("close [%v] websocket connection error: %v", ecp.modName, err)
	}
	ecp.rateLimit = nil
}

func (ecp *EdgecoreProxy) registryJobs() {
	// allocate new slice each time, DO NOT use previous allocated memory
	ecp.jobs = make([]ConnLoopJobFunc, 0)
	ecp.jobs = append(ecp.jobs, ConnLoopJobFunc{
		interval: neverStop,
		do:       ecp.processMsg,
	})
	ecp.jobs = append(ecp.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       ecp.heartbeatInit,
	})
}

// GetJobProxy Get jobs and websocket connection
func (ecp *EdgecoreProxy) GetJobProxy() *JobProxy {
	return &JobProxy{
		conn: ecp.conn,
		jobs: ecp.jobs,
		ctx:  ecp.ctx,
		cf:   ecp.cf,
	}
}
