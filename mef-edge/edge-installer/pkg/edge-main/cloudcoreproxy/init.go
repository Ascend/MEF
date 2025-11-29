// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package cloudcoreproxy
package cloudcoreproxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/checker/msgchecker"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
	"edge-installer/pkg/edge-main/msgconv"
)

var wsWriteError = errors.New("websocket-connection write message error")

const (
	handshakeTimeout    = 10 * time.Second
	connectInterval     = 5 * time.Second
	defaultReadLimit    = 1.5 * 1024 * 1024
	readBufferSize      = 1024
	writeBufferSize     = 1024
	defaultProjectID    = "e632aba927ea4ac2b575ec1603d56f10"
	retryWaitTime       = 300
	waitTimeSecond      = 1 * time.Second
	restartIntervalTime = 2 * time.Second
	cloudCoreMsgRate    = 10
	cloudCoreBurstSize  = 20
)

// cloudCoreProxy cloudcore client proxy
type cloudCoreProxy struct {
	ctx                          context.Context
	cancel                       context.CancelFunc
	wsConnect                    *websocket.Conn
	waitGroup                    sync.WaitGroup
	rateLimit                    *limiter.RpsLimiter
	bandwidthLimit               *limiter.ClientBandwidthLimiter
	upstreamMsgAdaptationProxy   *msgconv.Proxy
	downstreamMsgAdaptationProxy *msgconv.Proxy
	enable                       bool
}

// NewCloudCoreProxy cloud core client proxy
func NewCloudCoreProxy(enable bool) model.Module {
	module := &cloudCoreProxy{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	module.upstreamMsgAdaptationProxy = &msgconv.Proxy{MessageSource: msgconv.Edge, DispatchFunc: module.SendToCloud}
	module.downstreamMsgAdaptationProxy = &msgconv.Proxy{
		MessageSource: msgconv.Cloud, DispatchFunc: modulemgr.SendMessage}
	return module
}

// Name module name
func (c *cloudCoreProxy) Name() string {
	return constants.ModCloudCore
}

// Enable module enable
func (c *cloudCoreProxy) Enable() bool {
	return c.enable
}

// Stop module stop
func (c *cloudCoreProxy) Stop() bool {
	c.cancel()
	if c.bandwidthLimit != nil {
		c.bandwidthLimit.Stop()
	}
	return true
}

func getWsCertInfo() (*certutils.TlsCertInfo, error) {
	cfgPath, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get client cert path failed, error: %v", err)
		return nil, errors.New("get client cert path failed")
	}

	rootCaFile := filepath.Join(cfgPath, constants.CloudCoreCertPathName, constants.RootCaName)
	certFile := filepath.Join(cfgPath, constants.CloudCoreCertPathName, constants.ClientCertName)
	keyFile := filepath.Join(cfgPath, constants.CloudCoreCertPathName, constants.ClientKeyName)

	if _, err = x509.CheckCertsChainReturnContent(rootCaFile); err != nil {
		hwlog.RunLog.Errorf("check cloud core root ca failed: %s", err.Error())
		return nil, errors.New("check cloud core root ca failed")
	}

	hwlog.RunLog.Info("websocket start to connect access")
	kmcCfg, err := util.GetKmcConfig("")
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config error: %v", err)
		return nil, errors.New("get kmc config failed")
	}

	return &certutils.TlsCertInfo{
		RootCaPath: rootCaFile,
		CertPath:   certFile,
		KeyPath:    keyFile,
		KmcCfg:     kmcCfg,
		WithBackup: true,
	}, nil
}

