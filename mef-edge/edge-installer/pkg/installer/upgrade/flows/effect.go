// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
