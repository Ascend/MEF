// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package downloadmgr

import (
	"errors"
	"fmt"
	"math"
	"time"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
)

const (
	defaultTimeOutInterval = time.Second * 30
	progressPreparing      = 10
	progressDownloading    = 40
	progressVerifying      = 70
	progressSuccess        = 100
	minPwdLength           = 8
	maxPwdLength           = 20
)

type downloadProcess struct {
	cert            []byte
	crlContent      []byte
	progress        uint64
	sfwDownloadInfo util.SoftwareDownloadInfo
}

type edgeReportUpgradeResInfo struct {
	SerialNumber string              `json:"serialNumber"`
	ProgressInfo config.ProgressInfo `json:"upgradeResInfo"`
}

func (d *downloadMgr) processDownloadSoftware(msg model.Message) error {
	hwlog.RunLog.Info("start to process download software")
	var dp downloadProcess
	if err := msg.ParseContent(&dp.sfwDownloadInfo); err != nil {
		hwlog.RunLog.Errorf("get download process failed: %v", err)
		return errors.New("get download process failed")
	}
	defer utils.ClearSliceByteMemory(dp.sfwDownloadInfo.DownloadInfo.Password)

	processTasks := []func() error{
		dp.checkDownloadInfo,
		dp.getSoftwareCert,
		dp.prepareDownloadDir,
		dp.checkDownloadDir,
		dp.downloadSoftware,
		dp.verifyAndUnpack,
	}

	for _, task := range processTasks {
		err := task()
		reportDownloadProcess(dp.progress, err)
		if err != nil {
			dp.cleanDownloadDir()
			hwlog.RunLog.Errorf("process download software failed, %v", err)
			return err
		}
	}
	return nil
}

func (dp *downloadProcess) checkDownloadInfo() error {
	infoChecker := checker.GetAndChecker(checker.GetStringChoiceChecker("SoftwareName",
		[]string{constants.MEFEdgeName}, true),
		&checker.ModelChecker{
			Field:    "DownloadInfo",
			Required: true,
			Checker: checker.GetAndChecker(
				checker.GetHttpsUrlChecker("Package", true, true),
				checker.GetHttpsUrlChecker("SignFile", true, true),
				checker.GetHttpsUrlChecker("CrlFile", true, true),
				checker.GetRegChecker("UserName", "^[a-zA-Z0-9]{6,32}$", true),
				checker.GetListChecker("Password",
					checker.GetUintChecker("", 0, math.MaxUint8, true),
					minPwdLength,
					maxPwdLength,
					true,
				),
			),
		})
	if ret := infoChecker.Check(dp.sfwDownloadInfo); !ret.Result {
		return fmt.Errorf("check software download para failed: %s", ret.Reason)
	}
	dp.progress = progressPreparing
	return nil
}

func (dp *downloadProcess) cleanDownloadDir() {
	req := config.DirReq{
		Path:     constants.EdgeDownloadPath,
		ToDelete: true,
	}
	if err := sendDirReq(req); err != nil {
		hwlog.RunLog.Warnf("clean download dir failed: %v", err)
	}
}

func (dp *downloadProcess) prepareDownloadDir() error {
	req := config.DirReq{
		Path: constants.EdgeDownloadPath,
	}
	if err := sendDirReq(req); err != nil {
		return fmt.Errorf("prepare download dir failed, %v", err)
	}
	return nil
}

func sendDirReq(req config.DirReq) error {
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.DownloadManagerName,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptUpdate,
		Resource:              constants.InnerPrepareDir,
		Content:               req,
		TransferStructIntoStr: true,
	})
	if err != nil {
		return fmt.Errorf("create message error: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, defaultTimeOutInterval)
	if err != nil {
		return fmt.Errorf("send request failed, %v", err)
	}
	var respContent string
	if err = resp.ParseContent(&respContent); err != nil {
		return fmt.Errorf("get resp content failed: %v", err)
	}

	if respContent != "OK" {
		return fmt.Errorf("edge-om process error: %s", respContent)
	}
	return nil
}

func (dp *downloadProcess) checkDownloadDir() error {
	uid, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return fmt.Errorf("get user id faild, error: %v", err)
	}
	if _, err = fileutils.CheckOwnerAndPermission(constants.EdgeDownloadPath, constants.ModeUmask077, uid); err != nil {
		return fmt.Errorf("check download dir failed, check owner and permission error: %v", err)
	}
	if _, err = fileutils.CheckOriginPath(constants.EdgeDownloadPath); err != nil {
		return fmt.Errorf("check download dir failed, %v", err)
	}
	return nil
}

func (dp *downloadProcess) getSoftwareCert() error {
	req := config.CertReq{
		CertName: constants.SoftwareCertName,
	}
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.DownloadManagerName,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptGet,
		Resource:              constants.InnerCert,
		Content:               req,
		TransferStructIntoStr: true,
	})
	if err != nil {
		return fmt.Errorf("get software cert failed, create message error: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, defaultTimeOutInterval)
	if err != nil {
		return fmt.Errorf("send request for software cert failed, %v", err)
	}
	var certResp config.CertResp
	if err = resp.ParseContent(&certResp); err != nil {
		return fmt.Errorf("convert para failed: %v", err)
	}
	if certResp.ErrorMsg != "" {
		return fmt.Errorf("get software cert failed, edge-om process error: %s", certResp.ErrorMsg)
	}
	dp.cert = certResp.CertContent
	dp.crlContent = certResp.CrlContent
	dp.progress = progressDownloading

	return nil
}

func (dp *downloadProcess) verifyAndUnpack() error {
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:      constants.DownloadManagerName,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptPost,
		Resource:    constants.InnerSoftwareVerification,
		Content:     nil,
	})
	if err != nil {
		return fmt.Errorf("verify downloaded files failed, create message error: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, defaultTimeOutInterval)
	if err != nil {
		return fmt.Errorf("send verify request failed, %v", err)
	}
	var respContent string
	if err = resp.ParseContent(&respContent); err != nil {
		return fmt.Errorf("get resp content failed: %v", err)
	}

	if respContent != "OK" {
		return fmt.Errorf("verify and unpack downloaded files failed, edge-om process error: %s", respContent)
	}
	dp.progress = progressSuccess
	return nil
}

func reportDownloadProcess(progress uint64, err error) {
	hwlog.RunLog.Info("report software download progress to cloud")
	info := config.ProgressInfo{
		Progress: progress,
	}
	if err != nil {
		info.Res = constants.Failed
		info.Msg = err.Error()
	} else {
		info.Res = constants.Success
	}

	var edgeReport = edgeReportUpgradeResInfo{
		SerialNumber: configpara.GetInstallerConfig().SerialNumber,
		ProgressInfo: info,
	}
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.DownloadManagerName,
		Destination:           constants.ModEdgeHub,
		Operation:             constants.OptReport,
		Resource:              constants.ResDownloadProgress,
		Content:               edgeReport,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report software download progress to cloud failed, create message error: %v", err)
		return
	}
	if err = modulemgr.SendMessage(msg); err != nil {
		hwlog.RunLog.Errorf("report software download progress to cloud failed, send msg error: %v", err)
	}
}
