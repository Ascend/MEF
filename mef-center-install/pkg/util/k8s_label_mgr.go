// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// K8sLabelMgr is a struct that used to manager mef-center-node label on current working node
type K8sLabelMgr struct {
}

func (klm *K8sLabelMgr) getCurrentNodeName() (string, error) {
	localIps, err := GetLocalIps()
	if err != nil {
		hwlog.RunLog.Errorf("get local IP failed: %s", err.Error())
		return "", err
	}

	ret, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "nodes", "-o", "wide")
	if err != nil {
		hwlog.RunLog.Errorf("get current node failed: %s", err.Error())
		return "", errors.New("get current node failed")
	}
	lines := strings.Split(ret, "\n")

	for _, localIp := range localIps {
		ipReg := fmt.Sprintf("\\s*%s\\s*", localIp)
		for _, line := range lines {
			found, err := regexp.MatchString(ipReg, line)
			if err != nil {
				hwlog.RunLog.Errorf("get current node name on reg match failed: %s", err.Error())
				return "", errors.New("get current node name failed")
			}
			if !found {
				continue
			}
			r := regexp.MustCompile("\\S+")
			res := r.FindAllString(line, SplitStringCount)
			if len(res) < NodeSplitCount {
				hwlog.RunLog.Errorf("get current node name failed: find invalid data")
				return "", errors.New("get current node name failed")
			}
			return res[0], nil
		}
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

	_, err = envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "label", "node", nodeName,
		"--overwrite", K8sLabelSet)
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

	ret, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "nodes", "-o", "wide", "-l",
		K8sLabel)
	if err != nil {
		hwlog.RunLog.Errorf("check k8s label existence failed: %s", err.Error())
		return false, err
	}

	nodeNameReg := fmt.Sprintf("^%s\\s", nodeName)
	lines := strings.Split(ret, "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(nodeNameReg, line)
		if err != nil {
			hwlog.RunLog.Errorf("check k8s label on reg match failed: %s", err.Error())
			return false, errors.New("check k8s label failed")
		}
		if found {
			return true, nil
		}
	}

	return false, nil
}
