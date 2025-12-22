// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

var certCrlPairCache sync.Map

// CertCrlPair represents a pair of ca and crl
type CertCrlPair struct {
	CertPEM string
	CrlPEM  string
}

func init() {
	certCrlPairCache.Store(common.SoftwareCertName, nil)
	certCrlPairCache.Store(common.ImageCertName, nil)
}

// GetCertCrlPairCache return ca and crl by name from cache,
// which supposed return error only in certs request failed form edge-manager to cert-manager.
func GetCertCrlPairCache(name string) (CertCrlPair, error) {
	hwlog.RunLog.Infof("start to get cert [%s] from cache", name)
	// cert name error case, check name first
	cert, ok := certCrlPairCache.Load(name)
	if !ok {
		return CertCrlPair{}, errors.New("unknown cert name")
	}
	// if cache is nil, get form cert-manager
	if cert == nil {
		return getCertFromCertMgr(name)
	}
	certCrlPair, ok := cert.(CertCrlPair)
	if !ok {
		return CertCrlPair{}, errors.New("unknown type of certificate's content")
	}
	return certCrlPair, nil
}

// SetCertCrlPairCache refresh ca and crl cache by name
// which will be called when the certs(image, software) is updated in cert-manager.
func SetCertCrlPairCache(name, certPEM, crlPEM string) {
	_, ok := certCrlPairCache.Load(name)
	if !ok {
		return
	}
	hwlog.RunLog.Infof("start to set cert [%s] to cache", name)
	certCrlPairCache.Store(name, CertCrlPair{CertPEM: certPEM, CrlPEM: crlPEM})
}

func getCertFromCertMgr(name string) (CertCrlPair, error) {
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
		return CertCrlPair{}, err
	}
	hwlog.RunLog.Infof("request [%s] cert cache from cert-manager successful", name)

	crlStr, err := reqCertParams.GetCrl(name)
	if err != nil {
		hwlog.RunLog.Errorf("request [%s] crl cache from cert-manager failed, %v", name, err)
		return CertCrlPair{}, err
	}

	certCrlPair := CertCrlPair{CertPEM: certStr, CrlPEM: crlStr}
	certCrlPairCache.Store(name, certCrlPair)
	return certCrlPair, nil
}
