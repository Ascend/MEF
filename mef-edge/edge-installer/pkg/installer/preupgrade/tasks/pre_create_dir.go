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

// Package tasks for flow prepare upgrade task
package tasks

import (
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
)

type prepareDir struct {
	softwareName string
}

// PrepareDir prepare upgrade dir
func PrepareDir(softwareName string) common.Task {
	return &prepareDir{softwareName: softwareName}
}

func checkAndCreatPackageDir(dirName string) error {
	if fileutils.IsExist(dirName) {
		if err := fileutils.DeleteAllFileWithConfusion(dirName); err != nil {
			return err
		}
	}
	if err := fileutils.CreateDir(dirName, fileutils.Mode700); err != nil {
		hwlog.RunLog.Errorf("create package directory failed, error: %v", err)
		return err
	}

	return nil
}

// Run task
func (p *prepareDir) Run() error {
	if err := checkAndCreatPackageDir(filepath.Join(constants.PkgPath, p.softwareName)); err != nil {
		hwlog.RunLog.Errorf("create package path failed: %v", err)
		return err
	}
	if err := checkAndCreatPackageDir(filepath.Join(constants.UnpackPath, p.softwareName)); err != nil {
		hwlog.RunLog.Errorf("create un package path failed: %v", err)
		return err
	}

	if err := checkAndCreatPackageDir(filepath.Join(constants.UnpackPath, p.softwareName,
		constants.UnpackZipDir)); err != nil {
		hwlog.RunLog.Errorf("create un package path failed: %v", err)
		return err
	}

	hwlog.RunLog.Info("prepare dir success")
	return nil
}
