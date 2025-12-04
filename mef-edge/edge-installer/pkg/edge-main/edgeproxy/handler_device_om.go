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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/checker/msgchecker"
	"edge-installer/pkg/edge-main/common/msglistchecker"
	"edge-installer/pkg/edge-main/job"
	"edge-installer/pkg/edge-main/msgconv"
)

// DeviceOmProxy device-om conn proxy
type DeviceOmProxy struct {
	proxyBase

	conn               *websocket.Conn
	rateLimit          *limiter.RpsLimiter
	bandwidthLimit     *limiter.ClientBandwidthLimiter
	modName            string
	jobs               []ConnLoopJobFunc
	ctx                context.Context
	cf                 context.CancelFunc
	msgAdaptationProxy *msgconv.Proxy
}

// Start start device om proxy
func (dop *DeviceOmProxy) Start(conn *websocket.Conn) error {
	if conn == nil {
		return fmt.Errorf("the input of device om proxy is invalid")
	}
	dop.modName = constants.ModDeviceOm
	if err := RegistryConn(constants.ModDeviceOm, conn); err != nil {
		hwlog.RunLog.Errorf("register %s websocket connection error: %v", dop.modName, err)
		if err = conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close [%s] websocket error: %v", dop.modName, err)
		}
		return fmt.Errorf("register %s websocket connection error: %v", dop.modName, err)
	}
	hwlog.RunLog.Infof("client [%s] is connected", dop.modName)
	dop.conn = conn
	if err := dop.checkDeviceOmIp(); err != nil {
		dop.close()
		return err
	}
	rpsLimiter := limiter.NewRpsLimiter(constants.MsgRate, constants.BurstSize)
	dop.rateLimit = rpsLimiter

	dop.bandwidthLimit = limiter.NewClientBandwidthLimiter(&limiter.BandwidthLimiterConfig{
		MaxThroughput: constants.MaxMsgThroughput,
		Period:        constants.MsgThroughputPeriod,
	})
	dop.ctx, dop.cf = context.WithCancel(context.Background())
	dop.msgAdaptationProxy = &msgconv.Proxy{MessageSource: msgconv.Cloud, DispatchFunc: routeMsg}
	dop.lastAlive = time.Now()
	dop.registryJobs()
	ProcessJob(dop)
	select {
	case _, _ = <-dop.ctx.Done():
		dop.close()
	}
	return nil
}

func (dop *DeviceOmProxy) processMsg() error {
	msgData, err := dop.readMsg(dop.conn)
	if err != nil {
		return err
	}
	if dop.rateLimit == nil || !dop.rateLimit.Allow() {
		return nil
	}
	if !dop.bandwidthLimit.Allow(len(msgData)) {
		return nil
	}
	// both err and msgData are nil means message is not TextMessage, ignore it
	if msgData == nil {
		return nil
	}
	dop.handleMsg(msgData)
	return nil
}

// do business operation for message
func (dop *DeviceOmProxy) handleMsg(msg []byte) {
	var fdMsg model.Message
	if err := common.UnmarshalKubeedgeMessage(msg, &fdMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal %v message failed, error: %v", dop.modName, err)
		return
	}

	if !msglistchecker.NewFdMsgHeaderValidator().Check(&fdMsg) {
		hwlog.RunLog.Errorf("message is invalid, operation will be aborted, router: %+v  header:%+v",
			fdMsg.KubeEdgeRouter, fdMsg.Header)
		return
	}

	if err := common.UpdateFdAddrInfo(&fdMsg); err != nil {
		hwlog.RunLog.Errorf("UpdateFdAddrInfo error: %v, operation will be aborted", err)
		return
	}

	common.MsgOptLog(&fdMsg)

	msgValidator := msgchecker.NewMsgValidator(nil)
	if err := msgValidator.Check(&fdMsg); err != nil {
		fdMsg.Content = nil
		hwlog.RunLog.Errorf("check msg failed: %v", err)
		FeedbackError(err, "", &fdMsg)
		return
	}

	newMsg, err := common.MsgInProcess(&fdMsg)
	if err != nil {
		hwlog.RunLog.Errorf("MsgInProcess %v error: %v", dop.modName, err)
		return
	}
	hwlog.RunLog.Infof("[routeToEdge], route: %+v,  {ID: %s, parentID: %s}", fdMsg.KubeEdgeRouter,
		fdMsg.Header.ID, fdMsg.Header.ParentID)
	// response ASAP is this msg is response to a sync request
	if IsSyncMsgResp(newMsg) {
		if err := modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("response to %v error: %v", newMsg.GetParentId(), err)
		}
		return
	}

	if err := dop.msgAdaptationProxy.DispatchMessage(newMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed: %v", err)
	}
}

func routeMsg(newMsg *model.Message) error {
	dest, err := GetMsgDest(constants.ModDeviceOm, newMsg)
	if err != nil {
		hwlog.RunLog.Errorf("GetMsgDest error: %v", err)
		return fmt.Errorf("get msg dest error: %v", err)
	}
	switch dest.DestType {
	case MsgDestTypeModule:
		newMsg.Router.Destination = dest.DestName
		if err = modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("send message to module %v error: %v", dest.DestName, err)
			return err
		}
	case MsgDestTypeWs:
		if err := SendMsgToWs(newMsg, dest.DestName); err != nil {
			hwlog.RunLog.Errorf("send msg to websocket conn error: %v", err)
			return err
		}
	default:
		hwlog.RunLog.Errorf("unknown destination type: %v", dest.DestType)
		return errors.New("unknown dest type")
	}
	return nil
}

