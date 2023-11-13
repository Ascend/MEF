// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util for edge-manager
package util

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"k8s.io/api/core/v1"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"

	"edge-manager/pkg/constants"
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

// GetCertContent get cert content
func GetCertContent(certName string) (certutils.ClientCertResp, error) {
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
	rootCaRes, err := reqCertParams.GetRootCa(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert content from cert-manager failed, error: %v", err)
		return certutils.ClientCertResp{}, fmt.Errorf("query cert content from cert-manager failed, error: %v", err)
	}
	queryCertRes := certutils.ClientCertResp{
		CertName:    certName,
		CertContent: rootCaRes,
		CertOpt:     common.Update,
	}
	return queryCertRes, nil
}
