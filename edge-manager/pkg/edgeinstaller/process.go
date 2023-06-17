// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller processing used in edge-installer module
package edgeinstaller

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/database"
	"edge-manager/pkg/nodemanager"
	"huawei.com/mindxedge/base/common"
)

func getNodeNum(nodeNums []int64) ([]string, error) {
	var nodeIds []string
	for _, nodeNum := range nodeNums {
		var node nodemanager.NodeInfo
		if err := database.GetDb().Model(nodemanager.NodeInfo{}).Where("id = ?", nodeNum).First(&node).Error; err != nil {
			hwlog.RunLog.Errorf("get nodeInfo failed, error: %v", err)
			return []string{}, err
		}
		nodeIds = append(nodeIds, node.UniqueName)
	}

	return nodeIds, nil
}

func respRestful(message *model.Message) (*UpgradeSfwReq, error) {
	var respContent = common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	upgradeSfwReq, err := constructContentToRestful(message)
	if err != nil {
		respContent = common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
		return nil, errors.New("edge-installer construct content to restful module failed")
	}
	respToRestful, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("edge-installer new response failed, error: %v", err)
		return nil, errors.New("edge-installer new response failed")
	}
	respToRestful.FillContent(respContent)
	if err = modulemgr.SendMessage(respToRestful); err != nil {
		hwlog.RunLog.Errorf("%s send response to restful failed", common.EdgeInstallerName)
		return nil, err
	}

	return upgradeSfwReq, nil
}

func constructContentToRestful(message *model.Message) (*UpgradeSfwReq, error) {
	var upgradeSfwReq UpgradeSfwReq
	if err := common.ParamConvert(message.GetContent(), &upgradeSfwReq); err != nil {
		return nil, err
	}

	if err := upgradeSfwReq.checkUpgradeSfwReq(); err != nil {
		return nil, err
	}

	return &upgradeSfwReq, nil
}
