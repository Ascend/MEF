// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package edgehub this file for edge hub module register
package edgehub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

const (
	retryWaitTime     = 300
	reconnectInterval = 5 * time.Second
	waitTimeSecond    = 1 * time.Second
)

// edgeHub edge hub module
type edgeHub struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewEdgeHub new edge hub
func NewEdgeHub(enable bool) model.Module {
	module := &edgeHub{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (m *edgeHub) Name() string {
	return constants.ModEdgeHub
}

// Enable module enable
func (m *edgeHub) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *edgeHub) Stop() bool {
	m.cancel()
	return true
}

// Start module start running
func (m *edgeHub) Start() {
	// wait inner process ready before connect to cloud
	time.Sleep(constants.StartWsWaitTime)
	hwlog.RunLog.Info("-------------------edge hub start--------------------------")

	if !m.waitCfgReady() {
		hwlog.RunLog.Error("the config from edge om not ready, can not start edge hub")
		return
	}
	common.ConnFlagEdgehub = false
	if err := m.connect(); err != nil {
		hwlog.RunLog.Errorf("init edge hub ws client failed, stop edgehub: %v", err)
		return
	}
	common.ConnFlagEdgehub = true
	hwlog.RunLog.Info("init edge hub ws client success")

	ctx, cancel := context.WithCancel(context.Background())
	go m.checkCloudcoreIsConnected(ctx)
	m.setAndPublishConnStatus(true)

	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Error("-------------------edge hub exit--------------------------")
			return
		default:
		}
		receivedMsg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("edge hub get receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("[routeToCenter], route: %+v, {ID: %s, parentID: %s}", receivedMsg.Router,
			receivedMsg.Header.Id, receivedMsg.Header.ParentId)

		if err = m.dispatch(receivedMsg, cancel); err != nil {
			hwlog.RunLog.Errorf("edgehub send message [header: %+v, router: %+v] to mef-center failed, error: %v",
				receivedMsg.Header, receivedMsg.Router, err)
		}
	}
}

func (m *edgeHub) connect() error {
	netConfig, err := getNetConfigFromEdgeOM()
	if err != nil {
		return fmt.Errorf("get net config failed: %v", err)
	}
	defer utils.ClearSliceByteMemory(netConfig.Token)

	var i int
	for i = 0; i < constants.TryConnectNet; i++ {
		err := initClient(netConfig)
		if err == nil {
			break
		}
		// MEFEdge will be locked by MEFCenter if the token verification fails too many times.
		// Skip reconnection if we detect a token error.
		if errors.Is(err, errCloudhubAuth) {
			return errors.New("cloudhub auth failed")
		}
		hwlog.RunLog.Errorf("init edge hub ws client failed: %v", err)
		time.Sleep(constants.StartWsWaitTime)
	}
	if i == constants.TryConnectNet {
		return errors.New("has reached the maximum number of the connection attempts")
	}

	return nil
}

func (m *edgeHub) checkCloudcoreIsConnected(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("close check cloudcore is connected")
			return
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			if m.checkCloudcoreIsConn() {
				continue
			}
			hwlog.RunLog.Warn("cloudcore proxy not ready for a while, restart edgehub")
			if err := proxy.Stop(); err != nil {
				common.ConnFlagEdgehub = false
				hwlog.RunLog.Errorf("stop connection error: %v", err)
				continue
			}
			m.setAndPublishConnStatus(false)
			time.Sleep(reconnectInterval)
			if err := proxy.Start(); err != nil {
				hwlog.RunLog.Errorf("start connection error: %v", err)
				continue
			}
			m.setAndPublishConnStatus(true)
		}
	}
}
func (m *edgeHub) checkCloudcoreIsConn() bool {
	for i := 0; i < retryWaitTime; i++ {
		if common.ConnFlagCloudcore {
			return true
		}
		time.Sleep(waitTimeSecond)
	}
	return false
}

func (m *edgeHub) waitCfgReady() bool {
	for i := 0; i < retryWaitTime; i++ {
		if configpara.CheckCfgIsReady() {
			return true
		}
		hwlog.RunLog.Info("wait the get config is ready ...")
		time.Sleep(waitTimeSecond)
	}
	return false
}

func (m *edgeHub) dispatch(msg *model.Message, cancel context.CancelFunc) error {
	if isCertUpdateNotify(msg) {
		go func() {
			if err := doCertUpdate(msg); err != nil {
				hwlog.RunLog.Errorf("do edge-hub cert update error: %v", err)
			}
		}()
		return nil
	}
	if isCenterDeleteNode(msg) {
		if err := stopConn(cancel); err != nil {
			hwlog.OpLog.Errorf("[%v@%v][%v %v %v][msgId: %v]", configpara.GetNetConfig().NetType,
				configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Failed, msg.Header.Id)
			return err
		}
		hwlog.OpLog.Infof("[%v@%v][%v %v %v][msgId: %v]", configpara.GetNetConfig().NetType,
			configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Success, msg.Header.Id)
		return nil
	}
	if err := sendMsgToServer(msg); err != nil {
		hwlog.RunLog.Errorf("send message to server failed: %v", err)
		return errors.New("send message to server failed failed")
	}
	return nil
}

