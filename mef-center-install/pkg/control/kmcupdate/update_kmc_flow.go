// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package kmcupdate this file for update kmc flow
package kmcupdate

import (
	"errors"
	"fmt"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// UpdateKmcFlow is the struct for update kmc flow
type UpdateKmcFlow struct {
	pathMgr *util.InstallDirPathMgr
}

// NewUpdateKmcFlow create UpdateKmcTask instance
func NewUpdateKmcFlow(pathMgr *util.InstallDirPathMgr) *UpdateKmcFlow {
	return &UpdateKmcFlow{pathMgr: pathMgr}
}

func (muk *UpdateKmcFlow) getModules() []string {
	return []string{util.CertManagerName, util.NginxManagerName, util.EdgeManagerName, util.MefCenterRootName}
}

// RunFlow is the main func to start a task
func (muk *UpdateKmcFlow) RunFlow() error {
	var failedModule []string
	for _, module := range muk.getModules() {
		task := NewManualUpdateKmcTask(muk.pathMgr.ConfigPathMgr, module)

		if err := task.RunTask(); err != nil {
			failedModule = append(failedModule, module)
		}
	}

	if len(failedModule) == 0 {
		return nil
	}
	fmt.Printf("update module %s's kmc key failed\n", failedModule)

	return errors.New("update kmc key failed")
}
