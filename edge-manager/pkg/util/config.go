// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util for edge-manager
package util

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
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
		hwlog.RunLog.Errorf("get image pull secret failed, error:%v", err)
		return "", err
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

	defer func() {
		common.ClearSliceByteMemory(secretByte)
	}()
	authSli := strings.Split(secretStr, secretSplit)
	if len(authSli) < secretLen {
		hwlog.RunLog.Error("parse secret content failed")
		return "", errors.New("parse secret content failed")
	}
	return strings.Trim(authSli[authPosition], secretTrim), nil
}

// GetCertContent get cert content
func GetCertContent(certName string) (certutils.ClientCertResp, error) {
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
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
