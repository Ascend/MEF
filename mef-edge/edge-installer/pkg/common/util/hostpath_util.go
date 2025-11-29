// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
