// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

type getConfigHandler struct{}

// podConfig is used to response pod config query from edge-main
var podConfig config.PodConfig

// Handle getConfigHandler handle entry
func (ch *getConfigHandler) Handle(msg *model.Message) error {
	var content string
	if err := msg.ParseContent(&content); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return errors.New("parse content failed")
	}
	var resp string
	switch content {
	case constants.PodCfgResource:
		resp = ch.getPodConfig()
	case constants.NetMgrConfigKey:
		resp = ch.getNetMgrConfig()
		if resp != constants.Failed {
			// make sure token msg is cleared after sending response to avoid sending '0000' response
			defer time.AfterFunc(constants.WsSycMsgWaitTime, func() { utils.ClearStringMemory(resp) })
		}
	case constants.InstallerConfigKey:
		resp = ch.getInstallConfig()
	case constants.SoftwareCert:
		resp = ch.getSoftwareCert()
	case constants.AlarmCertConfig:
		resp = ch.getAlarmCertConfig()
	case constants.EdgeOmCapabilities:
		resp = ch.getEdgeOmCaps()
	default:
		resp = constants.Failed
	}
	ch.sendResponse(msg, resp)
	if resp != constants.Failed {
		hwlog.RunLog.Infof("successfully sending %s config response", content)
	}
	return nil
}

func (ch *getConfigHandler) sendResponse(msg *model.Message, resp string) {
	newResponse, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("get new response failed: %v", err)
		return
	}
	if err = newResponse.FillContent(resp); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return
	}
	err = sendHandlerReplyMsg(newResponse)
	if err != nil {
		hwlog.RunLog.Errorf("send config handler response failed: %v", err)
		return
	}
}

func (ch *getConfigHandler) getPodConfig() string {
	bytes, err := json.Marshal(podConfig)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data failed: %v", err)
		return constants.Failed
	}
	return string(bytes)
}

func (ch *getConfigHandler) getInstallConfig() string {
	dbMgr, err := ch.initDbMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init db mgr failed: %v", err)
		return constants.Failed
	}
	installConfig, err := config.GetInstall(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("get install config failed: %v", err)
		return constants.Failed
	}
	bytes, err := json.Marshal(installConfig)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data failed: %v", err)
		return constants.Failed
	}
	return string(bytes)
}

func (ch *getConfigHandler) getNetMgrConfig() string {
	dbMgr, err := ch.initDbMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init db mgr failed: %v", err)
		return constants.Failed
	}
	return getNetConfig(dbMgr)
}

func (ch *getConfigHandler) getSoftwareCert() string {
	hwlog.RunLog.Info("start to get cert for model file update")
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return constants.Failed
	}
	certPath := configPathMgr.GetImageCertPath()
	certContent, err := certutils.GetCertContentWithBackup(certPath)
	if err != nil {
		hwlog.RunLog.Errorf("get cert for model file update failed, error: %v", err)
		return constants.Failed
	}
	return string(certContent)
}

func (ch *getConfigHandler) getAlarmCertConfig() string {
	dbMgr, err := ch.initDbMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init db mgr failed: %v", err)
		return constants.Failed
	}

	period, err := dbMgr.GetAlarmConfig(constants.CertCheckPeriodDB)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm config cert check period failed: %v", err)
		return constants.Failed
	}
	threshold, err := dbMgr.GetAlarmConfig(constants.CertOverdueThresholdDB)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm config cert overdue threshold failed: %v", err)
		return constants.Failed
	}
	alarmCertCfg := config.AlarmCertCfg{
		CheckPeriod:      period,
		OverdueThreshold: threshold,
	}

	cfgBytes, err := json.Marshal(alarmCertCfg)
	if err != nil {
		hwlog.RunLog.Errorf("marshal alarm cert config failed: %v", err)
		return constants.Failed
	}
	return string(cfgBytes)
}

func (ch *getConfigHandler) getEdgeOmCaps() string {
	caps := config.StaticInfo{ProductCapabilityEdge: config.GetCapabilityMgr().GetCaps()}
	capsBytes, err := json.Marshal(caps)
	if err != nil {
		hwlog.RunLog.Errorf("marshal edge om caps failed: %v", err)
		return constants.Failed
	}
	return string(capsBytes)
}

func (ch *getConfigHandler) initDbMgr() (*config.DbMgr, error) {
	edgeOmCfg, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get config dir failed: %v", err)
		return &config.DbMgr{}, errors.New("get config dir failed")
	}
	return config.NewDbMgr(edgeOmCfg, constants.DbEdgeOmPath), nil
}
