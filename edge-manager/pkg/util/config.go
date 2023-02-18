// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package util for edge-manager
package util

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	if err != nil || len(secret.Data) == 0 {
		return "", errors.New("get image pull secret from k8s failed")
	}
	secretByte, ok := secret.Data[v1.DockerConfigJsonKey]
	if !ok {
		return "", errors.New("get data of image pull secret failed")
	}
	defer func() {
		common.ClearSliceByteMemory(secretByte)
	}()
	authSli := strings.Split(string(secretByte), secretSplit)
	if len(authSli) != secretLen {
		return "", errors.New("parse secret content failed")
	}
	return strings.Trim(authSli[authPosition], secretTrim), nil
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
	var rootCaRes string
	var err error
	for i := 0; i < certutils.DefaultCertRetryTime; i++ {
		rootCaRes, err = reqCertParams.GetRootCa(certName)
		if err == nil {
			break
		}
		time.Sleep(certutils.DefaultCertWaitTime)
	}
	if rootCaRes == "" {
		return certutils.QueryCertRes{}, fmt.Errorf("get %s cert content failed, error: %v", certName, err)
	}
	queryCertRes := certutils.QueryCertRes{
		CertName: certName,
		Cert:     rootCaRes,
	}
	return queryCertRes, nil
}
