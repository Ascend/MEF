// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package monitors common func
package monitors

import (
	"encoding/json"
	"errors"
	"net/http"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

// GetAlarmMonitorList get mef monitor list, only cert alarm support currently.
func GetAlarmMonitorList(dbPath string) []AlarmMonitor {
	var alarmMonitor []AlarmMonitor
	registers := []struct {
		monitorName AlarmMonitor
		err         error
	}{
		{certTask, registerCertMonitor(dbPath)},
	}

	for _, register := range registers {
		if register.err != nil {
			hwlog.RunLog.Error(register.err)
			continue
		}
		alarmMonitor = append(alarmMonitor, register.monitorName)
	}

	return alarmMonitor
}

// SendAlarms send alarms
func SendAlarms(alarms ...*requests.AlarmReq) error {
	if len(alarms) == 0 {
		hwlog.RunLog.Error("alarm is required")
		return errors.New("alarm is required")
	}
	var alarmReqs []requests.AlarmReq
	for _, alm := range alarms {
		if alm == nil {
			hwlog.RunLog.Error("alarm req can not be nil pointer")
			return errors.New("alarm req can not be nil pointer")
		}
		alarmReqs = append(alarmReqs, *alm)
	}

	hostIp, err := common.GetHostIP("HOST_IP")
	if err != nil {
		hwlog.RunLog.Errorf("get host ip failed, error: %v", err)
		return errors.New("get host ip failed")
	}

	alarmsReq := requests.AlarmsReq{
		Alarms: alarmReqs,
		Sn:     "",
		Ip:     hostIp,
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new alarm msg failed, error: %v", err)
		return errors.New("new alarm msg failed")
	}
	msg.SetRouter(common.AlarmManagerName, common.AlarmManagerName, http.MethodPost, requests.ReportAlarmRouter)
	if err = msg.FillContent(alarmsReq, true); err != nil {
		hwlog.RunLog.Errorf("fill add alarm req failed: %v", err)
		return errors.New("fill content failed")
	}

	if err = modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send async message failed, error: %v", err)
		return errors.New("send async message failed")
	}

	return nil
}

func updateImportedCertsInfo() error {
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: utils.RootCaPath,
			CertPath:   utils.ServerCertPath,
			KeyPath:    utils.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	certsInfo, err := reqCertParams.GetImportedCertsInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get imported certs info from cert-manager failed, error: %v", err)
		return errors.New("get imported certs info from cert-manager failed")
	}

	resp := requests.ImportedCertsInfo{}
	if err = json.Unmarshal([]byte(certsInfo), &resp); err != nil {
		hwlog.RunLog.Errorf("unmarshal imported certs info failed, error: %v", err)
		return errors.New("unmarshal imported certs info failed")
	}
	getCertsInfoFlag = true
	importedCertsInfo = resp
	return nil
}