func (c *cloudCoreProxy) connectToCloud() error {
	var tlsCfg *tls.Config
	var err error
	var certInfo *certutils.TlsCertInfo
	if certInfo, err = getWsCertInfo(); err != nil {
		return err
	}
	if tlsCfg, err = certutils.GetTlsCfgWithPath(*certInfo); err != nil {
		return err
	}

	dialer := &websocket.Dialer{
		TLSClientConfig:  tlsCfg,
		HandshakeTimeout: handshakeTimeout,
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
	}

	netConfig := configpara.GetNetConfig()
	if netConfig.IP == "" || netConfig.Port == 0 {
		return fmt.Errorf("ip is invalid")
	}

	sn := configpara.GetInstallerConfig().SerialNumber
	if sn == "" {
		return fmt.Errorf("sn is invalid")
	}

	var headers = http.Header{}
	headers.Set("node_id", strings.ToLower(sn))
	headers.Set("project_id", defaultProjectID)
	headers.Set("ConnectionUse", "msg")
	serverAddr := fmt.Sprintf("wss://%s:%d/%s/%s/events", netConfig.IP,
		constants.DefaultCloudCoreWsPort, defaultProjectID, strings.ToLower(sn))
	hwlog.RunLog.Info("websocket client try to connect server")

	var connect *websocket.Conn
	for i := 0; i < constants.DefaultTryCount; i++ {
		connect, _, err = dialer.Dial(serverAddr, headers)
		if err == nil {
			hwlog.RunLog.Infof("dial %s successfully", serverAddr)
			break
		}

		hwlog.RunLog.Errorf("Init websocket connection failed %s", err.Error())

		time.Sleep(connectInterval)
	}
	if connect == nil {
		return errors.New("max retry count reached when connecting to cloud")
	}

	connect.SetReadLimit(defaultReadLimit)
	c.wsConnect = connect

	return nil
}

// Start cloud core client proxy
func (c *cloudCoreProxy) Start() {
	time.Sleep(constants.StartWsWaitTime)
	hwlog.RunLog.Info("-----------start cloud core client proxy-------------")

	if !c.checkEdgeHubIsReady() {
		hwlog.RunLog.Error("the edge hub not ready, can not start cloud core")
		return
	}

	certMng, err := newCertManager()
	if err != nil {
		hwlog.RunLog.Errorf("new cert manager error:%v", err)
		return
	}

	if err := c.initLimiters(); err != nil {
		hwlog.RunLog.Errorf("init cloudCore proxy limiters error:%v", err)
		return
	}

	c.run(certMng)
}

func (c *cloudCoreProxy) run(certMng certManager) {
	for {
		select {
		case <-c.ctx.Done():
			c.rateLimit = nil
			hwlog.RunLog.Warn("cloud core proxy stop")
			return
		default:
		}
		if !c.checkEdgeHubIsReady() {
			hwlog.RunLog.Error("the edge hub not ready, can not start cloud core")
			return
		}
		if err := certMng.start(); err != nil {
			hwlog.RunLog.Errorf("start cert manager error:%v", err)
			return
		}

		if err := c.connectToCloud(); err != nil {
			hwlog.RunLog.Errorf("connect cloud failed:%v", err)
			time.Sleep(constants.StartWsWaitTime)
			continue
		}
		common.ConnFlagCloudcore = true
		const goroutineCount = 2
		c.waitGroup.Add(goroutineCount)
		ctx, cancel := context.WithCancel(c.ctx)
		go c.routeToCloud(ctx, cancel)
		go c.routeToEdge(ctx, cancel)

		c.waitGroup.Wait()
		if err := c.wsConnect.Close(); err != nil {
			hwlog.RunLog.Warnf("failed to close cloudcore connection, error: %v", err)
		}
		common.ConnFlagCloudcore = false
		hwlog.RunLog.Warn("connection is broken, will reconnect to cloud core")
		time.Sleep(constants.StartWsWaitTime)

		time.Sleep(restartIntervalTime)
	}
}

func (c *cloudCoreProxy) initLimiters() error {
	rpsLimiter := limiter.NewRpsLimiter(cloudCoreMsgRate, cloudCoreBurstSize)
	c.rateLimit = rpsLimiter

	c.bandwidthLimit = limiter.NewClientBandwidthLimiter(&limiter.BandwidthLimiterConfig{
		MaxThroughput: constants.MaxMsgThroughput,
		Period:        constants.MsgThroughputPeriod,
	})
	return nil
}

