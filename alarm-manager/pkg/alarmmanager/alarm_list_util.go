// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module to support query from north
package alarmmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"alarm-manager/pkg/types"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"

	"huawei.com/mindxedge/base/common"
)

const (
	trueStr  = "true"
	falseStr = "false"
)

// getQueryType returns the query type is by SerialNum or groupId,neither will return err
func getQueryType(input interface{}) (string, types.ListAlarmOrEventReq, error) {
	req, ok := input.(types.ListAlarmOrEventReq)
	if !ok {
		hwlog.RunLog.Error("failed to convert params")
		return "", types.ListAlarmOrEventReq{}, errors.New("failed to convert params")
	}
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

func dealRequest(input interface{}, AlarmOrEvent string) *common.RespMsg {
	hwlog.RunLog.Infof("start listing all %s", AlarmOrEvent)
	queryIdType, standardParam, err := getQueryType(input)
	if err != nil {
		hwlog.RunLog.Error("failed to convert parameters")
		return &common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}
	}
	if checkRes := NewAlarmListerChecker().Check(standardParam); !checkRes.Result {
		hwlog.RunLog.Errorf("list %s para checking failed,err:%s", queryIdType, checkRes.Reason)
		return &common.RespMsg{Status: common.ErrorParamInvalid, Msg: checkRes.Reason}
	}

	req, ok := input.(types.ListAlarmOrEventReq)
	if !ok {
		hwlog.RunLog.Error("failed to convert list center alarms/events inputs")
		return &common.RespMsg{Status: common.ErrorParamConvert}
	}

	if queryIdType == centerNodeQueryType {
		return listCenterAlarmOrEvents(req, AlarmOrEvent)
	}
	if queryIdType == serialNumQuery {
		return listEdgeNodeAlarmsOrEvents(req, AlarmOrEvent)
	}
	if queryIdType == groupIdQueryType {
		return listGroupNodesAlarmsOrEvents(req, AlarmOrEvent)
	}
	if queryIdType == fullEdgeNodesQueryType {
		return listAllEdgeNodesAlarmsOrEvents(req, AlarmOrEvent)
	}
	// fullNodesQueryType
	return listFullAlarmOrEvents(req, AlarmOrEvent)
}

func getListResp(alarms []AlarmInfo, total int64) types.ListAlarmsResp {
	resp := types.ListAlarmsResp{
		Total:   total,
		Records: make([]types.AlarmBriefInfo, 0),
	}
	for _, alarm := range alarms {
		resp.Records = append(resp.Records, convertToDigestInfo(alarm))
	}
	return resp
}

func listAllEdgeNodesAlarmsOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	count, err := AlarmDbInstance().countAlarmsOrEventsOfEdgeNodes(AlarmOrEvent)
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

func listFullAlarmOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
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

func listCenterAlarmOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
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

func listEdgeNodeAlarmsOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
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

func listGroupNodesAlarmsOrEvents(req types.ListAlarmOrEventReq, queryIdType string) *common.RespMsg {
	router := common.Router{
		Source:      common.AlarmManagerClientName,
		Destination: common.AlarmManagerClientName,
		Option:      common.Get,
		Resource:    common.GetSnsByGroup,
	}

	edgeReq := requests.NodeGroupReq{GroupId: req.GroupId}
	bytes, err := json.Marshal(edgeReq)
	if err != nil {
		hwlog.RunLog.Error("failed to marshal")
		return &common.RespMsg{Status: common.ErrorParamInvalid}
	}
	resp := common.SendSyncMessageByRestful(string(bytes), &router, time.Second)
	nodeSns, err := parseEdgeManagerResp(resp, req.GroupId)
	if err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal group node list from edge-manager,err:%s", err.Error())
		return &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: err.Error()}
	}
	respMsg, err := getGroupAlarmsOrEvents(nodeSns, queryIdType, req)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm}
	}
	hwlog.RunLog.Infof("succeed listing group %s info", queryIdType)
	return &common.RespMsg{Status: common.Success, Data: respMsg}
}

func getGroupAlarmsOrEvents(nodeSns []string, queryIdType string, req types.ListAlarmOrEventReq) (
	types.ListAlarmsResp, error) {
	resp := types.ListAlarmsResp{
		Records: make([]types.AlarmBriefInfo, 0),
		Total:   0,
	}
	count, err := AlarmDbInstance().countAlarmsOrEventsOfNodes(nodeSns, queryIdType)
	if err != nil {
		return resp, fmt.Errorf("failed to count %s", queryIdType)
	}
	resp.Total = count
	if count == 0 {
		hwlog.RunLog.Infof("succeed listing nodes %s info", queryIdType)
		return resp, nil
	}

	alarmsNode, err := AlarmDbInstance().listAlarmsOrEventsOfGroup(req.PageNum, req.PageSize, nodeSns, queryIdType)
	if err != nil {
		return resp, errors.New("faild to list alarms of in db while list group alarms")
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
		return []string{}, errors.New(common.ErrorMap[resp.Status])
	}
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return []string{}, errors.New("decode nodeGroup information failed")
	}
	var nodeSns []string
	if err := json.Unmarshal(dataBytes, &nodeSns); err != nil {
		return []string{}, errors.New("nodeGroup information convert failed")
	}
	return nodeSns, nil
}

func convertToDigestInfo(alarm AlarmInfo) types.AlarmBriefInfo {
	return types.AlarmBriefInfo{
		ID:        alarm.Id,
		Sn:        alarm.SerialNumber,
		Ip:        alarm.Ip,
		Severity:  alarm.PerceivedSeverity,
		Resource:  alarm.Resource,
		CreatedAt: alarm.CreatedAt,
		AlarmType: alarm.AlarmType,
	}
}

func getAlarmOrEventDbDetail(input interface{}, queryType string) *common.RespMsg {
	hwlog.RunLog.Infof("start to get %s information", queryType)
	inputId, ok := input.(uint64)
	if !ok {
		hwlog.RunLog.Errorf("failed to convert input alarmInfoId[%v] to uint64", input)
		return &common.RespMsg{Status: common.ErrorTypeAssert, Msg: "failed to convert input to int"}
	}
	if chkRes := NewGetAlarmChecker().Check(inputId); !chkRes.Result {
		hwlog.RunLog.Errorf("failed to check alarmInfoId[%d],err:%s", inputId, chkRes.Reason)
		return &common.RespMsg{Status: common.ErrorParamInvalid, Msg: chkRes.Reason}
	}
	alarmInfo, err := AlarmDbInstance().getAlarmOrEventInfoByAlarmInfoId(inputId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hwlog.RunLog.Errorf("alarmId[%d] not found", inputId)
		return &common.RespMsg{Status: common.ErrorGetAlarmDetail,
			Msg: fmt.Sprintf("alarmId[%d] not found", inputId)}
	}
	if err != nil {
		hwlog.RunLog.Errorf("failed to get alarm[alarmId:%d],err:%s", inputId, err.Error())
		return &common.RespMsg{Status: common.ErrorGetAlarmDetail}
	}
	// judge the type of alarm is alarm instead of event
	if alarmInfo.AlarmType != queryType {
		hwlog.RunLog.Errorf("the inputID[%d] is not an ID of %s", inputId, queryType)
		return &common.RespMsg{Status: common.ErrorParamInvalid,
			Msg: fmt.Sprintf("the inputID[%d] is not an ID of %s", inputId, queryType)}
	}

	hwlog.RunLog.Infof("succeeded to get %s detail[alarmInfoId:%d]", queryType, inputId)
	return &common.RespMsg{Status: common.Success, Data: alarmInfo}
}
