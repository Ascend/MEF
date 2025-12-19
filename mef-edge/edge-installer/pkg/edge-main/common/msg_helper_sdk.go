// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package common msg helper for sdk
package common

import (
	"errors"
	"strings"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/configpara"
)

var (
	centerIp       string
	originalMsgMap sync.Map
)

// MEFOpLog record mef operation log
func MEFOpLog(msg *model.Message) {
	if ignoreMEFOpLog(msg) {
		return
	}

	centerIp = configpara.GetNetConfig().IP
	operation := msg.KubeEdgeRouter.Operation
	resource := msg.KubeEdgeRouter.Resource
	id := msg.Header.ID

	recordMEFOpLog(centerIp, operation, resource, constants.Start, id)
	originalMsgMap.Store(id, operation+":"+resource)
	go deleteUnusedMsg(id)
	return
}

func ignoreMEFOpLog(msg *model.Message) bool {
	return msg.KubeEdgeRouter.Operation == constants.OptQuery || msg.KubeEdgeRouter.Operation == constants.OptResponse
}

func deleteUnusedMsg(id string) {
	const deleteInterval = 30 * time.Second
	tick := time.NewTicker(deleteInterval)
	defer tick.Stop()
	select {
	case <-tick.C:
		originalMsgMap.Delete(id)
	}
}

// MEFOpLogWithRes record mef operation log  with result
func MEFOpLogWithRes(resp *model.Message) {
	if ignoreMEFOpLogWithRes(resp) {
		return
	}

	originalOpt, originalRes, err := getMEFOriginalOpAndRes(resp)
	if err != nil {
		return
	}
	if originalOpt == constants.OptQuery || originalOpt == constants.OptResponse {
		return
	}
	originalId := resp.Header.ParentID

	defer originalMsgMap.Delete(resp.Header.ParentID)
	var content string
	if err = resp.ParseContent(&content); err != nil {
		recordMEFOpLog(centerIp, originalOpt, originalRes, constants.Failed, originalId)
		return
	}
	if content != constants.OK {
		recordMEFOpLog(centerIp, originalOpt, originalRes, constants.Failed, originalId)
		return
	}
	recordMEFOpLog(centerIp, originalOpt, originalRes, constants.Success, originalId)
}

func ignoreMEFOpLogWithRes(resp *model.Message) bool {
	return resp.Header.ParentID == "" || resp.KubeEdgeRouter.Operation != constants.OptResponse
}

func getMEFOriginalOpAndRes(msg *model.Message) (string, string, error) {
	// find the message matching the parent id
	opResFromMap, exist := originalMsgMap.Load(msg.Header.ParentID)
	if !exist {
		return "", "", errors.New("load operation and res from map failed")
	}
	opRes, ok := opResFromMap.(string)
	if !ok {
		return "", "", errors.New("the type of operation and resource from map is invalid")
	}
	opAndRes := strings.Split(opRes, ":")
	const opAndResLen = 2
	if len(opAndRes) != opAndResLen {
		return "", "", errors.New("split operation and resource failed")
	}
	return opAndRes[0], opAndRes[1], nil
}

func recordMEFOpLog(ip, operation, resource, result, msgId string) {
	switch result {
	case constants.Start, constants.Success:
		hwlog.OpLog.Infof("[%s@%s] %s %s %s, [msgId:%s]", constants.MEF, ip, operation, resource, result, msgId)
	case constants.Failed:
		hwlog.OpLog.Errorf("[%s@%s] %s %s %s, [msgId:%s]", constants.MEF, ip, operation, resource, result, msgId)
	default:
		hwlog.RunLog.Error("error operation log result")
	}
}
