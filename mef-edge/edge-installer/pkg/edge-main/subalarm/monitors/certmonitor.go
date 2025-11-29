// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package monitors for cert monitor
package monitors

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

var (
	certOverdueThreshold int
	certTask             = &cronTask{
		alarmId:         almutils.CertAbnormal,
		name:            certMonitorName,
		checkStatusFunc: isCertOverdue,
	}
)

const (
	certMonitorName       = "cert"
	mefCertImportPathName = "hub_certs_import"
	rootCertName          = "cloud_root.crt"
)

func isCertOverdue() error {
	hwlog.RunLog.Infof("cert [%s] overdue time check start...", rootCertName)
	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get component config dir failed, error: %v", err)
		return errors.New("get component config dir failed")
	}

	certPath := filepath.Join(cfgDir, mefCertImportPathName, rootCertName)
	certData, err := fileutils.LoadFile(certPath)
	if err != nil {
		hwlog.RunLog.Errorf("load cert [%s] failed: %v", rootCertName, err)
		return fmt.Errorf("load cert [%s] failed", rootCertName)
	}

	if err = x509.CheckCertsOverdue(certData, certOverdueThreshold); err != nil {
		hwlog.RunLog.Errorf("check cert [%s] overdue failed, error: %v", rootCertName, err)
		return fmt.Errorf("check cert [%s] overdue failed", rootCertName)
	}

	hwlog.RunLog.Infof("cert [%s] overdue time check pass", rootCertName)
	return nil
}

func getAlarmCertCfg() (*config.AlarmCertCfg, error) {
	respContent, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:      constants.AlarmManager,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptGet,
		Resource:    constants.ResConfig,
		Content:     constants.AlarmCertConfig,
	})
	if err != nil {
		hwlog.RunLog.Errorf("send get alarm config message to edge om failed, error: %v", err)
		return nil, errors.New("send get alarm config message to edge om failed")
	}

	if respContent == constants.Failed {
		hwlog.RunLog.Error("resp content is failed, edge-om get alarm cert config failed")
		return nil, errors.New("edge-om get alarm cert config failed failed")
	}
	var cfg config.AlarmCertCfg
	if err = json.Unmarshal([]byte(respContent), &cfg); err != nil {
		hwlog.RunLog.Errorf("unmarshal resp content failed, error: %v", err)
		return nil, errors.New("unmarshal resp failed")
	}

	return &cfg, nil
}

// GetMEFMonitorList get mef monitor list, only cert alarm support currently.
func GetMEFMonitorList() []almutils.AlarmMonitor {
	for i := 0; i < constants.TryConnectNet; i++ {
		cfg, err := getAlarmCertCfg()
		if err != nil || cfg.CheckPeriod == 0 {
			time.Sleep(constants.StartWsWaitTime)
			continue
		}

		certTask.interval = time.Duration(cfg.CheckPeriod) * constants.Day
		certOverdueThreshold = cfg.OverdueThreshold
		return []almutils.AlarmMonitor{
			certTask,
		}
	}

	hwlog.RunLog.Error("get alarm cert config failed, reached the maximum number of the attempts")
	return []almutils.AlarmMonitor{}
}
