// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows this file for effect flow
package flows

import (
	"errors"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
	commonTasks "edge-installer/pkg/installer/common/tasks"
	"edge-installer/pkg/installer/upgrade/tasks"

	"huawei.com/mindx/common/hwlog"
)

type effectFlow struct {
	pathMgr *pathmgr.PathManager
}

// NewEffectFlow create effect flow instance
func NewEffectFlow(pathMgr *pathmgr.PathManager) common.Flow {
	return &effectFlow{
		pathMgr: pathMgr,
	}
}

// RunTasks run upgrade tasks
func (ef *effectFlow) RunTasks() error {
	hwlog.RunLog.Info("------------------process upgrade task success------------------")
	postProcess := tasks.PostEffectProcessTask{
		PostProcessBaseTask: commonTasks.PostProcessBaseTask{
			WorkPathMgr: ef.pathMgr.WorkPathMgr,
			LogPathMgr:  ef.pathMgr.LogPathMgr,
		},
		ConfigPathMgr: ef.pathMgr.ConfigPathMgr,
	}
	if err := postProcess.Run(); err != nil {
		return errors.New("upgrade post process task failed")
	}
	return nil
}
