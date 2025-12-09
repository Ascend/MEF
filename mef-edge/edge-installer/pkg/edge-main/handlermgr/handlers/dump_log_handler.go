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

// Package handlers
package handlers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/cloudcert"
	"edge-installer/pkg/edge-main/common/configpara"
)

// TaskErrorInfo defines task error info
type TaskErrorInfo struct {
	Id      string `json:"id"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// DumpLogReq defines a dump log request
type DumpLogReq struct {
	Module string `json:"module"`
	TaskId string `json:"taskId"`
}

const (
	packLogTimeout            = 30 * time.Second
	uploadLogTimeout          = 10 * time.Minute
	dumpSingleNodeLogTaskName = `dumpSingleNodeLog`
	regexpTaskIdStr           = `[-_a-zA-Z0-9.]{1,128}`
	regexpSingleNodeTaskIdStr = "^" + dumpSingleNodeLogTaskName + regexpTaskIdStr + "$"
)

var (
	dumpLogHandlerInst dumpLogHandler
	regexpTaskId       = regexp.MustCompile(regexpSingleNodeTaskIdStr)
	localFilePath      = filepath.Join(
		constants.LogCollectTempDir, constants.EdgeMain, constants.LogCollectTempFileName)
)

func getDumpLogHandler() *dumpLogHandler {
	return &dumpLogHandlerInst
}

type dumpLogHandler struct {
	resultCh chan string
	running  int32
}

func (h *dumpLogHandler) Handle(msg *model.Message) error {
	taskId, err := parseAndCheckArgs(msg)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse or check args, %v", err)
		return fmt.Errorf("failed to parse or check args, %v", err)
	}

	process := dumpLogProcess{
		handler: h,
		taskId:  taskId,
	}
	swapped := atomic.CompareAndSwapInt32(&h.running, 0, 1)
	if !swapped {
		hwlog.RunLog.Error("dump log handler busy")
		busyErr := errors.New("dump log handler busy")
		process.feedbackError(busyErr)
		return busyErr
	}

	go func() {
		defer atomic.StoreInt32(&h.running, 0)
		if err := process.process(); err != nil {
			process.feedbackError(err)
		}
	}()
	return nil
}

type dumpLogProcess struct {
	handler     *dumpLogHandler
	taskId      string
	tlsCertInfo *certutils.TlsCertInfo
	netConfig   config.NetManager
}

func (p *dumpLogProcess) process() error {
	p.netConfig = configpara.GetNetConfig()
	if p.netConfig.IP == "" {
		hwlog.RunLog.Error("failed to get net config, ip is invalid")
		return errors.New("failed to get net config, ip is invalid")
	}

	hwlog.RunLog.Info("start to dump log")
	hwlog.OpLog.Infof("[%s@%s] %s %s", p.netConfig.NetType, p.netConfig.IP,
		constants.OptPost, constants.ResDumpLogTask)
	if err := p.doProcess(); err != nil {
		hwlog.RunLog.Errorf("dump log failed, %v", err)
		hwlog.OpLog.Errorf("[%s@%s] %s %s failed",
			p.netConfig.NetType, p.netConfig.IP, constants.OptPost, constants.ResDumpLogTask)
		return err
	}
	hwlog.RunLog.Info("dump log success")
	hwlog.OpLog.Infof("[%s@%s] %s %s success",
		p.netConfig.NetType, p.netConfig.IP, constants.OptPost, constants.ResDumpLogTask)
	return nil
}

func (p *dumpLogProcess) doProcess() error {
	if err := p.packLogs(); err != nil {
		hwlog.RunLog.Errorf("failed to pack logs, %v", err)
		return errors.New("failed to pack logs")
	}

	tlsCertInfo, err := cloudcert.GetEdgeHubCertInfo()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get tls cert info, %v", err)
		return errors.New("failed to get tls cert info")
	}
	p.tlsCertInfo = tlsCertInfo
	hwlog.RunLog.Info("get tls config successful")

	if err := p.uploadLogs(); err != nil {
		hwlog.RunLog.Errorf("failed to upload logs, %v", err)
		return errors.New("failed to upload logs")
	}
	return nil
}

func (p *dumpLogProcess) packLogs() error {
	request, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("failed to create request message, %v", err)
	}
	request.SetRouter(constants.ModHandlerMgr, constants.ModEdgeOm, constants.OptPost, constants.ResPackLogRequest)
	if err := modulemgr.SendMessage(request); err != nil {
		return fmt.Errorf("failed to send request, %v", err)
	}

	p.handler.resultCh = make(chan string)
	defer close(p.handler.resultCh)
	timer := time.NewTimer(packLogTimeout)
	defer timer.Stop()
	select {
	case status, ok := <-p.handler.resultCh:
		if !ok {
			return errors.New("channel is closed")
		}
		if status != constants.OK {
			return fmt.Errorf("got an unsuccessful response %s", status)
		}
	case <-timer.C:
		return errors.New("timeout")
	}
	hwlog.RunLog.Info("pack log successful")
	return nil
}

func (p *dumpLogProcess) uploadLogs() error {
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file, %v", err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			hwlog.RunLog.Errorf("failed to close temp file, %v", err)
		}
	}()
	fileChecker := fileutils.NewFileLinkChecker(false)
	if err := fileChecker.Check(localFile, localFilePath); err != nil {
		return fmt.Errorf("failed to check temp file, %v", err)
	}

	localFileStat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get local file size, %v", err)
	}

	sha256Inst := sha256.New()
	if _, err := io.Copy(sha256Inst, localFile); err != nil {
		return fmt.Errorf("failed to calculate checksum, %v", err)
	}

	hwlog.RunLog.Info("start to upload file")
	url := fmt.Sprintf("https://%s:%d/logmgmt/dump/upload", p.netConfig.IP, p.netConfig.Port)
	headers := map[string]interface{}{
		"Task-Id":         p.taskId,
		"Package-Size":    localFileStat.Size(),
		"Sha256-Checksum": fmt.Sprintf("%x", sha256Inst.Sum(nil)),
	}
	respBytes, err := httpsmgr.GetHttpsReq(url, *p.tlsCertInfo, headers).SetReadTimeout(uploadLogTimeout).
		PostFile(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	if string(respBytes) != constants.OK {
		return fmt.Errorf("unexpected response from center: %s", respBytes)
	}
	hwlog.RunLog.Info("upload logs successful")
	return nil
}

func (p *dumpLogProcess) feedbackError(dumpErr error) {
	hwlog.RunLog.Errorf("process log dumping failed, %v, start to feedback error", dumpErr.Error())
	resp, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("feedback failed, create message error, %v", err)
		return
	}
	resp.SetRouter(constants.ModHandlerMgr, constants.ModEdgeHub, constants.OptReport, constants.ResDumpLogTaskError)
	if err = resp.FillContent(TaskErrorInfo{Id: p.taskId, Message: dumpErr.Error()}); err != nil {
		hwlog.RunLog.Errorf("fill task err into content failed: %v", err)
		return
	}

	if err := modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("feedback result failed, %v", err)
		return
	}
	hwlog.RunLog.Info("feedback result success")
}

func parseAndCheckArgs(msg *model.Message) (string, error) {
	var req DumpLogReq
	if err := msg.ParseContent(&req); err != nil {
		return "", fmt.Errorf("parma convert error: %v", err)
	}
	if !(req.Module == "edgeNode" && regexpTaskId.MatchString(req.TaskId)) {
		return "", errors.New("invalid argument")
	}
	return req.TaskId, nil
}
