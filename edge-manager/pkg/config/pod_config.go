// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
)

// PodConfig is the cache of pod config
var PodConfig podConfigInfo

// podConfigInfo [struct] for save pod config
type podConfigInfo struct {
	HostPath                []string
	MaxPodNumberPerNode     int64 `json:"maxPodNumberPerNode,omitempty"`
	MaxDsNumberPerNodeGroup int64 `json:"maxDsNumberPerNodeGroup,omitempty"`
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
				"path [%s] is invalid [%s], and won't be effective", hostPath[i], checkResult.Reason)
			continue
		}
		hostPathTmp = append(hostPathTmp, filepath.Clean(hostPath[i]))
	}
	return hostPathTmp
}

// CheckAndModifyMaxLimitNumber [method] init max pod/daemonSet number for check
func CheckAndModifyMaxLimitNumber(number int64) int64 {
	const (
		defaultMaxLimitNumber = 20
		minLimitNumber        = 1
		maxLimitNumber        = 128
	)
	if number < minLimitNumber || number > maxLimitNumber {
		return defaultMaxLimitNumber
	}
	return number
}
