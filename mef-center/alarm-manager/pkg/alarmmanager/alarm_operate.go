// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	maxOneNodeAlarmCount = 20
	maxOneNodeEventCount = 50
)

func dealAlarmsReq(msg *model.Message) interface{} {
	var req requests.AlarmsReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("alarms req param parse failed: %s", err.Error())
		return nil
	}

	if checkResult := NewDealAlarmsReqChecker().Check(req); !checkResult.Result {
		hwlog.RunLog.Errorf("alarms req param check failed: %s", checkResult.Reason)
		return nil
	}

	for _, alarmReq := range req.Alarms {
		dealer := GetAlarmReqDealer(&alarmReq, req.Sn, req.Ip)
		if err := dealer.deal(); err != nil {
			hwlog.RunLog.Errorf("deal alarms req failed: %s", err.Error())
			return nil
		}
	}

	return nil
}

// AlarmReqDealer is the struct to deal with one alarm request
type AlarmReqDealer struct {
	req       *requests.AlarmReq
	sn        string
	ip        string
	alarmInfo *AlarmInfo
}

// GetAlarmReqDealer is the func to create an AlarmReqDealer
func GetAlarmReqDealer(req *requests.AlarmReq, sn string, ip string) *AlarmReqDealer {
	return &AlarmReqDealer{
		req: req,
		sn:  sn,
		ip:  ip,
	}
}

func (ard *AlarmReqDealer) deal() error {
	if ard.req == nil {
		return errors.New("req is nil")
	}

	alarmInfo, err := ard.getAlarmInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info failed: %s", err.Error())
		return errors.New("get alarm info failed")
	}
	ard.alarmInfo = alarmInfo

	if ard.req.Type == alarms.AlarmType {
		return ard.dealAlarm()
	} else {
		return ard.dealEvent()
	}
}

func (ard *AlarmReqDealer) getAlarmInfo() (*AlarmInfo, error) {
	parsedTime, err := time.Parse(time.RFC3339, ard.req.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("parse time failed: %s", err.Error())
	}
	return &AlarmInfo{
		AlarmType:           ard.req.Type,
		CreatedAt:           parsedTime,
		SerialNumber:        ard.sn,
		Ip:                  ard.ip,
		AlarmId:             ard.req.AlarmId,
		AlarmName:           ard.req.AlarmName,
		PerceivedSeverity:   ard.req.PerceivedSeverity,
		DetailedInformation: ard.req.DetailedInformation,
		Suggestion:          ard.req.Suggestion,
		Reason:              ard.req.Reason,
		Impact:              ard.req.Impact,
		Resource:            ard.req.Resource,
	}, nil
}

func (ard *AlarmReqDealer) dealAlarm() error {
	if ard.req.NotificationType == alarms.ClearFlag {
		return ard.clearAlarm()
	} else {
		return ard.addAlarm()
	}
}

func (ard *AlarmReqDealer) clearAlarm() error {
	ret, err := AlarmDbInstance().getAlarmInfo(ard.req.AlarmId, ard.sn)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info from db failed: %s", err.Error())
		return errors.New("get alarm info from db failed")
	}

	// do not record log when alarm does not exist
	if len(ret) == 0 {
		return nil
	}

	if ard.alarmInfo == nil {
		return errors.New("alarm info is nil")
	}
	if err = AlarmDbInstance().deleteAlarm(ard.alarmInfo); err != nil {
		hwlog.RunLog.Errorf("%v [%s:%s] %v %v: clear alarm %v from db failed: %s",
			time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn,
			http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
		return errors.New("delete alarm data failed")
	}

	hwlog.RunLog.Infof("%v [%s:%s] %v %v: clear alarm %v from db success", time.Now().Format(time.RFC3339Nano),
		ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	return nil
}

func (ard *AlarmReqDealer) addAlarm() error {
	count, err := AlarmDbInstance().getNodeAlarmCount(ard.sn)
	if err != nil {
		hwlog.RunLog.Errorf("get node alarm count failed: %s", err.Error())
		return errors.New("get node alarm count failed")
	}

	if count >= maxOneNodeAlarmCount {
		hwlog.RunLog.Errorf("node %s's alarm has reached the max counts", ard.sn)
		return errors.New("node's alarm count have reached the max counts")
	}

	ret, err := AlarmDbInstance().getAlarmInfo(ard.req.AlarmId, ard.sn)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info from db failed: %s", err.Error())
		return errors.New("get alarm info from db failed")
	}

	// the alarm already exists, ignore
	if len(ret) != 0 {
		return nil
	}

	if ard.alarmInfo == nil {
		return errors.New("alarm info is nil")
	}
	if err = AlarmDbInstance().addAlarmInfo(ard.alarmInfo); err != nil {
		hwlog.RunLog.Errorf("%v [%s:%s] %v %v: add alarm %v into db failed: %s", time.Now().
			Format(time.RFC3339Nano), ard.ip, ard.sn,
			http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
		return errors.New("add alarm into db failed")
	}

	hwlog.RunLog.Infof("%v [%s:%s] %v %v: add alarm %v into db success",
		time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	return nil
}

func (ard *AlarmReqDealer) dealEvent() error {
	count, err := AlarmDbInstance().getNodeEventCount(ard.sn)
	if err != nil {
		hwlog.RunLog.Errorf("get node event count failed: %s", err.Error())
		return errors.New("get node event count failed")
	}

	if count >= maxOneNodeEventCount {
		oldestEvent, err := AlarmDbInstance().getNodeOldEvent(ard.sn, maxOneNodeEventCount-1)
		if err != nil {
			hwlog.RunLog.Errorf("get node oldest event failed: %s", err.Error())
			return errors.New("get node oldest event failed")
		}

		if err = AlarmDbInstance().deleteAlarms(oldestEvent); err != nil {
			hwlog.RunLog.Errorf("%v [%s:%s] %v %v: delete oldest event[%s] failed: %s",
				time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn, http.MethodPost,
				requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
			return errors.New("delete oldest event failed")
		}
		hwlog.RunLog.Infof("%v [%s:%s] %v %v: delete oldest event[%s] success", time.Now().Format(time.RFC3339Nano),
			ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	}

	if ard.alarmInfo == nil {
		return errors.New("alarm info is nil")
	}
	if err = AlarmDbInstance().addAlarmInfo(ard.alarmInfo); err != nil {
		hwlog.RunLog.Errorf("%v [%s:%s] %v %v: add event %v into db failed: %s",
			time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn,
			http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
		return errors.New("add new event into db failed")
	}

	hwlog.RunLog.Infof("%v [%s:%s] %v %v: add event %v into db success",
		time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	return nil
}

func dealNodeClearReq(msg *model.Message) interface{} {
	var reqs requests.ClearNodeAlarmReq
	if err := msg.ParseContent(&reqs); err != nil {
		hwlog.RunLog.Errorf("clear node alarm req param parse failed: %v", err)
		return common.FAIL
	}

	if err := AlarmDbInstance().deleteBySn(reqs.Sn); err != nil {
		hwlog.RunLog.Errorf("delete alarm info by sn [%s] failed: %s", reqs.Sn, err.Error())
		return common.FAIL
	}

	hwlog.RunLog.Infof("clear all alarms of node %s success", reqs.Sn)
	return common.OK
}
