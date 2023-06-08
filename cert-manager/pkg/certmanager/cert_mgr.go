// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"cert-manager/pkg/certmanager/certchecker"
)

var (
	lock            sync.Mutex
	certsToImported = map[string]*time.Timer{
		common.ImageCertName:    nil,
		common.SoftwareCertName: nil,
		common.WsCltName:        nil,
		common.NorthernCertName: nil,
	}
)

func isCertImported(certName string) bool {
	const certQueryLogInterval = 60 * time.Second

	caFilePath := getRootCaPath(certName)
	timer, exist := certsToImported[certName]
	if !exist || utils.IsExist(caFilePath) {
		return true
	}

	if timer == nil {
		timer = time.NewTimer(certQueryLogInterval)
		certsToImported[certName] = timer
		hwlog.RunLog.Warnf("%s cert is not be imported yet", certName)
	}
	select {
	case _, ok := <-timer.C:
		if !ok {
			hwlog.RunLog.Error("cert query suppression timer's channel is unexpected closed")
			return false
		}
		timer.Reset(certQueryLogInterval)
	default:
	}
	return false
}

// getCertByCertName query root ca with cert name
func getCertByCertName(certName string) ([]byte, error) {
	caFilePath := getRootCaPath(certName)
	certData, err := utils.LoadFile(caFilePath)
	if certData == nil {
		hwlog.RunLog.Errorf("load root cert failed: %v", err)
		return nil, errors.New("load root cert failed")
	}
	return certData, nil
}

// issueServiceCert issue service certificate with csr file, only support pem type csr
func issueServiceCert(certName string, serviceCsr string) ([]byte, error) {
	csrByte, err := base64.StdEncoding.DecodeString(serviceCsr)
	if err != nil {
		hwlog.RunLog.Error("base64 decode service csr failed")
		return nil, errors.New("base64 decode service csr failed")
	}
	srvDer, _ := pem.Decode(csrByte)
	if srvDer == nil {
		hwlog.RunLog.Error("Decode service csr pem failed")
		return nil, errors.New("decode service csr pem failed")
	}

	keyFilePath := getRootKeyPath(certName)
	caFilePath := getRootCaPath(certName)
	initCertMgr := certutils.InitRootCertMgr(caFilePath, keyFilePath, certName, nil)
	if _, err := initCertMgr.GetRootCaPair(); err != nil {
		if _, err = initCertMgr.NewRootCa(); err != nil {
			hwlog.RunLog.Errorf("Init root ca info failed: %v", err)
			return nil, err
		}
	}

	certBytes, err := initCertMgr.IssueServiceCert(srvDer.Bytes)
	if err != nil {
		hwlog.RunLog.Errorf("issue service cert info failed: %v", err)
		return nil, err
	}

	certPem := certutils.PemWrapCert(certBytes)
	hwlog.RunLog.Info("issue service cert success")
	return certPem, nil
}

// saveCaContent save ca content to File
func saveCaContent(certName string, caContent []byte) error {
	caFilePath := getRootCaPath(certName)
	if err := utils.MakeSureDir(caFilePath); err != nil {
		hwlog.RunLog.Errorf("create %s ca folder failed, error: %v", certName, err)
		return fmt.Errorf("create %s ca folder failed, error: %v", certName, err)
	}
	if err := common.WriteData(caFilePath, caContent); err != nil {
		hwlog.RunLog.Errorf("save %s cert file failed, error:%s", certName, err)
		return fmt.Errorf("save %s ca file failed", certName)
	}
	hwlog.RunLog.Infof("save %s ca file success", certName)
	return nil
}

// removeCaFile delete ca File
func removeCaFile(certName string) error {
	caFilePath := getRootCaPath(certName)
	if utils.IsExist(caFilePath) {
		// remove the ca file
		if err := common.DeleteFile(caFilePath); err != nil {
			hwlog.RunLog.Errorf("remove %s ca file failed, error: %v", certName, err)
			return fmt.Errorf("remove %s ca file failed, error: %v", certName, err)
		}
	}
	hwlog.RunLog.Infof("delete %s ca file success", caFilePath)
	return nil
}

func updateClientCert(certName, certOpt string, certContent []byte) error {
	lock.Lock()
	defer lock.Unlock()
	hwlog.RunLog.Info("start update cert file")
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
			KmcCfg:     nil,
		},
	}
	cert := certutils.UpdateClientCert{
		CertName:    certName,
		CertContent: certContent,
		CertOpt:     certOpt,
	}
	certName, err := reqCertParams.UpdateCertFile(cert)
	if err != nil {
		hwlog.RunLog.Errorf("update %s ca file failed, error:%v", certName, err)
		return fmt.Errorf("update %s ca file failed", certName)
	}
	return nil
}

// ExportRootCa export cert file
func ExportRootCa(c *gin.Context) {
	hwlog.OpLog.Info("export cert file start")
	certName := c.Query("certName")
	if !certchecker.CheckIfCanExport(certName) {
		msg := fmt.Sprintf("export cert [%s] root ca not support", certName)
		hwlog.OpLog.Errorf(msg)
		common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
		return
	}
	ca, err := getCertByCertName(certName)
	if err != nil {
		msg := fmt.Sprintf("get cert [%s] root ca failed", certName)
		hwlog.OpLog.Errorf("%s, error:%v", msg, err)
		common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
	c.Header(common.ContentType, "text/plain; charset=utf-8")
	c.Header(common.TransferEncoding, "chunked")
	c.Header(common.ContentDisposition, fmt.Sprintf("attachment; filename=%s", util.RootCaFileName))
	c.Writer.WriteHeaderNow()
	if _, err := c.Writer.Write(ca); err != nil {
		msg := fmt.Sprintf("export cert [%s] root ca failed", certName)
		hwlog.OpLog.Errorf("%s, error: %v", msg, err)
		common.ConstructResp(c, common.ErrorExportRootCa, msg, nil)
		return
	}
	c.Writer.Flush()
	hwlog.OpLog.Infof("export cert [%s] root ca success", certName)
}
