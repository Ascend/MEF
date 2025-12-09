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

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/configpara"
)

// EdgeOmProxy msg do
type EdgeOmProxy struct {
	proxyBase

	conn      *websocket.Conn
	rateLimit *limiter.RpsLimiter
	modName   string
	jobs      []ConnLoopJobFunc
	ctx       context.Context
	cf        context.CancelFunc
}

// Start start edge om proxy
func (eop *EdgeOmProxy) Start(conn *websocket.Conn) error {
	if conn == nil {
		return fmt.Errorf("the input of edge om proxy is invalid")
	}
	eop.modName = constants.ModEdgeOm
	if err := RegistryConn(constants.ModEdgeOm, conn); err != nil {
		hwlog.RunLog.Errorf("register %s websocket connection error: %v", eop.modName, err)
		if err = conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close [%s] websocket error: %v", eop.modName, err)
		}
		return fmt.Errorf("register %s websocket connection error: %v", eop.modName, err)
	}
	hwlog.RunLog.Infof("client [%s] is connected", eop.modName)
	eop.conn = conn
	rpsLimiter := limiter.NewRpsLimiter(constants.MsgRate, constants.BurstSize)
	eop.rateLimit = rpsLimiter

	eop.ctx, eop.cf = context.WithCancel(context.Background())
	eop.lastAlive = time.Now()
	eop.registryJobs()
	ProcessJob(eop)
	select {
	case _, _ = <-eop.ctx.Done():
		eop.close()
	}
	return nil
}

func (eop *EdgeOmProxy) processMsg() error {
	if eop.rateLimit == nil || !eop.rateLimit.Allow() {
		return nil
	}
	msgData, err := eop.readMsg(eop.conn)
	if err != nil {
		return err
	}
	// both err and msgData are nil means message is not TextMessage, ignore it
	if msgData == nil {
		return nil
	}
	eop.handleMsg(msgData)
	return nil
}

// do business operation for message
func (eop *EdgeOmProxy) handleMsg(msg []byte) {
	var innerMsg model.Message
	if err := json.Unmarshal(msg, &innerMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal %v message failed, error: %v", eop.modName, err)
		return
	}
	newMsg, err := common.MsgInProcess(&innerMsg)
	if err != nil {
		hwlog.RunLog.Errorf("MsgInProcess %v error: %v", eop.modName, err)
		return
	}
	if isResStatic(newMsg) {
		handleResStaticMsg(newMsg)
		return
	}
	// response ASAP is this msg is response to a sync request
	if IsSyncMsgResp(newMsg) {
		if err = modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("response to %v error: %v", newMsg.GetParentId(), err)
		}
		return
	}
	dest, err := GetMsgDest(constants.ModEdgeOm, newMsg)
	if err != nil {
		hwlog.RunLog.Errorf("GetMsgDest error: %v", err)
		return
	}
	switch dest.DestType {
	case MsgDestTypeModule:
		newMsg.Router.Destination = dest.DestName
		if err = modulemgr.SendMessage(newMsg); err != nil {
			hwlog.RunLog.Errorf("send message to module %v error: %v", dest.DestName, err)
			return
		}
	case MsgDestTypeWs:
		if err := SendMsgToWs(newMsg, dest.DestName); err != nil {
			hwlog.RunLog.Errorf("send msg to websocket conn error: %v", err)
			return
		}
	default:
		hwlog.RunLog.Errorf("Unknown dest type: %v", dest.DestType)
		return
	}
}

func (eop *EdgeOmProxy) heartbeat() error {
	if time.Now().Sub(eop.lastAlive) > defaultHeartbeatTimeout {
		hwlog.RunLog.Infof("%v heartbeat check timeout. check time:%v, last alive time: %v",
			eop.modName, time.Now().Format(time.RFC3339), eop.lastAlive.Format(time.RFC3339))
		return fmt.Errorf("client %s is lost", eop.modName)
	}
	pingMsg := fmt.Sprintf("ping message from %v to %v", constants.ModEdgeProxy, eop.modName)
	if err := SendHeartbeatToPeer(websocket.PingMessage, pingMsg, eop.modName); err != nil {
		hwlog.RunLog.Errorf("send heartbeat ping to %v error: %v", eop.modName, err)
	}
	return nil
}

func (eop *EdgeOmProxy) heartbeatInit() error {
	eop.conn.SetPingHandler(func(appData string) error {
		eop.lastAlive = time.Now()
		if err := SendHeartbeatToPeer(websocket.PongMessage, appData, eop.modName); err != nil {
			hwlog.RunLog.Errorf("response heartbeat pong to %v error: %v", eop.modName, err)
		}
		return nil
	})
	eop.conn.SetPongHandler(func(appData string) error {
		eop.lastAlive = time.Now()
		return nil
	})
	eop.conn.SetCloseHandler(func(code int, text string) error {
		hwlog.RunLog.Infof("client %s close connection: %v", eop.modName, text)
		eop.cf()
		return nil
	})
	return nil
}

