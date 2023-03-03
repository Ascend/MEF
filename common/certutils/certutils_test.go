// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certutils for
package certutils

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

func makeSingCert() error {
	testDir := "/tmp/mef-test-certs/"
	err := utils.MakeSureDir(testDir)
	if err != nil {
		fmt.Printf("make sure test dir failed: %v", err)
		return err
	}
	defer func() {
		err := os.RemoveAll(testDir)
		if err != nil {
			return
		}
	}()
	rootCaFilePath := path.Join(testDir, "test_root.ca")
	rootPrivFilePath := path.Join(testDir, "test_root.key")
	kmcCfg := &common.KmcCfg{
		SdpAlgID:       common.Aes256gcm,
		PrimaryKeyPath: path.Join(testDir, "master.ks"),
		StandbyKeyPath: path.Join(testDir, "backup.ks"),
		DoMainId:       common.DoMainId,
	}
	initCertMgr := InitRootCertMgr(rootCaFilePath, rootPrivFilePath, "MEF Center test", kmcCfg)
	if _, err := initCertMgr.NewRootCa(); err != nil {
		fmt.Printf("init new root ca failed: %v", err)
		return err
	}
	svrCertPath := path.Join(testDir, "test_cert.crt")
	svcKeyPath := path.Join(testDir, "test_cert.key")
	componentCert := SelfSignCert{
		RootCertMgr: initCertMgr,
		SvcCertPath: svrCertPath,
		SvcKeyPath:  svcKeyPath,
		CommonName:  "MEF Test",
		San:         CertSan{DnsName: []string{"MEF TEST DNS"}, IpAddr: []net.IP{net.ParseIP("127.0.0.1")}},
		KmcCfg: &common.KmcCfg{
			SdpAlgID:       common.Aes256gcm,
			PrimaryKeyPath: path.Join(testDir, "master.ks"),
			StandbyKeyPath: path.Join(testDir, "backup.ks"),
			DoMainId:       common.DoMainId,
		},
	}
	if err := componentCert.CreateSignCert(); err != nil {
		fmt.Printf("create sign cert failed: %v", err)
		return err
	}
	err = getTlsCfgWithPath(rootCaFilePath, svrCertPath, svcKeyPath, testDir)
	if err != nil {
		return err
	}
	return nil
}

func getTlsCfgWithPath(rootCaFilePath, svrCertPath, svcKeyPath, testDir string) error {
	tlsCertPath := TlsCertInfo{
		RootCaPath: rootCaFilePath,
		CertPath:   svrCertPath,
		KeyPath:    svcKeyPath,
		SvrFlag:    true,
		KmcCfg: &common.KmcCfg{
			SdpAlgID:       common.Aes256gcm,
			PrimaryKeyPath: path.Join(testDir, "master.ks"),
			StandbyKeyPath: path.Join(testDir, "backup.ks"),
			DoMainId:       common.DoMainId,
		},
	}
	_, err := GetTlsCfgWithPath(tlsCertPath)
	if err != nil {
		return err
	}
	return nil
}

func TestCertUtils(t *testing.T) {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
	convey.Convey("test make sign cert", t, func() {
		res := makeSingCert()
		convey.So(res, convey.ShouldBeNil)
	})
}