func getNetConfigFromEdgeOM() (*config.NetManager, error) {
	hwlog.RunLog.Info("start to get net config token from edge-om")
	var i int
	var netConfig *config.NetManager
	var err error
	for i = 0; i < constants.DefaultTryCount; i++ {
		netConfig, err = tryGetNetConfigFromEdgeOM()
		if err == nil {
			break
		}
		hwlog.RunLog.Errorf("get config failed: %v", err)
		time.Sleep(constants.StartWsWaitTime)
	}

	if i == constants.DefaultTryCount {
		return nil, errors.New("get net config from edge om failed, has reached max number of the connection attempts")
	}
	hwlog.RunLog.Info("get net config token from edge-om success")
	return netConfig, nil
}

func tryGetNetConfigFromEdgeOM() (*config.NetManager, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return nil, fmt.Errorf("create model message error: %s", err.Error())
	}
	msg.SetRouter(
		"",
		constants.ModEdgeOm,
		constants.OptGet,
		constants.ResConfig,
	)
	if err = msg.FillContent(constants.NetMgrConfigKey); err != nil {
		return nil, fmt.Errorf("fill resource type into content failed: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, constants.WsSycMsgWaitTime)
	if err != nil {
		return nil, fmt.Errorf("send sync message error: %s", err.Error())
	}
	var data []byte
	cfg := &config.NetManager{}
	if err = resp.ParseContent(&data); err != nil {
		return nil, fmt.Errorf("get resp content failed: %v", err)
	}
	defer utils.ClearSliceByteMemory(data)
	if err = json.Unmarshal(data, cfg); err != nil {
		return nil, errors.New("unmarshal message error")
	}
	return cfg, nil
}

func getConfig() (*config.NetManager, error) {
	netConfig := configpara.GetNetConfig()

	if netConfig.IP == "" {
		return nil, fmt.Errorf("ip is invalid")
	}

	return &netConfig, nil
}

func (m *edgeHub) setAndPublishConnStatus(status bool) {
	common.ConnFlagEdgehub = status
	go m.publishConnectCloudResult()
}

func (m *edgeHub) publishConnectCloudResult() {
	const syncConnectStatusInterval = 5 * time.Second
	for {
		if proxy.IsConnected() && m.publishConnectInfo() == nil {
			return
		}
		time.Sleep(syncConnectStatusInterval)
	}
}

func (m *edgeHub) publishConnectInfo() error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message failed for publishing connect info, error: %v", err)
		return err
	}
	sendMsg.SetRouter(m.Name(), constants.ModEdgeOm, constants.OptReport, constants.ResEdgeCloudConnection)
	if err = sendMsg.FillContent(common.ConnFlagEdgehub); err != nil {
		hwlog.RunLog.Errorf("fill result into content failed: %v", err)
		return errors.New("fill result into content failed")
	}

	resp, err := modulemgr.SendSyncMessage(sendMsg, constants.WsSycMsgWaitTime)
	if err != nil {
		hwlog.RunLog.Errorf("%s sends message to %s failed", m.Name(), constants.ModEdgeOm)
		return err
	}
	var respContent string
	if err := resp.ParseContent(&respContent); err != nil || respContent != constants.OK {
		hwlog.RunLog.Error("unknown response content for publishing connect info")
		return errors.New("unknown response content")
	}
	return nil
}

func isCertUpdateNotify(msg *model.Message) bool {
	if msg == nil {
		return false
	}
	return msg.GetOption() == constants.OptGet && msg.GetResource() == constants.ResCertUpdate
}

func isCenterDeleteNode(msg *model.Message) bool {
	if msg == nil {
		return false
	}
	return msg.GetOption() == constants.OptDelete && msg.GetResource() == constants.DeleteNodeMsg
}

func stopConn(cancel context.CancelFunc) error {
	common.ConnFlagEdgehub = false
	cancel()
	if err := proxy.Stop(); err != nil {
		hwlog.RunLog.Errorf("stop connection error: %v", err)
		return err
	}
	hwlog.RunLog.Info("stop connection")
	return nil
}

func sendMsgToServer(msg *model.Message) error {
	sender, err := getCltSender()
	if err != nil {
		return fmt.Errorf("edge hub get client sender failed: %v", err)
	}
	response, err := common.MsgOutProcess(msg)
	if err != nil {
		return err
	}

	if !msglistchecker.NewMefMsgHeaderValidator().Check(response) {
		hwlog.RunLog.Errorf("check msg header[%+v] router[%+v] failed", response.Header, response.KubeEdgeRouter)
		return errors.New("not allowed message to mef-center")
	}

	installConfig := configpara.GetInstallerConfig()
	response.SetNodeId(installConfig.SerialNumber)
	if err = sender.Send(response); err != nil {
		return fmt.Errorf("edge hub sender send message to mef-center failed: %v", err)
	}
	return nil
}
