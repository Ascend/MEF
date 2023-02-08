// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils
package certutils

import (
	"errors"
	"fmt"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// SelfSignCert self singed cert struct
type SelfSignCert struct {
	RootCertMgr *RootCertMgr
	SvcCertPath string
	SvcKeyPath  string
	CommonName  string
	KmcCfg      *common.KmcCfg
	San         CertSan
}

// CreateSignCert create a new singed cert for root ca and service cert
func (sc *SelfSignCert) CreateSignCert() error {
	if sc.RootCertMgr == nil {
		return errors.New("root cert mgr is nil, can not create sign cert")
	}
	if _, getErr := sc.RootCertMgr.GetRootCaPair(); getErr != nil {
		hwlog.RunLog.Warnf("get root ca pair failed: %s, start to create new ca", getErr)
		if _, err := sc.RootCertMgr.NewRootCa(); err != nil {
			return fmt.Errorf("get root ca pair for create sign cert failed, "+
				"get root failed [%v] and new root failed [%v]", getErr, err)
		}
	}
	csr, err := CreateCsr(sc.SvcKeyPath, sc.CommonName, sc.KmcCfg, sc.San)
	if err != nil {
		return err
	}

	certBytes, err := sc.RootCertMgr.IssueServiceCert(csr)
	if err != nil {
		return err
	}

	err = saveCertWithPem(sc.SvcCertPath, certBytes)
	if err != nil {
		return errors.New("save self singed cert with pem failed: " + err.Error())
	}

	return nil
}
