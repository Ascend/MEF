// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"errors"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"

	"edge-manager/pkg/constants"
)

var certCache sync.Map

func init() {
	certCache.Store(common.SoftwareCertName, nil)
	certCache.Store(common.ImageCertName, nil)
}

// GetCertCache return cert cache by name,
// which supposed return error only in certs request failed form edge-manager to cert-manager.
func GetCertCache(name string) (string, error) {
	hwlog.RunLog.Infof("start to get cert [%s] from cache", name)
	// cert name error case, check name first
	cert, ok := certCache.Load(name)
	if !ok {
		return "", errors.New("unknown cert name")
	}
	// if cache is nil, get form cert-manager
	if cert == nil {
		return getCertFromCertMgr(name)
	}
	certStr, ok := cert.(string)
	if !ok {
		return "", errors.New("unknown type of certificate's content")
	}
	return certStr, nil
}

// SetCertCache refresh cert cache by name
// which will be called when the certs(image, software) is updated in cert-manager.
func SetCertCache(name, content string) {
	_, ok := certCache.Load(name)
	if !ok {
		return
	}
	hwlog.RunLog.Infof("start to set cert [%s] to cache", name)
	certCache.Store(name, content)
}

func getCertFromCertMgr(name string) (string, error) {
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: constants.RootCaPath,
			CertPath:   constants.ServerCertPath,
			KeyPath:    constants.ServerKeyPath,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
	certStr, err := reqCertParams.GetRootCa(name)
	if err != nil {
		hwlog.RunLog.Errorf("request [%s] cert cache from cert-manager failed, %v", name, err)
		return "", err
	}
	hwlog.RunLog.Infof("request [%s] cert cache from cert-manager successful", name)
	certCache.Store(name, certStr)
	return certStr, nil
}
