// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util for edge-manager
package util

import (
	"errors"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"k8s.io/api/core/v1"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/kubeclient"
)

const (
	secretLen    = 3
	authPosition = 1
	secretSplit  = `:{`
	secretTrim   = `"`
)

// GetImageAddress get image address
func GetImageAddress() (string, error) {
	secret, err := kubeclient.GetKubeClient().GetSecret(kubeclient.DefaultImagePullSecretKey)
	if err != nil {
		hwlog.RunLog.Error("get image pull secret from k8s failed")
		return "", errors.New("get image pull secret from k8s failed")
	}
	secretByte, ok := secret.Data[v1.DockerConfigJsonKey]
	if !ok {
		hwlog.RunLog.Error("get data of image pull secret failed")
		return "", errors.New("get data of image pull secret failed")
	}
	secretStr := string(secretByte)
	if secretStr == kubeclient.DefaultImagePullSecretValue {
		hwlog.RunLog.Warnf("secret %s is not configured yet", kubeclient.DefaultImagePullSecretKey)
		return "", nil
	}

	authSli := strings.Split(secretStr, secretSplit)
	defer func() {
		utils.ClearStringMemory(secretStr)
		common.ClearSliceByteMemory(secretByte)
		for i := 0; i < len(authSli); i++ {
			utils.ClearStringMemory(authSli[i])
		}
	}()
	if len(authSli) < secretLen {
		hwlog.RunLog.Error("parse secret content failed")
		return "", errors.New("parse secret content failed")
	}
	resBytes := make([]byte, len(authSli[authPosition]))
	copy(resBytes, authSli[authPosition])
	return strings.Trim(string(resBytes), secretTrim), nil
}
