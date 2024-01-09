// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module handler
package alarmmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	trueStr  = "true"
	falseStr = "false"

	centerNodeQueryType    = "CenterNodeQuery"
	serialNumQuery         = "SerialNumQuery"
	groupIdQueryType       = "GroupIdQuery"
	fullNodesQueryType     = "FullNodesQuery"
	fullEdgeNodesQueryType = "FullEdgeNodesQuery"
)

func listAlarms(msg *model.Message) (interface{}, error) {
	return dealRequest(msg, alarms.AlarmType), nil
}

func listEvents(msg *model.Message) (interface{}, error) {
	return dealRequest(msg, alarms.EventType), nil
}

func getAlarmDetail(msg *model.Message) (interface{}, error) {
	return getAlarmOrEventDbDetail(msg, alarms.AlarmType), nil
}

func getEventDetail(msg *model.Message) (interface{}, error) {
	return getAlarmOrEventDbDetail(msg, alarms.EventType), nil
}

// getQueryType returns the query type is by SerialNum or groupId,neither will return err
func getQueryType(req utils.ListAlarmOrEventReq) (string, utils.ListAlarmOrEventReq, error) {
	if req.IfCenter == trueStr {
		return centerNodeQueryType, req, nil
	}
	// without any param will return all
	if req.IfCenter == "" && req.Sn == "" && req.GroupId == 0 {
		return fullNodesQueryType, req, nil
	}

	if req.IfCenter == falseStr && req.Sn == "" && req.GroupId == 0 {
		return fullEdgeNodesQueryType, req, nil
	}
	if req.IfCenter == "" {
		req.IfCenter = falseStr
	}
	if req.Sn != "" {
		return serialNumQuery, req, nil
	}
	// groupId is not empty
	return groupIdQueryType, req, nil
}

func dealRequest(msg *model.Message, AlarmOrEvent string) *common.RespMsg {
	hwlog.RunLog.Infof("start listing all %s", AlarmOrEvent)
	var req utils.ListAlarmOrEventReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Error("failed to convert list center alarms/events inputs")
		return &common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	queryIdType, standardParam, err := getQueryType(req)
	if err != nil {
		hwlog.RunLog.Error("failed to convert parameters")
		return &common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}
	}
	if checkRes := NewAlarmListerChecker().Check(standardParam); !checkRes.Result {
		hwlog.RunLog.Errorf("list %s para checking failed,err:%s", queryIdType, checkRes.Reason)
		return &common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkRes.Reason}
	}

	if queryIdType == centerNodeQueryType {
		return listCenterAlarmOrEvents(req, AlarmOrEvent)
	}
	if queryIdType == serialNumQuery {
		return listEdgeAlarmsOrEventsBySn(req, AlarmOrEvent)
	}
	if queryIdType == groupIdQueryType {
		return listEdgeAlarmsOrEventsByGroupId(req, AlarmOrEvent)
	}
	if queryIdType == fullEdgeNodesQueryType {
		return listEdgeAlarmsOrEvents(req, AlarmOrEvent)
	}
	// fullNodesQueryType
	return listFullAlarmOrEvents(req, AlarmOrEvent)
}

func getListResp(alarms []AlarmInfo, total int64) utils.ListAlarmsResp {
	resp := utils.ListAlarmsResp{
		Total:   total,
		Records: make([]utils.AlarmBriefInfo, 0),
	}
	for _, alarm := range alarms {
		resp.Records = append(resp.Records, convertToDigestInfo(alarm))
	}
	return resp
}

