// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// K8sLabelMgr is a struct that used to manager mef-center-node label on current working node
type K8sLabelMgr struct {
}

func (klm *K8sLabelMgr) getCurrentNodeName() (string, error) {
	localIps, err := GetPublicIps()
	if err != nil {
		hwlog.RunLog.Errorf("get local IP failed: %s", err.Error())
		return "", err
	}

	for _, localIp := range localIps {
		ipReg := fmt.Sprintf("\\s*%s\\s*", localIp)
		cmd := fmt.Sprintf(GetNodeCmdPattern, ipReg)
		nodeName, err := common.RunCommand("sh", false, common.DefCmdTimeoutSec, "-c", cmd)
		if err != nil {
			hwlog.RunLog.Errorf("get current node failed: %s", err.Error())
			return "", errors.New("get current node failed")
		}
		if nodeName == "" {
			continue
		}

		return nodeName, nil
	}

	hwlog.RunLog.Error("no valid node matches the device ip found")
	return "", errors.New("no valid node matches the device ip found")
}

// PrepareK8sLabel is used to prepare a mef-center-node label on current working node
// it will create it if it doesn't exist
func (klm *K8sLabelMgr) PrepareK8sLabel() error {
	nodeName, err := klm.getCurrentNodeName()
	if err != nil {
		return err
	}

	if strings.ContainsAny(nodeName, common.IllegalChars) {
		hwlog.RunLog.Error("the nodeName contains illegal characters")
		return errors.New("the nodeName contains illegal characters")
	}

	cmd := fmt.Sprintf(SetLabelCmdPattern, nodeName, K8sLabel)
	_, err = common.RunCommand("sh", false, common.DefCmdTimeoutSec, "-c", cmd)
	if err != nil {
		hwlog.RunLog.Errorf("set mef label failed: %s", err.Error())
		return err
	}
	return nil
}

// CheckK8sLabel is used to check if mef-center-node label exists on current working node
func (klm *K8sLabelMgr) CheckK8sLabel() (bool, error) {
	nodeName, err := klm.getCurrentNodeName()
	if err != nil {
		return false, err
	}

	if strings.ContainsAny(nodeName, common.IllegalChars) {
		hwlog.RunLog.Error("the nodeName contains illegal characters")
		return false, errors.New("the nodeName contains illegal characters")
	}

	nodeNameReg := fmt.Sprintf("'^%s\\s'", nodeName)
	cmd := fmt.Sprintf(CheckLabelCmdPattern, K8sLabel, nodeNameReg)
	ret, err := common.RunCommand("sh", false, common.DefCmdTimeoutSec, "-c", cmd)
	if err != nil {
		hwlog.RunLog.Errorf("check k8s label existence failed: %s", err.Error())
		return false, err
	}

	if ret != strconv.Itoa(LabelCount) {
		return false, nil
	}

	return true, nil
}
