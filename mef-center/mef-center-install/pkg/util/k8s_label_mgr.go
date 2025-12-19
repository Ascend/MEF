// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"errors"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
)

const regexpNodeName = "^[a-z0-9]([a-z0-9-.]{0,251}[a-z0-9])?$"

// K8sLabelMgr is a struct that used to manager mef-center-node label on current working node
type K8sLabelMgr struct {
}

func (klm *K8sLabelMgr) getMasterNodeName() (string, error) {
	ret, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "nodes", "-o",
		"custom-columns=:.metadata.name", "--selector=node-role.kubernetes.io/master=", "--no-headers")
	if err != nil {
		hwlog.RunLog.Errorf("get current node failed: %s", err.Error())
		return "", errors.New("get current node failed")
	}

	lines := strings.Split(ret, "\n")
	if len(lines) != 1 {
		hwlog.RunLog.Errorf("there is more than one master node")
		return "", errors.New("there is more than one master node")
	}
	res := checker.GetRegChecker("", regexpNodeName, true).Check(lines[0])
	if !res.Result {
		hwlog.RunLog.Error("the nodeName contains illegal characters")
		return "", errors.New("the nodeName contains illegal characters")
	}

	hwlog.RunLog.Info("get node name success")
	return lines[0], nil
}

// PrepareK8sLabel is used to prepare a mef-center-node label on current working node
// it will create it if it doesn't exist
func (klm *K8sLabelMgr) PrepareK8sLabel() error {
	nodeName, err := klm.getMasterNodeName()
	if err != nil {
		return err
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
	nodeName, err := klm.getMasterNodeName()
	if err != nil {
		return false, err
	}

	ret, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "nodes", "-o",
		"custom-columns=:.status.addresses[*].address", "-l", K8sLabel, "--no-headers")
	if err != nil {
		hwlog.RunLog.Errorf("check k8s label existence failed: %s", err.Error())
		return false, err
	}

	lines := strings.Split(ret, "\n")
	// only one currently
	const maxK8sLabelNodes = 20
	if len(lines) > maxK8sLabelNodes {
		hwlog.RunLog.Error("the number of k8s label nodes exceed the upper limit")
		return false, errors.New("the number of k8s label nodes exceed the upper limit")
	}
	for _, line := range lines {
		parts := strings.Split(line, ",")
		for _, part := range parts {
			if part == nodeName {
				return true, nil
			}
		}
	}

	return false, nil
}
