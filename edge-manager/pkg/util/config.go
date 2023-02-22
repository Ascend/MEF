// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util for edge-manager
package util

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/kubeclient"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

const (
	secretLen    = 3
	authPosition = 1
	secretSplit  = `:{`
	secretTrim   = `"`
)

// GetImageAddress get image address
func GetImageAddress() (string, error) {
	secret, err := kubeclient.GetKubeClient().GetSecret(certutils.DefaultSecretName)
	if secretNotFound(err) {
		hwlog.RunLog.Warnf("secrets %s not found", certutils.DefaultSecretName)
		return "", nil
	}
	if err != nil {
		hwlog.RunLog.Errorf("get image pull secret failed, error:%v", err)
		return "", err
	}
	secretByte, ok := secret.Data[v1.DockerConfigJsonKey]
	if !ok {
		hwlog.RunLog.Error("get data of image pull secret failed")
		return "", errors.New("get data of image pull secret failed")
	}
	defer func() {
		common.ClearSliceByteMemory(secretByte)
	}()
	authSli := strings.Split(string(secretByte), secretSplit)
	if len(authSli) < secretLen {
		hwlog.RunLog.Error("parse secret content failed")
		return "", errors.New("parse secret content failed")
	}
	return strings.Trim(authSli[authPosition], secretTrim), nil
}

func secretNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), certutils.SecretNotFound)
}

// GetCertContent get cert content
func GetCertContent(certName string) (certutils.QueryCertRes, error) {
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    RootCaPath,
			CertPath:      ServerCertPath,
			KeyPath:       ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	rootCaRes, err := reqCertParams.GetRootCa(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert content from cert-manager failed, error: %v", err)
		return certutils.QueryCertRes{}, fmt.Errorf("query cert content from cert-manager failed, error: %v", err)
	}
	queryCertRes := certutils.QueryCertRes{
		CertName: certName,
		Cert:     rootCaRes,
	}
	return queryCertRes, nil
}