func (eop *EdgeOmProxy) syncCfg(cfgType string) error {
	cfgMap := map[string]interface{}{
		constants.InstallerConfigKey: &config.InstallerConfig{},
		constants.NetMgrConfigKey:    &config.NetManager{},
		constants.PodCfgResource:     &config.PodConfig{},
		constants.EdgeOmCapabilities: &config.StaticInfo{},
	}

	cfg, ok := cfgMap[cfgType]
	if !ok {
		hwlog.RunLog.Errorf("config %s not support", cfgType)
		return fmt.Errorf("config %s not support", cfgType)
	}

	for {
		select {
		case <-eop.ctx.Done():
			return fmt.Errorf("get config %s failed, because context is done", cfgType)
		default:

		}

		var err error
		var data string
		if data, err = eop.queryCfgFromEdgeOM(cfgType); err != nil {
			hwlog.RunLog.Errorf("get %s config from %s failed:%v", cfgType, eop.modName, err)
			time.Sleep(constants.WsSycMsgRetryInterval)
			continue
		}

		if err = json.Unmarshal([]byte(data), cfg); err != nil {
			hwlog.RunLog.Errorf("unmarshal message error: %s", err.Error())
			time.Sleep(constants.WsSycMsgRetryInterval)
			continue
		}
		configpara.SetCfgPara(cfg)
		break
	}
	hwlog.RunLog.Infof("get %s config success", cfgType)

	return nil
}

func (eop *EdgeOmProxy) syncPodConfig() error {
	return eop.syncCfg(constants.PodCfgResource)
}

func (eop *EdgeOmProxy) syncNetMngConfig() error {
	return eop.syncCfg(constants.NetMgrConfigKey)
}

func (eop *EdgeOmProxy) syncInstallerConfig() error {
	return eop.syncCfg(constants.InstallerConfigKey)
}

func (eop *EdgeOmProxy) syncCapabilities() error {
	return eop.syncCfg(constants.EdgeOmCapabilities)
}

func (eop *EdgeOmProxy) queryCfgFromEdgeOM(resourceType string) (string, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return "", fmt.Errorf("create model message error: %s", err.Error())
	}
	msg.SetRouter(
		"",
		constants.ModEdgeOm,
		constants.OptGet,
		constants.ResConfig,
	)
	if err = msg.FillContent(resourceType); err != nil {
		return "", fmt.Errorf("fill resource type into content failed: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, constants.WsSycMsgWaitTime)
	if err != nil {
		return "", fmt.Errorf("send sync message error: %s", err.Error())
	}

	var data string
	if err = resp.ParseContent(&data); err != nil {
		return "", fmt.Errorf("get resp content failed: %v", err)
	}
	if data == constants.Failed {
		return "", fmt.Errorf("msg content is error")
	}

	return data, nil
}

func (eop *EdgeOmProxy) close() {
	defer hwlog.RunLog.Infof("client [%v] is disconnected", eop.modName)
	if err := UnRegistryConn(eop.modName); err != nil {
		hwlog.RunLog.Errorf("unregister %v websocket connection error: %v", eop.modName, err)
	}
	if eop.conn == nil {
		return
	}
	if err := eop.conn.Close(); err != nil {
		hwlog.RunLog.Errorf("close [%v] websocket connection error: %v", eop.modName, err)
	}
	eop.rateLimit = nil
}

func (eop *EdgeOmProxy) reportAlarm() error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message failed, error: %s", err.Error())
		return err
	}
	sendMsg.SetRouter(constants.AlarmManager, constants.ModEdgeOm, constants.OptReport, constants.ReportAlarmMsg)
	if err = sendMsg.FillContent(true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return errors.New("fill content failed")
	}
	return modulemgr.SendAsyncMessage(sendMsg)
}

func (eop *EdgeOmProxy) registryJobs() {
	// allocate new slice each time, DO NOT use previous allocated memory
	eop.jobs = make([]ConnLoopJobFunc, 0)
	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: neverStop,
		do:       eop.processMsg,
	})
	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       eop.heartbeatInit,
	})
	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: defaultHeartbeatDuration,
		do:       eop.heartbeat,
	})

	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       eop.syncPodConfig,
	})

	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       eop.syncNetMngConfig,
	})

	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       eop.syncInstallerConfig,
	})

	eop.jobs = append(eop.jobs, ConnLoopJobFunc{
		interval: doOneTime,
		do:       eop.syncCapabilities,
	})

}

// GetJobProxy Get jobs and websocket connection
func (eop *EdgeOmProxy) GetJobProxy() *JobProxy {
	return &JobProxy{
		conn: eop.conn,
		jobs: eop.jobs,
		ctx:  eop.ctx,
		cf:   eop.cf,
	}
}

func isResStatic(msg *model.Message) bool {
	if msg == nil {
		return false
	}
	return msg.GetResource() == constants.ResStatic
}

func handleResStaticMsg(msg *model.Message) {
	var contentStr string
	if err := msg.ParseContent(&contentStr); err != nil {
		hwlog.RunLog.Errorf("get res content failed: %v", err)
		return
	}
	var info config.StaticInfo
	err := json.Unmarshal([]byte(contentStr), &info)
	if err != nil {
		hwlog.RunLog.Errorf("content type not right： %s", err.Error())
		return
	}
	config.GetCapabilityCache().SetEdgeOmCaps(info)
	// this msg is to fd
	if msg.GetOption() == constants.OptUpdate {
		config.GetCapabilityCache().Notify()
	}
}
