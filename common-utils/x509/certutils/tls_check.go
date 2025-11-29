// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package certutils

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	hwx509 "huawei.com/mindx/common/x509"
)

func checkCaFileInCrl(verifiedChains [][]*x509.Certificate, crl *pkix.CertificateList) error {
	if crl == nil {
		return nil
	}

	for _, chain := range verifiedChains {
		if err := checkSingleChain(chain, crl); err != nil {
			return err
		}
	}

	return nil
}

func checkSingleChain(chain []*x509.Certificate, crl *pkix.CertificateList) error {
	for idx, cert := range chain {
		if cert == nil {
			continue
		}
		// root ca can revoke mid ca; mid ca can revoke work cert
		if cert.CheckCRLSignature(crl) != nil || idx == 0 {
			// the CRL is not signed by this cert
			// idx == 0 is the working certificat, it only could be revoked by Issuer
			continue
		}
		for _, revoked := range crl.TBSCertList.RevokedCertificates {
			// example: root ca revoke mid ca
			if revoked.SerialNumber.Cmp(chain[idx-1].SerialNumber) == 0 {
				hwlog.RunLog.Errorf("cert [%s] revoked by [%s]\n", chain[idx-1].Subject, cert.Subject)
				return fmt.Errorf("the peer certificate has been revoked")
			}
		}
	}
	return nil
}

func checkTlsCrl(tlsCertInfo TlsCertInfo) error {
	if !tlsCertInfo.GetCrlData().ContainsData() {
		return nil
	}
	if _, err := hwx509.ParseCrls(tlsCertInfo.GetCrlData()); err != nil {
		return fmt.Errorf("load crl data failed, %v", err)
	}
	return nil
}

func checkTlsKey(tlsCertInfo TlsCertInfo) error {
	if tlsCertInfo.KeyPath == "" {
		return nil
	}
	return CheckKey(tlsCertInfo.KeyPath, tlsCertInfo.KmcCfg)
}

func checkTlsRootCa(tlsCertInfo TlsCertInfo) error {
	if tlsCertInfo.RootCaPath == "" {
		return nil
	}
	if _, err := GetCertContent(tlsCertInfo.RootCaPath); err != nil {
		return err
	}
	return nil
}

func checkTlsCert(tlsCertInfo TlsCertInfo) error {
	if tlsCertInfo.CertPath == "" {
		return nil
	}
	if _, err := GetCertContent(tlsCertInfo.CertPath); err != nil {
		return err
	}
	return nil
}

// CheckKey check whether a key path with tls config can be loaded and decrypted
func CheckKey(path string, kmcConfig *kmc.SubConfig) error {
	keyBytes, err := fileutils.LoadFile(path)
	if err != nil {
		return fmt.Errorf("load key file from path [%s] failed", path)
	}
	decryptKeyByte, err := kmc.DecryptContent(keyBytes, kmcConfig)
	if err != nil {
		return fmt.Errorf("load key file from path [%s] failed, decrypt key content failed", path)
	}
	defer utils.ClearSliceByteMemory(decryptKeyByte)
	return nil
}
