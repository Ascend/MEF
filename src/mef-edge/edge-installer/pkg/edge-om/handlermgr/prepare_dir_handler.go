// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

// prepareDirHandler prepare work directory for file download.
// inner message handler, so do not need to record operation logs.
type prepareDirHandler struct{}

func (p *prepareDirHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle prepare directory message form edge-main")
	var req config.DirReq
	if err := msg.ParseContent(&req); err != nil {
		p.sendResponse(msg, "parse request parameter failed")
		hwlog.RunLog.Errorf("parse request parameter failed, error: %v", err)
		return errors.New("parse request parameter failed")
	}

	var processErr error
	if req.ToDelete {
		processErr = processDeleteDirRequest(req)
	} else {
		processErr = processCreateDirRequest(req)
	}
	if processErr != nil {
		p.sendResponse(msg, processErr.Error())
		hwlog.RunLog.Errorf("prepare directory for software failed, error: %v", processErr)
		return errors.New("prepare directory for software failed")
	}

	p.sendResponse(msg, "OK")
	return nil
}

func (p *prepareDirHandler) sendResponse(msg *model.Message, respMsg string) {
	newResp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("new response message failed, error: %v", err)
		return
	}
	if err = newResp.FillContent(respMsg); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return
	}
	if err = sendHandlerReplyMsg(newResp); err != nil {
		hwlog.RunLog.Errorf("send prepare directory message response failed, error: %v", err)
	}
}

func processCreateDirRequest(req config.DirReq) error {
	hwlog.RunLog.Info("start to create directory for software download")
	if fileutils.IsExist(req.Path) {
		if err := fileutils.DeleteAllFileWithConfusion(req.Path); err != nil {
			hwlog.RunLog.Errorf("delete path %s failed: %v", req.Path, err)
			return err
		}
	}

	linkChecker := fileutils.NewFileLinkChecker(false)
	ownerChecker := fileutils.NewFileOwnerChecker(true, false, fileutils.RootUid, fileutils.RootGid)
	modeChecker := fileutils.NewFileModeChecker(true, fileutils.DefaultWriteFileMode, false, false)

	linkChecker.SetNext(ownerChecker)
	linkChecker.SetNext(modeChecker)
	// make sure /home/data dir is in correct permission
	if err := fileutils.SetPathPermission(
		filepath.Dir(req.Path), constants.Mode755, false, true, linkChecker); err != nil {
		hwlog.RunLog.Errorf("set path %s's permission failed: %v", filepath.Dir(req.Path), err)
		return fmt.Errorf("failed to set permission of root dir %s, %v", filepath.Dir(req.Path), err)
	}

	if err := fileutils.CreateDir(req.Path, constants.Mode700, linkChecker); err != nil {
		hwlog.RunLog.Errorf("create dir %s failed: %v", req.Path, err)
		return fmt.Errorf("failed to create dir %s, %v", req.Path, err)
	}
	if err := util.SetPathOwnerGroupToMEFEdge(req.Path, true, true); err != nil {
		hwlog.RunLog.Errorf("set path %s's owner to MEFEdge failed: %v", req.Path, err)
		return err
	}
	hwlog.RunLog.Info("successfully prepare directory for software download")
	return nil
}

func processDeleteDirRequest(req config.DirReq) error {
	hwlog.RunLog.Info("start to clean directory of software download")
	if err := fileutils.DeleteAllFileWithConfusion(req.Path); err != nil {
		hwlog.RunLog.Errorf("delete path %s failed: %v", req.Path, err)
		return fmt.Errorf("delete directory [%s] failed, %v", req.Path, err)
	}
	hwlog.RunLog.Info("successfully clean directory of software download")
	return nil
}
