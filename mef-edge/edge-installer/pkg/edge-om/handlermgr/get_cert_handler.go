// Copyright (c) 2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

// getCertHandler prepare cert for file download.
// inner message handler, so do not need to record operation logs
type getCertHandler struct {
}

// Handle [method] handle message
func (g getCertHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle get cert message form edge-main")
	var req config.CertReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse cert info para failed: %v", err)
		return errors.New("parse cert info para failed")
	}
	var processErr error
	certResp := config.CertResp{CertReq: req}
	switch req.CertName {
	case constants.SoftwareCertName:
		certResp.CertContent, processErr = getSoftwareCert()
		if processErr != nil {
			hwlog.RunLog.Errorf("get software cert failed: %v", processErr)
			certResp.ErrorMsg = processErr.Error()
			break
		}
		certResp.CrlContent, processErr = getSoftwareCrl()
		if processErr != nil {
			hwlog.RunLog.Errorf("get software crl failed: %v", processErr)
			certResp.ErrorMsg = processErr.Error()
		}
	default:
		processErr = errors.New("invalid cert name")
		hwlog.RunLog.Errorf("invalid cert name: %v", req.CertName)
		certResp.ErrorMsg = fmt.Sprintf("invalid cert name: %v", req.CertName)
	}

	resp, err := msg.NewResponse()
	if err != nil {
		return err
	}
	resp.SetRouter(constants.ModEdgeOm, constants.InnerClient, constants.OptResponse, constants.InnerCert)
	if err = resp.FillContent(certResp, true); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return errors.New("fill resp into content failed")
	}
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("response msg failed: %v", err)
		return err
	}
	return processErr
}

func getSoftwareCert() ([]byte, error) {
	hwlog.RunLog.Info("start to get software cert for software download")
	certRootDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		return nil, fmt.Errorf("get software root cert failed, error: %v", err)
	}
	rootCaPath := filepath.Join(certRootDir, constants.SoftwareCertName, constants.RootCaName)

	certContent, err := certutils.GetCertContentWithBackup(rootCaPath)
	if err != nil {
		return nil, fmt.Errorf("get software cert failed, error: %v", err)
	}
	hwlog.RunLog.Info("successfully get software cert for software download")
	return certContent, nil
}

func getSoftwareCrl() ([]byte, error) {
	hwlog.RunLog.Info("start to get software crl for software download")
	certRootDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		return nil, fmt.Errorf("get software root cert failed, error: %v", err)
	}
	crlPath := filepath.Join(certRootDir, constants.SoftwareCertName, constants.CrlName)
	// if CRL file not exists, skip resp it to edge-main
	if !fileutils.IsExist(crlPath) {
		return nil, nil
	}
	certContent, err := certutils.GetCrlContentWithBackup(crlPath)
	if err != nil {
		return nil, fmt.Errorf("get software crl failed, error: %v", err)
	}
	hwlog.RunLog.Info("successfully get software crl for software download")
	return certContent, nil
}
