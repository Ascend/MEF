// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"path/filepath"
	"strings"
	"sync"

	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
)

var (
	once            sync.Once
	whiteListMap    = map[string]struct{}{}
	modelFilePrefix = "/var/lib/docker/modelfile/"
	modelFileDir    = "/var/lib/docker/modelfile"
)

// GetHostNameMapPath [method] for getting hostPath volume of container volumes
func GetHostNameMapPath(volumes []v1.Volume) []string {
	var hostNameMapPath []string
	for _, volume := range volumes {
		if volume.VolumeSource.HostPath == nil {
			continue
		}
		hostNameMapPath = append(hostNameMapPath, volume.VolumeSource.HostPath.Path)
	}
	return hostNameMapPath
}

func initWhiteList(whiteList []string) {
	once.Do(func() {
		for _, white := range whiteList {
			whiteListMap[white] = struct{}{}
		}
	})
}

// InFdWhiteList [method] for check out if target path is in whitelist
func InFdWhiteList(hostPath string, whiteList []string) bool {
	initWhiteList(whiteList)
	if strings.HasPrefix(hostPath, modelFilePrefix) {
		_, ok := whiteListMap[constants.ModeFileActiveDir]
		return ok
	}
	_, ok := whiteListMap[filepath.Clean(hostPath)]
	return ok
}

// InCenterWhiteList [method] for check out if target path is in whitelist Center message
func InCenterWhiteList(hostPath string, whiteList []string) bool {
	initWhiteList(whiteList)
	// message from mef center do not allow to use model file yet
	if hostPath == modelFileDir {
		return false
	}
	_, ok := whiteListMap[filepath.Clean(hostPath)]
	return ok
}
