// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"huawei.com/mindx/common/checker"
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

func dealAlarmReq(msg *model.Message) (interface{}, error) {
	var reqs requests.AddAlarmReq
	if err := msg.ParseContent(&reqs); err != nil {
		hwlog.RunLog.Errorf("param convert failed: %s", err.Error())
		return nil, errors.New("param convert failed")
	}

	snChecker := checker.GetOrChecker(
		checker.GetSnChecker("", true),
		checker.GetStringChoiceChecker("", []string{alarms.CenterSn}, true),
	)
	ret := snChecker.Check(reqs.Sn)
	if !ret.Result {
		hwlog.RunLog.Error("deal alarm para check failed: unsupported serial number received")
		return nil, errors.New("deal alarm para check failed")
	}

	ret = checker.GetIpV4Checker("", true).Check(reqs.Ip)
	if !ret.Result {
		hwlog.RunLog.Error("deal alarm para check failed: unsupported Ip received")
		return nil, errors.New("deal alarm para check failed")
	}

	if len(reqs.Alarms) > maxOneNodeAlarmCount {
		hwlog.RunLog.Error("alarms request exceeds the max count limitation")
		return nil, errors.New("alarm request exceeds the max count limitation")
	}

	for _, req := range reqs.Alarms {
		if checkResult := NewDealAlarmChecker().Check(req); !checkResult.Result {
			hwlog.RunLog.Errorf("deal alarm para check failed: %s", checkResult.Reason)
			return nil, errors.New("deal alarm para check failed")
		}

		dealer := GetAlarmReqDealer(&req, reqs.Sn, reqs.Ip)
		if err := dealer.deal(); err != nil {
			hwlog.RunLog.Errorf("deal alarm req failed: %s", err.Error())
			return nil, errors.New("deal alarm req failed")
		}
	}

	return nil, nil
}

// AlarmReqDealer is the struct to deal with one alarm request
type AlarmReqDealer struct {
	req *requests.AlarmReq
	sn  string
	ip  string
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
		return ard.dealAlarmClear()
	} else {
		return ard.dealAlarmAdd()
	}
}

func (ard *AlarmReqDealer) dealAlarmClear() error {
	alarmInfoData, err := ard.getAlarmInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info failed: %s", err.Error())
		return errors.New("get alarm info failed")
	}

	ret, err := AlarmDbInstance().getAlarmInfo(ard.req.AlarmId, ard.sn)
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info from db failed: %s", err.Error())
		return errors.New("get alarm info from db failed")
	}

	// do not record log when alarm does not exist
	if len(ret) == 0 {
		return nil
	}

	if err := AlarmDbInstance().deleteOneAlarm(alarmInfoData); err != nil {
		hwlog.RunLog.Errorf("%v [%s:%s] %v %v: clear alarm %v from db failed: %s",
			time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn,
			http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
		return errors.New("delete alarm data failed")
	}

	hwlog.RunLog.Infof("%v [%s:%s] %v %v: clear alarm %v from db success", time.Now().Format(time.RFC3339Nano),
		ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	return nil
}

func (ard *AlarmReqDealer) dealAlarmAdd() error {
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

	if len(ret) != 0 {
		return nil
	}

	alarmInfoData, err := ard.getAlarmInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info failed: %s", err.Error())
		return errors.New("get alarm info failed")
	}

	if err = AlarmDbInstance().addAlarmInfo(alarmInfoData); err != nil {
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

		if err = AlarmDbInstance().deleteAlarmInfos(oldestEvent); err != nil {
			hwlog.RunLog.Errorf("%v [%s:%s] %v %v: delete oldest event[%s] failed: %s",
				time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn, http.MethodPost,
				requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
			return errors.New("delete oldest event failed")
		}
		hwlog.RunLog.Infof("%v [%s:%s] %v %v: delete oldest event[%s] success", time.Now().Format(time.RFC3339Nano),
			ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	}

	eventData, err := ard.getAlarmInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get alarm info failed: %s", err.Error())
		return errors.New("get alarm info failed")
	}

	if err = AlarmDbInstance().addAlarmInfo(eventData); err != nil {
		hwlog.RunLog.Errorf("%v [%s:%s] %v %v: add event %v into db failed: %s",
			time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn,
			http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId, err.Error())
		return errors.New("add new event into db failed")
	}

	hwlog.RunLog.Infof("%v [%s:%s] %v %v: add event %v into db success",
		time.Now().Format(time.RFC3339Nano), ard.ip, ard.sn, http.MethodPost, requests.ReportAlarmRouter, ard.req.AlarmId)
	return nil
}

func dealNodeClearReq(msg *model.Message) (interface{}, error) {
	var reqs requests.ClearNodeAlarmReq
	if err := msg.ParseContent(&reqs); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return common.FAIL, nil
	}

	if err := AlarmDbInstance().deleteBySn(reqs.Sn); err != nil {
		hwlog.RunLog.Errorf("delete alarm info by sn [%s] failed: %s", reqs.Sn, err.Error())
		return common.FAIL, nil
	}

	hwlog.RunLog.Infof("clear all alarms from node %s success", reqs.Sn)
	return common.OK, nil
}