// FeedbackError [method] response error message to quest
func FeedbackError(err error, info string, msg *model.Message) {
	if msg == nil {
		return
	}

	const (
		metaManagerModuleName = "metaManager"
		defaultSendInterval   = 5 * time.Second
		defaultRetryTimes     = 3
	)

	var errInfo string
	if err != nil {
		errInfo = "FAILED: " + info + err.Error()
	} else {
		errInfo = "FAILED: " + info
	}

	resp, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create response message failed, %s", err.Error())
		return
	}
	resp.SetKubeEdgeRouter(
		metaManagerModuleName,
		msg.KubeEdgeRouter.Group,
		constants.OptError,
		msg.KubeEdgeRouter.Resource,
	)
	resp.SetRouter(
		metaManagerModuleName,
		constants.ModDeviceOm,
		constants.OptResponse,
		msg.KubeEdgeRouter.Resource,
	)

	resp.Header.ID = resp.Header.Id
	resp.Header.ParentID = msg.Header.ID
	if err = resp.FillContent(errInfo); err != nil {
		hwlog.RunLog.Errorf("fill err info into content failed: %v", err)
		return
	}

	var sendError error
	for i := 0; i < defaultRetryTimes; i++ {
		if sendError = SendMsgToWs(resp, constants.ModDeviceOm); sendError == nil {
			hwlog.RunLog.Info("feedback error to cloud success")
			break
		}
		time.Sleep(defaultSendInterval)
	}
	if sendError != nil {
		hwlog.RunLog.Warnf("feedback error to cloud failed, error: %v", sendError)
	}
}

func (dop *DeviceOmProxy) heartbeatInit() error {
	dop.conn.SetPingHandler(func(appData string) error {
		dop.lastAlive = time.Now()
		if err := SendHeartbeatToPeer(websocket.PongMessage, appData, dop.modName); err != nil {
			hwlog.RunLog.Errorf("response heartbeat pong to %v error: %v", dop.modName, err)
		}
		return nil
	})
	dop.conn.SetPongHandler(func(appData string) error {
		dop.lastAlive = time.Now()
		return nil
	})
	dop.conn.SetCloseHandler(func(code int, text string) error {
		hwlog.RunLog.Infof("client %s close connection: %v", dop.modName, text)
		dop.cf()
		return nil
	})
	return nil
}

func (dop *DeviceOmProxy) heartbeat() error {
	if time.Now().Sub(dop.lastAlive) > defaultHeartbeatTimeout {
		hwlog.RunLog.Infof("%v heartbeat check timeout. check time:%v, last alive time: %v",
			dop.modName, time.Now().Format(time.RFC3339), dop.lastAlive.Format(time.RFC3339))
		return fmt.Errorf("client %s is lost", dop.modName)
	}
	return nil
}

func (dop *DeviceOmProxy) close() {
	defer hwlog.RunLog.Infof("client [%v] is disconnected", dop.modName)
	if err := UnRegistryConn(dop.modName); err != nil {
		hwlog.RunLog.Errorf("unregister %v websocket connection error: %v", dop.modName, err)
	}
	if dop.conn == nil {
		return
	}
	if err := dop.conn.Close(); err != nil {
		hwlog.RunLog.Errorf("close [%v] websocket connection error: %v", dop.modName, err)
	}
	if dop.rateLimit != nil {
		dop.rateLimit = nil
	}
	if dop.bandwidthLimit != nil {
		dop.bandwidthLimit.Stop()
	}
}

func (dop *DeviceOmProxy) makeSureSendMsgToEdgeOm() error {
	for {
		err := dop.publishConnectInfo()
		if err == nil {
			break
		}
		hwlog.RunLog.Errorf("publish msg failed: %v", err)
		time.Sleep(constants.StartWsWaitTime)
	}
	return nil
}

func (dop *DeviceOmProxy) publishConnectInfo() error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message failed, error: %v", err)
		return err
	}
	sendMsg.SetRouter("", constants.ModEdgeOm, constants.OptReport, constants.DeviceOmConnectMsg)
	if err = sendMsg.FillContent(true); err != nil {
		hwlog.RunLog.Errorf("fill result into content failed: %v", err)
		return errors.New("fill result into content failed")
	}
	return modulemgr.SendMessage(sendMsg)
}

func (dop *DeviceOmProxy) registryJobs() {
	// allocate new slice each time, DO NOT use previous allocated memory
	dop.jobs = make([]ConnLoopJobFunc, 0)
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: neverStop,
		do:       dop.processMsg,
	})
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       dop.heartbeatInit,
	})
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: heartbeatCheckInterval,
		do:       dop.heartbeat,
	})
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: job.PodStatusInterval,
		do:       job.SyncPodStatus,
	})
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       dop.makeSureSendMsgToEdgeOm,
	})
	dop.jobs = append(dop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       job.SyncNodeStatus,
	})
}

// GetJobProxy Get jobs and websocket connection
func (dop *DeviceOmProxy) GetJobProxy() *JobProxy {
	return &JobProxy{
		conn: dop.conn,
		jobs: dop.jobs,
		ctx:  dop.ctx,
		cf:   dop.cf,
	}
}

func (dop *DeviceOmProxy) checkDeviceOmIp() error {
	deviceOmIpPort := dop.conn.RemoteAddr().String()
	ipAndPort := strings.Split(deviceOmIpPort, constants.IpPortSeparator)
	if len(ipAndPort) != constants.IpPortSliceLen {
		hwlog.RunLog.Error("device om ip port slice length failed")
		return errors.New("device om ip port slice length failed")
	}

	if ipAndPort[0] != constants.LocalIp {
		hwlog.RunLog.Error("device om ip is wrong")
		return errors.New("device om ip is wrong")
	}
	return nil
}
