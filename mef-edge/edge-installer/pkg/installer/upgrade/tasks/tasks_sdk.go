// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_A500

// Package tasks for some methods that are performed on the non a500 device
package tasks

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
)

func (swp *SetWorkPathTask) prepareCfgBackupDir() error {
	// !MEFEdge_A500 do not prepare config backup dir
	return nil
}

func (p *PostEffectProcessTask) copyResetScriptToP7() error {
	// !MEFEdge_A500 do not copy reset script to p7
	return nil
}

func (p *PostEffectProcessTask) smoothConfig() error {
	if err := p.smoothCommonConfig(); err != nil {
		return err
	}
	if err := config.SmoothAlarmConfigDB(); err != nil {
		hwlog.RunLog.Errorf("smooth edge_om alarm config to db failed, error: %v", err)
		return errors.New("smooth edge_om alarm config to db failed")
	}
	return nil
}

func (p *PostEffectProcessTask) refreshDefaultCfgDir() error {
	// !MEFEdge_A500 do not need refresh default config backup dir
	return nil
}
