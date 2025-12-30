// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package websocketmgr

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"
)

// TestMain - test cases entry
func TestMain(m *testing.M) {
	patch := gomonkey.ApplyFunc(certutils.GetTlsCfgWithPath, replaceGetTlsCfgWithPath)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patch)
}

// replace "GetTlsCfgWithPath" for getting rid of kmc module
func replaceGetTlsCfgWithPath(tlsCertInfo certutils.TlsCertInfo) (*tls.Config, error) {
	caCert, err := os.ReadFile(tlsCertInfo.RootCaPath)
	if err != nil {
		fmt.Printf("Read ca cert error: %v", err)
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	certs, err := tls.LoadX509KeyPair(tlsCertInfo.CertPath, tlsCertInfo.KeyPath)
	if err != nil {
		fmt.Printf("Load X509 Key Pair error: %v", err)
		return nil, err
	}
	return &tls.Config{
		Rand:               rand.Reader,
		Certificates:       []tls.Certificate{certs},
		ClientCAs:          certPool,
		RootCAs:            certPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}, nil
}
