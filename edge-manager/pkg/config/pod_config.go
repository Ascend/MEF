// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"edge-manager/pkg/util"
	"huawei.com/mindx/common/hwlog"
)

// PodConfig is the cache of pod config
var PodConfig podConfigInfo

// podConfigInfo [struct] for save pod config
type podConfigInfo struct {
	HostPath []string
}

// CheckAndModifyHostPath [method] do check and modification job
func CheckAndModifyHostPath(hostPath []string) []string {
	const defaultMaxHostPathNumber = 256
	whiteListNumber := len(hostPath)
	if whiteListNumber > defaultMaxHostPathNumber {
		whiteListNumber = defaultMaxHostPathNumber
	}

	hostPathTmp := make([]string, 0, whiteListNumber)
	for i := 0; i < whiteListNumber; i++ {
		if checkResult := util.GetPathChecker("", true).Check(hostPath[i]); !checkResult.Result {
			hwlog.RunLog.Errorf("checking pod config host path, "+
				"path [%s] is invalid, and won't be effective", checkResult.Reason)
			continue
		}
		hostPathTmp = append(hostPathTmp, hostPath[i])
	}
	return hostPathTmp
}
