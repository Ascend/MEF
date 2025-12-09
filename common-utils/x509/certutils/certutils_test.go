// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certutils for
package certutils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
)

func makeSingCert() error {
	testDir := "/tmp/mef-test-certs/"
	if err := fileutils.MakeSureDir(testDir); err != nil {
		fmt.Printf("make sure test dir failed: %v", err)
		return err
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			return
		}
	}()
	rootCaFilePath := path.Join(testDir, "test_root.ca")
	rootPrivFilePath := path.Join(testDir, "test_root.key")
	kmcCfg := &kmc.SubConfig{
		SdpAlgID:       kmc.Aes256gcmId,
		PrimaryKeyPath: path.Join(testDir, "master.ks"),
		StandbyKeyPath: path.Join(testDir, "backup.ks"),
		DoMainId:       kmc.DefaultDoMainId,
	}
	initCertMgr := InitRootCertMgr(rootCaFilePath, rootPrivFilePath, "MEF-MEF Center test", kmcCfg)
	if _, err := initCertMgr.NewRootCa(); err != nil {
		fmt.Printf("init new root ca failed: %v", err)
		return err
	}
	svrCertPath := path.Join(testDir, "test_cert.crt")
	svcKeyPath := path.Join(testDir, "test_cert.key")
	componentCert := SelfSignCert{
		RootCertMgr:      initCertMgr,
		SvcCertPath:      svrCertPath,
		SvcKeyPath:       svcKeyPath,
		CommonNamePrefix: "MEF Test",
		San:              CertSan{DnsName: []string{"MEF TEST DNS"}, IpAddr: []net.IP{net.ParseIP("127.0.0.1")}},
		KmcCfg: &kmc.SubConfig{
			SdpAlgID:       kmc.Aes256gcmId,
			PrimaryKeyPath: path.Join(testDir, "master.ks"),
			StandbyKeyPath: path.Join(testDir, "backup.ks"),
			DoMainId:       kmc.DefaultDoMainId,
		},
	}
	if err := componentCert.CreateSignCert(); err != nil {
		fmt.Printf("create sign cert failed: %v", err)
		return err
	}
	if err := getTlsCfg(rootCaFilePath, svrCertPath, svcKeyPath, testDir); err != nil {
		return err
	}
	return nil
}

func getTlsCfg(rootCaFilePath, svrCertPath, svcKeyPath, testDir string) error {
	tlsCertPath := TlsCertInfo{
		RootCaPath: rootCaFilePath,
		CertPath:   svrCertPath,
		KeyPath:    svcKeyPath,
		SvrFlag:    true,
		KmcCfg: &kmc.SubConfig{
			SdpAlgID:       kmc.Aes256gcmId,
			PrimaryKeyPath: path.Join(testDir, "master.ks"),
			StandbyKeyPath: path.Join(testDir, "backup.ks"),
			DoMainId:       kmc.DefaultDoMainId,
		},
	}
	if _, err := GetTlsCfgWithPath(tlsCertPath); err != nil {
		return err
	}
	return nil
}

func TestCertUtils(t *testing.T) {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitHwLogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
	convey.Convey("test make sign cert", t, func() {
		res := makeSingCert()
		convey.So(res, convey.ShouldNotBeNil)
	})
}

const (
	normalCrl = `-----BEGIN X509 CRL-----
MIICUTCBugIBATANBgkqhkiG9w0BAQsFADB3MQswCQYDVQQGEwJDTjESMBAGA1UE
CAwJR3Vhbmdkb25nMREwDwYDVQQHDAhTaGVuemhlbjEVMBMGA1UECgwMVGVzdCBD
b21wYW55MRgwFgYDVQQLDA9UZXN0IERlcGFydG1lbnQxEDAOBgNVBAMMB1Rlc3Qg
Q0EXDTI0MDUxNzA3MDYwM1oXDTM0MDUxNTA3MDYwM1qgDzANMAsGA1UdFAQEAgIQ
ADANBgkqhkiG9w0BAQsFAAOCAYEAXcuhqY+46ltmRxyODEk4CWNJySYHXDyLTtYa
6NaCudDeJvy8sldrcQzApztJymScaIPAvVI6EOQ2JkKKh6hR1mgUXtxdnz/M7vZF
DiMmAZAwqPjDFQmDqUto46UD/+SHCxUfVnq4oWOVInOIFxDyQEn5EcvUh6nI9K3h
XHLeLE1DVop0e0aiid/S2RA0LojFPQEY3UbCK2qmGnBENDXrEV6jpUfihhhqTIkf
0j8J+am32vw4ZqEWrx+PTen8OZx+xPi9YDNDRPcnqyYbSKuPHivvYHvdM+xNBTrg
XdQUMudgRZlzyTvIv584Mje2MFHZQzx8ULAFnmU8SFsO1k3VPdVM0TlQ882ULPDw
J2BdDQ67NPzNCttEbNdQdkh2pQLuwNd3jSr/S5+S1rr9gtVIbYyS5S6R1xop+eAe
mtxO2/91vbtvCu+Cu0gtJzOTJt/lw94WuJRDOxhDrLAi8J67eJnhqpyNhi0lOHJy
Uz76cZjj6dtKBB0Gg8oLj+ESJj5z
-----END X509 CRL-----`
	invalidCrl = ""
)

func TestGetCrlContentWithBackup(t *testing.T) {
	testcases := []struct {
		description string
		mainData    string
		backupData  string
		backupErr   error
		restoreErr  error
		wantErr     bool
	}{
		{
			description: "get crl successfully",
			mainData:    normalCrl,
		},
		{
			description: "restore crl successfully",
			mainData:    invalidCrl,
			backupData:  normalCrl,
		},
		{
			description: "backup crl failed",
			mainData:    normalCrl,
			backupErr:   errors.New("backup error"),
		},
		{
			description: "restore crl failed",
			mainData:    invalidCrl,
			restoreErr:  errors.New("restore error"),
			wantErr:     true,
		},
	}

	for _, tc := range testcases {
		convey.Convey(tc.description, t, func() {
			loadFileSeq := []gomonkey.OutputCell{
				{Values: []interface{}{[]byte(tc.mainData), nil}, Times: 1},
				{Values: []interface{}{[]byte(tc.backupData), nil}, Times: 1}}
			patches := gomonkey.ApplyFuncSeq(fileutils.LoadFile, loadFileSeq).
				ApplyFuncReturn(backuputils.BackUpFiles, tc.backupErr).
				ApplyFuncReturn(backuputils.RestoreFiles, tc.restoreErr)
			defer patches.Reset()

			_, err := GetCrlContentWithBackup("")
			verb := convey.ShouldBeNil
			if tc.wantErr {
				verb = convey.ShouldNotBeNil
			}
			convey.So(err, verb)
		})
	}
}