func (c *cloudCoreProxy) checkEdgeHubIsReady() bool {
	for i := 0; i < retryWaitTime; i++ {
		if common.ConnFlagEdgehub {
			return true
		}
		hwlog.RunLog.Info("wait the edge hub is ready ...")
		time.Sleep(waitTimeSecond)
	}
	return false
}

func (c *cloudCoreProxy) routeToCloud(ctx context.Context, cancelFunc context.CancelFunc) {
	defer func() {
		c.waitGroup.Done()
		cancelFunc()
	}()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Error("some error occurred, goroutine exited")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(c.Name())
		if err != nil {
			hwlog.RunLog.Errorf("cloud core client proxy receive module message failed: %v", err)
			continue
		}
		hwlog.RunLog.Infof("[routeToCloud], route: %+v, {ID: %s, parentID: %s}", msg.KubeEdgeRouter,
			msg.Header.ID, msg.Header.ParentID)

		common.MEFOpLogWithRes(msg)

		if !msglistchecker.NewCloudCoreMsgHeaderValidator(true).Check(msg) {
			hwlog.RunLog.Errorf("[routeToCloud] error: upstream msg is not allowed, "+
				"route: %+v, {ID: %s, parentID: %s}", msg.KubeEdgeRouter, msg.Header.ID, msg.Header.ParentID)
			continue
		}
		if err = c.upstreamMsgAdaptationProxy.DispatchMessage(msg); err != nil {
			hwlog.RunLog.Errorf("send data to cloud core failed: %v", err)
			if errors.Is(err, wsWriteError) {
				return
			}
		}
	}
}

func (c *cloudCoreProxy) SendToCloud(msg *model.Message) error {
	data, err := common.MarshalKubeedgeMessage(msg)
	if err != nil {
		hwlog.RunLog.Errorf("marshal msg failed, error: %v", err)
		return err
	}
	if err := c.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
		hwlog.RunLog.Errorf("write msg to cloud failed, %v", err)
		return wsWriteError
	}
	return nil
}

func (c *cloudCoreProxy) Receive() ([]byte, error) {
	messageType, reader, err := c.wsConnect.NextReader()
	if err != nil {
		return nil, err
	}
	if messageType != websocket.TextMessage {
		return nil, errors.New("msg type is not correct")
	}

	data, err := io.ReadAll(io.LimitReader(reader, defaultReadLimit))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *cloudCoreProxy) routeToEdge(ctx context.Context, cancelFunc context.CancelFunc) {
	defer func() {
		c.waitGroup.Done()
		cancelFunc()
	}()

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Error("some error occurred, goroutine exited")
			return
		default:
		}

		data, err := c.Receive()
		if err != nil {
			hwlog.RunLog.Errorf("receive message failed, error: %v", err)
			return
		}

		if !c.bandwidthLimit.Allow(len(data)) {
			hwlog.RunLog.Error("tps check failed, receive too much data")
			continue
		}

		var msg model.Message
		if err = common.UnmarshalKubeedgeMessage(data, &msg); err != nil {
			hwlog.RunLog.Errorf("unmarshal message failed, error: %v", err)
			continue
		}

		if c.rateLimit == nil || !c.rateLimit.Allow() {
			hwlog.RunLog.Warnf("msg [router: %+v, header: %+v] count up to limit per second",
				msg.KubeEdgeRouter, msg.Header)
			continue
		}

		hwlog.RunLog.Infof("[routeToEdge], route: %+v, {ID: %s, parentID: %s}", msg.KubeEdgeRouter,
			msg.Header.ID, msg.Header.ParentID)

		msgValidator := msgchecker.NewMsgValidator(msglistchecker.NewCloudCoreMsgHeaderValidator(false))
		if err = msgValidator.Check(&msg); err != nil {
			msg.Content = nil
			hwlog.RunLog.Errorf("check msg failed: %v", err)
			continue
		}

		common.MEFOpLog(&msg)

		msg.Router.Destination = constants.ModEdgeCore
		if err = c.downstreamMsgAdaptationProxy.DispatchMessage(&msg); err != nil {
			hwlog.RunLog.Errorf("send message to module %v error: %v", "Edge", err)
		}
	}
}
