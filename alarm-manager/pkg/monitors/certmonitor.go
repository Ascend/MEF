// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors defined cert monitor, include northbound cert, software repository cert, and image repository cert
package monitors

import (
	"errors"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	certMonitorName   = "cert"
	importedCertsNum  = 3
	getConfigTimes    = 5
	getConfigWaitTime = 5 * time.Second
)

var (
	certTask = &cronTask{
		name:      certMonitorName,
		resetFunc: certReset,
	}

	certOverdueThreshold int
	getCertsInfoFlag     bool
	importedCertsInfo    requests.ImportedCertsInfo
)

func registerCertMonitor(dbPath string) error {
	certTask.alarmIdFuncMap = make(map[string]func() error, importedCertsNum)
	certTask.alarmIdFuncMap[NorthCertAbnormal] = isNorthCertOverdue
	certTask.alarmIdFuncMap[SoftwareCertAbnormal] = isSoftwareCertOverdue
	certTask.alarmIdFuncMap[ImageCertAbnormal] = isImageCertOverdue

	for i := 0; i < getConfigTimes; i++ {
		alarmConfigDir := filepath.Dir(dbPath)
		alarmDbMgr := common.NewDbMgr(alarmConfigDir, common.AlarmConfigDBName)
		period, err := alarmDbMgr.GetAlarmConfig(common.CertCheckPeriodDB)
		if err != nil {
			hwlog.RunLog.Errorf("get alarm config cert check period failed, error: %v", err)
			time.Sleep(getConfigWaitTime)
			continue
		}
		threshold, err := alarmDbMgr.GetAlarmConfig(common.CertOverdueThresholdDB)
		if err != nil {
			hwlog.RunLog.Errorf("get alarm config cert overdue threshold failed, error: %v", err)
			time.Sleep(getConfigWaitTime)
			continue
		}

		certOverdueThreshold = threshold
		certTask.interval = time.Duration(period) * common.OneDay
		return nil
	}

	hwlog.RunLog.Error("register cert monitor failed, reached the maximum number of the attempts")
	return errors.New("register cert monitor failed, reached the maximum number of the attempts")
}

func certReset() {
	getCertsInfoFlag = false
	importedCertsInfo = requests.ImportedCertsInfo{}
}

func isNorthCertOverdue() error {
	if !getCertsInfoFlag {
		if err := updateImportedCertsInfo(); err != nil {
			hwlog.RunLog.Errorf("get imported certs info failed, error: %v", err)
			return errors.New("get imported certs info failed")
		}
	}
	if importedCertsInfo.NorthCert == nil {
		hwlog.RunLog.Info("north cert is nil")
		return nil
	}
	return x509.CheckCertsOverdue(importedCertsInfo.NorthCert, certOverdueThreshold)
}

func isSoftwareCertOverdue() error {
	if !getCertsInfoFlag {
		if err := updateImportedCertsInfo(); err != nil {
			hwlog.RunLog.Errorf("get imported certs info failed, error: %v", err)
			return errors.New("get imported certs info failed")
		}
	}
	if importedCertsInfo.SoftwareCert == nil {
		hwlog.RunLog.Info("software repository cert is nil")
		return nil
	}
	return x509.CheckCertsOverdue(importedCertsInfo.SoftwareCert, certOverdueThreshold)
}

func isImageCertOverdue() error {
	if !getCertsInfoFlag {
		if err := updateImportedCertsInfo(); err != nil {
			hwlog.RunLog.Errorf("get imported certs info failed, error: %v", err)
			return errors.New("get imported certs info failed")
		}
	}
	if importedCertsInfo.ImageCert == nil {
		hwlog.RunLog.Info("image repository cert is nil")
		return nil
	}
	return x509.CheckCertsOverdue(importedCertsInfo.ImageCert, certOverdueThreshold)
}
