// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package monitors defined cert monitor, include northbound cert, software repository cert, and image repository cert
package monitors

import (
	"errors"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
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
	certTask.alarmIdFuncMap[alarms.NorthCertAbnormal] = isNorthCertOverdue
	certTask.alarmIdFuncMap[alarms.SoftwareCertAbnormal] = isSoftwareCertOverdue
	certTask.alarmIdFuncMap[alarms.ImageCertAbnormal] = isImageCertOverdue

	for i := 0; i < getConfigTimes; i++ {
		alarmConfigDir := filepath.Dir(dbPath)
		alarmDbMgr := common.NewDbMgr(alarmConfigDir, common.AlarmConfigDBName)
		period, err := alarmDbMgr.GetAlarmConfig(common.CertCheckPeriodDB)
		if err != nil || period == 0 {
			hwlog.RunLog.Error("get alarm config cert check period failed")
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
	hwlog.RunLog.Info("north cert overdue time check start...")
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
	if err := x509.CheckCertsOverdue(importedCertsInfo.NorthCert, certOverdueThreshold); err != nil {
		hwlog.RunLog.Errorf("check north cert overdue failed, error: %v", err)
		return errors.New("check north cert overdue failed")
	}

	hwlog.RunLog.Info("north cert overdue time check pass")
	return nil
}

func isSoftwareCertOverdue() error {
	hwlog.RunLog.Info("software repository cert overdue time check start...")
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
	if err := x509.CheckCertsOverdue(importedCertsInfo.SoftwareCert, certOverdueThreshold); err != nil {
		hwlog.RunLog.Errorf("check software repository cert overdue failed, error: %v", err)
		return errors.New("check software repository cert overdue failed")
	}

	hwlog.RunLog.Info("software repository cert overdue time check pass")
	return nil
}

func isImageCertOverdue() error {
	hwlog.RunLog.Info("image repository cert overdue time check start...")
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
	if err := x509.CheckCertsOverdue(importedCertsInfo.ImageCert, certOverdueThreshold); err != nil {
		hwlog.RunLog.Errorf("check image repository cert overdue failed, error: %v", err)
		return errors.New("check image repository cert overdue failed")
	}

	hwlog.RunLog.Info("image repository cert overdue time check pass")
	return nil
}