func listEdgeAlarmsOrEvents(req utils.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	count, err := AlarmDbInstance().countEdgeAlarmsOrEvents(AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to count %s", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	alarmSlice, err := AlarmDbInstance().listAllEdgeAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get %s in db: %s", AlarmOrEvent, err.Error())
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	respMsg := getListResp(alarmSlice, count)
	hwlog.RunLog.Infof("succeed listing nodes %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func listFullAlarmOrEvents(req utils.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	count, err := AlarmDbInstance().countAlarmsOrEventsFullNodes(AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to count %s", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	alarmSlice, err := AlarmDbInstance().listAllAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get %s in db", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	respMsg := getListResp(alarmSlice, count)
	hwlog.RunLog.Infof("succeed listing nodes %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func listCenterAlarmOrEvents(req utils.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	count, err := AlarmDbInstance().countAlarmsOrEventsBySn(alarms.CenterSn, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to count %s", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListCenterNodeAlarm}
	}
	alarmSlice, err := AlarmDbInstance().listCenterAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get center nodes %s in db", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListCenterNodeAlarm}
	}
	respMsg := getListResp(alarmSlice, count)
	hwlog.RunLog.Infof("succeed listing center node %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func listEdgeAlarmsOrEventsBySn(req utils.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	count, err := AlarmDbInstance().countAlarmsOrEventsBySn(req.Sn, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to count %s", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm, Msg: fmt.Sprintf("failed to count %s", AlarmOrEvent)}
	}
	alarmSlice, err := AlarmDbInstance().listEdgeAlarmsOrEventsDb(req.PageNum, req.PageSize, req.Sn, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to list edge node[%s] %s in db,err:%s", req.Sn, AlarmOrEvent, err.Error())
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm}
	}
	respMsg := getListResp(alarmSlice, count)
	hwlog.RunLog.Infof("succeed listing edge node %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func listEdgeAlarmsOrEventsByGroupId(req utils.ListAlarmOrEventReq, queryType string) *common.RespMsg {
	router := common.Router{
		Source:      common.AlarmManagerClientName,
		Destination: common.AlarmManagerClientName,
		Option:      common.Get,
		Resource:    common.GetSnsByGroup,
	}

	getSnsReq := requests.GetSnsReq{GroupId: req.GroupId}
	bytes, err := json.Marshal(getSnsReq)
	if err != nil {
		hwlog.RunLog.Errorf("marshal req for getting sns by group id failed, error: %v", err)
		return &common.RespMsg{Status: common.ErrorParamInvalid}
	}
	resp := common.SendSyncMessageByRestful(string(bytes), &router, time.Second)
	nodeSns, err := parseEdgeManagerResp(resp, req.GroupId)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal resp for getting sns by group id failed, error: %v", err)
		return &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: err.Error()}
	}
	respMsg, err := getGroupAlarmsOrEvents(nodeSns, queryType, req)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm}
	}
	hwlog.RunLog.Infof("succeed listing group %s info", queryType)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func getGroupAlarmsOrEvents(nodeSns []string, queryType string, req utils.ListAlarmOrEventReq) (
	utils.ListAlarmsResp, error) {
	resp := utils.ListAlarmsResp{
		Records: make([]utils.AlarmBriefInfo, 0),
		Total:   0,
	}
	count, err := AlarmDbInstance().countAlarmsOrEventsBySns(nodeSns, queryType)
	if err != nil {
		return resp, fmt.Errorf("failed to count %s", queryType)
	}
	resp.Total = count
	if count == 0 {
		return resp, nil
	}

	alarmsNode, err := AlarmDbInstance().listAlarmsOrEventsOfGroup(req.PageNum, req.PageSize, nodeSns, queryType)
	if err != nil {
		return resp, errors.New("failed to list alarms of in db while list group alarms")
	}
	for _, alarmOfNode := range alarmsNode {
		resp.Records = append(resp.Records, convertToDigestInfo(alarmOfNode))
	}
	return resp, nil
}

func parseEdgeManagerResp(resp common.RespMsg, groupId uint64) ([]string, error) {
	status := resp.Status
	if status == common.ErrorNodeGroupNotFound {
		// unify with list alarms by sn,None existing sn or groupId will return Success code empty data
		hwlog.RunLog.Warnf("node group with id[%d] not found", groupId)
		return []string{}, nil
	}
	if status != common.Success {
		errMsg, ok := common.ErrorMap[resp.Status]
		if !ok {
			errMsg = "get sns from group id failed"
		}
		return []string{}, errors.New(errMsg)
	}
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return []string{}, errors.New("marshal sns information failed")
	}
	var nodeSns []string
	if err = json.Unmarshal(dataBytes, &nodeSns); err != nil {
		return []string{}, errors.New("unmarshal sns information failed")
	}
	return nodeSns, nil
}

func convertToDigestInfo(alarm AlarmInfo) utils.AlarmBriefInfo {
	return utils.AlarmBriefInfo{
		ID:        alarm.Id,
		Sn:        alarm.SerialNumber,
		Ip:        alarm.Ip,
		Severity:  alarm.PerceivedSeverity,
		Resource:  alarm.Resource,
		CreatedAt: alarm.CreatedAt,
		AlarmType: alarm.AlarmType,
	}
}

func getAlarmOrEventDbDetail(msg *model.Message, queryType string) *common.RespMsg {
	hwlog.RunLog.Infof("start to get %s information", queryType)
	var inputId uint64
	err := msg.ParseContent(&inputId)
	if err != nil {
		hwlog.RunLog.Errorf("parse content into uint64 failed: %v", err)
		return &common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	if chkRes := NewGetAlarmChecker().Check(inputId); !chkRes.Result {
		hwlog.RunLog.Errorf("check input id [%d] failed, error: %s", inputId, chkRes.Reason)
		return &common.RespMsg{Status: common.ErrorParamInvalid, Msg: chkRes.Reason}
	}
	alarmInfo, err := AlarmDbInstance().getAlarmOrEventInfoByAlarmInfoId(inputId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Errorf("id [%d] not found", inputId)
		return &common.RespMsg{Status: common.ErrorGetAlarmDetail,
			Msg: fmt.Sprintf("id [%d] not found", inputId)}
	}
	if err != nil {
		hwlog.RunLog.Errorf("failed to get alarm[id:%d], error: %s", inputId, err.Error())
		return &common.RespMsg{Status: common.ErrorGetAlarmDetail}
	}
	// judge the type of alarm is alarm instead of event
	if alarmInfo.AlarmType != queryType {
		hwlog.RunLog.Errorf("the inputID[%d] is not an ID of %s", inputId, queryType)
		return &common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("the inputID[%d] is not an ID of %s", inputId, queryType)}
	}

	hwlog.RunLog.Infof("succeeded to get %s detail [id:%d]", queryType, inputId)
	return &common.RespMsg{Status: common.Success, Data: alarmInfo}
}
