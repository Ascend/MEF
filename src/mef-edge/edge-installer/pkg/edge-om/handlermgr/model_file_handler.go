// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

const maxFileCount = 2000

type modelFileHandler struct {
}

func (m *modelFileHandler) Handle(msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("msg is nil")
	}

	var info types.ModelFileInfo
	if err := msg.ParseContent(&info); err != nil {
		hwlog.RunLog.Errorf("parse parameters failed: %v", err)
		return errors.New("parse parameters failed")
	}

	m.deleteModelFile(info)
	return nil
}

func (m *modelFileHandler) deleteTmpModelFile(info types.ModelFileInfo) {
	_, err := fileutils.RealDirCheck(constants.ModelFileRootPath, true, false)
	if err != nil {
		hwlog.RunLog.Errorf("delete tmp model file failed, because check dir [%s] failed:%v",
			constants.ModelFileRootPath, err)
		return
	}
	for _, fileInfo := range info.ModelFiles {
		fileDir := filepath.Join(constants.ModeFileDownloadDir, info.Uuid, fileInfo.Name)
		if err = fileutils.DeleteAllFileWithConfusion(fileDir); err != nil {
			hwlog.RunLog.Warnf("delete model file [%s] failed: %v", fileDir, err)
		}
	}
	hwlog.RunLog.Info("delete tmp model file successfully")
}

func (m *modelFileHandler) deletePodAllModelFile(info types.ModelFileInfo) {
	_, err := fileutils.RealDirCheck(constants.ModelFileRootPath, true, false)
	if err != nil {
		hwlog.RunLog.Errorf("delete pod all model file failed, because check dir [%s] failed:%v",
			constants.ModelFileRootPath, err)
		return
	}
	fileDir := filepath.Join(constants.ModeFileActiveDir, info.Uuid)
	if err = fileutils.DeleteAllFileWithConfusion(fileDir); err != nil {
		hwlog.RunLog.Warnf("delete model file [%s] failed: %v", fileDir, err)
	}

	fileDir = filepath.Join(constants.ModeFileDownloadDir, info.Uuid)
	if err = fileutils.DeleteAllFileWithConfusion(fileDir); err != nil {
		hwlog.RunLog.Warnf("delete model file [%s] failed: %v", fileDir, err)
	}

	hwlog.RunLog.Info("delete all model file successfully")
}

func (m *modelFileHandler) deletePodPartModelFile(info types.ModelFileInfo) {
	_, err := fileutils.RealDirCheck(constants.ModelFileRootPath, true, false)
	if err != nil {
		hwlog.RunLog.Errorf("delete part model file failed, because check dir [%s] failed:%v",
			constants.ModelFileRootPath, err)
		return
	}

	for _, fileInfo := range info.ModelFiles {
		fileDir := filepath.Join(constants.ModeFileActiveDir, info.Uuid, fileInfo.Name)
		if err = fileutils.DeleteAllFileWithConfusion(fileDir); err != nil {
			hwlog.RunLog.Warnf("delete model file [%s] failed: %v", fileDir, err)
		}

		fileDir = filepath.Join(constants.ModeFileDownloadDir, info.Uuid, fileInfo.Name)
		if err = fileutils.DeleteAllFileWithConfusion(fileDir); err != nil {
			hwlog.RunLog.Warnf("delete model file [%s] failed: %v", fileDir, err)
		}
	}

	hwlog.RunLog.Info("delete part model file successfully")
}

func (m *modelFileHandler) deleteModelFile(info types.ModelFileInfo) {
	hwlog.RunLog.Info("start to delete model file")
	switch info.Target {
	case constants.TargetTypeTemp:
		m.deleteTmpModelFile(info)
		break
	case constants.TargetTypeAll:
		m.deletePodAllModelFile(info)
		break
	default:
		m.deletePodPartModelFile(info)
		break
	}
}
