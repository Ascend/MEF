// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module to support query from north
package alarmmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

const (
	totalRecordsKey = "total"
	respDataKey     = "records"
	groupQueryRoute = "edgemanager/v1/nodegroup"
	trueStr         = "true"
	falseStr        = "false"
)

// getQueryType returns the query type is by nodeId or groupId,neither will return err
func getQueryType(input interface{}) (string, types.ListAlarmOrEventReq, error) {
	req, ok := input.(types.ListAlarmOrEventReq)
	if !ok {
		hwlog.RunLog.Error("failed to convert params")
		return "", types.ListAlarmOrEventReq{}, fmt.Errorf("failed to convert params")
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
		return nodeIdQueryType, req, nil
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
	if queryIdType == nodeIdQueryType {
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

func getListResMap(alarms *[]AlarmInfo) map[string]interface{} {
	respMap := make(map[string]interface{})
	alarmsMap := make(map[uint64]types.AlarmBriefInfo)
	for _, alarm := range *alarms {
		alarmsMap[alarm.Id] = convertToDigestInfo(alarm)
	}
	respMap[totalRecordsKey] = len(*alarms)
	respMap[respDataKey] = alarmsMap
	return respMap
}

func listAllEdgeNodesAlarmsOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	alarms, err := AlarmDbInstance().listAllEdgeAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get %s in db: %s", AlarmOrEvent, err.Error())
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	if len(*alarms) == 0 {
		return &common.RespMsg{Status: common.Success}
	}
	respMap := getListResMap(alarms)
	hwlog.RunLog.Infof("succeed listing nodes %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMap}
}

func listFullAlarmOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	alarms, err := AlarmDbInstance().listAllAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get %s in db", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListAlarm}
	}
	if len(*alarms) == 0 {
		return &common.RespMsg{Status: common.Success}
	}
	respMap := getListResMap(alarms)
	hwlog.RunLog.Infof("succeed listing nodes %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMap}
}

func listCenterAlarmOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	alarms, err := AlarmDbInstance().listCenterAlarmsOrEventsDb(req.PageNum, req.PageSize, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get center nodes %s in db", AlarmOrEvent)
		return &common.RespMsg{Status: common.ErrorListCenterNodeAlarm}
	}
	if len(*alarms) == 0 {
		return &common.RespMsg{Status: common.Success}
	}
	respMap := getListResMap(alarms)
	hwlog.RunLog.Infof("succeed listing center node %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMap}
}

func listEdgeNodeAlarmsOrEvents(req types.ListAlarmOrEventReq, AlarmOrEvent string) *common.RespMsg {
	alarms, err := AlarmDbInstance().listEdgeAlarmsOrEventsDb(req.PageNum, req.PageSize, req.Sn, AlarmOrEvent)
	if err != nil {
		hwlog.RunLog.Errorf("failed to list edge node[%d] %s in db,err:%s", req.Sn, AlarmOrEvent, err.Error())
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm}
	}
	if len(*alarms) == 0 {
		return &common.RespMsg{Status: common.Success}
	}
	respMap := getListResMap(alarms)
	hwlog.RunLog.Infof("succeed listing edge node %s info", AlarmOrEvent)
	return &common.RespMsg{Status: common.Success, Data: respMap}
}

func listGroupNodesAlarmsOrEvents(req types.ListAlarmOrEventReq, queryIdType string) *common.RespMsg {
	// ask edgemanager for node list in a specific group
	url := fmt.Sprintf("https://%s:%d/%s?id=%d", common.EdgeMgrDns, common.EdgeMgrPort,
		groupQueryRoute, req.GroupId)
	clientSvcCert := certutils.TlsCertInfo{
		RootCaPath: RootCaPath,
		CertPath:   ServerCertPath,
		KeyPath:    ServerKeyPath,
		SvrFlag:    false,
	}
	httpsReq := httpsmgr.GetHttpsReq(url, clientSvcCert)
	const timeout = 3 * time.Second
	resp, err := httpsReq.GetWithTimeout(nil, timeout)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get group node list from edge-manager,err:%s", err.Error())
		return &common.RespMsg{Status: common.ErrorListGroupNodeFromEdgeMgr}
	}
	nodes, err := parseEdgeManagerResp(resp)
	if err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal group node list from edge-manager,err:%s", err.Error())
		return &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: err.Error()}
	}
	if len(nodes) == 0 {
		return &common.RespMsg{Status: common.Success}
	}
	respMap, err := getMapResult(nodes, queryIdType, req)
	if err != nil {
		hwlog.RunLog.Error(err.Error())
		return &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm}
	}
	hwlog.RunLog.Infof("succeed listing group %s info", EventType)
	return &common.RespMsg{Status: common.Success, Data: respMap}
}

func getMapResult(nodes []types.NodeInfo, queryIdType string, req types.ListAlarmOrEventReq) (
	map[string]interface{}, error) {
	respMap := make(map[string]interface{})
	alarmMap := make(map[uint64]types.AlarmBriefInfo)
	count := 0
	for _, node := range nodes {
		alarmsNode, err := AlarmDbInstance().listEdgeAlarmsOrEventsDb(req.PageNum, req.PageSize, node.Sn, queryIdType)
		if err != nil {
			return nil, fmt.Errorf("faild to list alarms of node[%s] in db while list group alarms", node.Sn)
		}
		for _, alarmOfNode := range *alarmsNode {
			count++
			alarmMap[alarmOfNode.Id] = convertToDigestInfo(alarmOfNode)
		}

	}
	respMap[totalRecordsKey] = count
	respMap[respDataKey] = alarmMap
	return respMap, nil
}

func parseEdgeManagerResp(respBytes []byte) ([]types.NodeInfo, error) {
	var resp common.RespMsg
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return []types.NodeInfo{}, err
	}
	status := resp.Status
	if status != common.Success {
		return []types.NodeInfo{}, fmt.Errorf(common.ErrorMap[resp.Status])
	}
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return []types.NodeInfo{}, fmt.Errorf("decode nodeGroup information failed")
	}
	var nodes types.NodeGroupDetailFromEdgeManager
	if err := json.Unmarshal(dataBytes, &nodes); err != nil {
		return []types.NodeInfo{}, fmt.Errorf("nodeGroup information convert failed")
	}
	return nodes.Nodes, nil
}

func convertToDigestInfo(alarm AlarmInfo) types.AlarmBriefInfo {
	return types.AlarmBriefInfo{
		ID:        alarm.Id,
		Sn:        alarm.SerialNumber,
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
	resMap := make(map[uint64]AlarmInfo)
	resMap[inputId] = *alarmInfo
	return &common.RespMsg{Status: common.Success, Data: resMap}
}
