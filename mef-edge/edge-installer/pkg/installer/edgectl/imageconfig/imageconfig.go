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

// Package imageconfig for edge control command image mapping config
package imageconfig

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/util"
)

// ImageCfgFlow image mapping config flow
type ImageCfgFlow struct {
	imageAddress string
}

type importImageCfgTask struct {
	imageAddress string
}

// NewImageCfgFlow create image mapping config flow instance
func NewImageCfgFlow(imageAddress string) *ImageCfgFlow {
	return &ImageCfgFlow{
		imageAddress: imageAddress,
	}
}

// RunTasks run image config task
func (icf ImageCfgFlow) RunTasks() error {
	importTask := importImageCfgTask{imageAddress: icf.imageAddress}
	if err := importTask.runTask(); err != nil {
		return errors.New("import image config failed")
	}

	return nil
}

func (ict *importImageCfgTask) runTask() error {
	return ict.importImageConfig()
}

func (ict *importImageCfgTask) importImageConfig() error {
	cfg, err := config.GetImageCfg()
	if err == gorm.ErrRecordNotFound {
		return ict.createImageCfg()
	}
	if err != nil {
		hwlog.RunLog.Errorf("get image config failed, error:%v", err)
		return err
	}
	return ict.updateImageCfg(cfg.ImageAddress)
}

func (ict *importImageCfgTask) updateImageCfg(oldAddress string) error {
	if oldAddress == ict.imageAddress {
		hwlog.RunLog.Warnf("image config %s is same, and no need to clear", oldAddress)
		return nil
	}
	if err := util.DeleteImageCertFile(oldAddress); err != nil {
		hwlog.RunLog.Errorf("delete image config %s failed, error:%v", oldAddress, err)
		return err
	}
	return ict.createImageCfg()
}

func (ict *importImageCfgTask) createImageCfg() error {
	imageConfig := &config.ImageConfig{
		ImageAddress: ict.imageAddress,
	}
	if err := config.SetImageCfg(imageConfig); err != nil {
		hwlog.RunLog.Errorf("set image config to db error, %s", err.Error())
		return fmt.Errorf("set image config to db error, %s", err.Error())
	}
	return nil
}
