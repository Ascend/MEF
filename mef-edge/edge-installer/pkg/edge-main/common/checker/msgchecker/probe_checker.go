// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package msgchecker

import (
	"errors"
	"net"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
)

const (
	minPort = 1
	maxPort = 65535
)

func checkProbePara(probe *types.Probe) error {
	if probe == nil {
		return nil
	}

	if !checkHttpProbePara(probe.ProbeHandler.HTTPGet) {
		return errors.New("container probe para check failed")
	}

	return nil
}

func checkHttpProbePara(httpGet *types.HTTPGetAction) bool {
	if httpGet == nil || len(httpGet.Path) == 0 {
		return true
	}

	if !checker.IsPathValid(httpGet.Path) {
		hwlog.RunLog.Error("check probe path invalid")
		return false
	}

	if httpGet.Host != "" && net.ParseIP(httpGet.Host) == nil {
		hwlog.RunLog.Error("check probe host ip invalid")
		return false
	}

	if httpGet.Port.IntVal < minPort || httpGet.Port.IntVal > maxPort {
		hwlog.RunLog.Error("check probe port invalid")
		return false
	}

	return true
}
