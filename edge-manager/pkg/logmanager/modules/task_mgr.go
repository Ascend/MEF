// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package modules enables collecting logs
package modules

import (
	"context"
	"errors"
	"sync"
	"time"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
)

const (
	clearTaskInterval     = 60 * time.Second
	maxTaskExecuteCycles  = 30
	maxTaskInactiveCycles = 5
	maxTaskQueryCycles    = 300
)

// TaskMgr manages tasks
type TaskMgr interface {
	// Start starts task manager
	Start()
	// NotifyProgress notifies progress of task
	NotifyProgress(progress logcollect.TaskProgress, nodeSn string) error
	// AddTask adds task
	AddTask(nodeSn, fileBaseName string) error
	// GetTaskProgress get task progress
	GetTaskProgress(nodeSn string) (logcollect.TaskProgress, error)
	// GetTaskPath get output path of task
	GetTaskPath(nodeSn string) (string, error)
}

type edgeTaskInfo struct {
	currentProgress logcollect.TaskProgress
	fileBaseName    string
	executeCycles   int
	inactiveCycles  int
}

type taskMgr struct {
	ctx   context.Context
	lock  sync.RWMutex
	tasks map[string]edgeTaskInfo
}

// NewTaskMgr creates task manager
func NewTaskMgr(ctx context.Context) TaskMgr {
	return &taskMgr{
		ctx:   ctx,
		tasks: make(map[string]edgeTaskInfo),
	}
}

func (p *taskMgr) Start() {
	go p.autoCleanLoop()
}

func (p *taskMgr) autoCleanLoop() {
	timer := time.NewTimer(clearTaskInterval)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-timer.C:
		}
		timer.Reset(clearTaskInterval)
		p.autoClean()
	}
}

func (p *taskMgr) autoClean() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.tasks == nil {
		return
	}
	for taskId := range p.tasks {
		task := p.tasks[taskId]
		task.executeCycles++
		task.inactiveCycles++

		if task.executeCycles >= maxTaskExecuteCycles || task.inactiveCycles >= maxTaskInactiveCycles {
			if task.isRunning() {
				task.currentProgress.Status = common.ErrorLogCollectEdgeBusiness
				task.currentProgress.Message = "task timeout"
			}
		}
		p.tasks[taskId] = task

		if task.executeCycles >= maxTaskQueryCycles {
			delete(p.tasks, taskId)
		}
	}
}

func (p *taskMgr) NotifyProgress(progress logcollect.TaskProgress, nodeSn string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	task, ok := p.tasks[nodeSn]
	if !ok {
		return errors.New("no such task")
	}
	if !task.isRunning() {
		return errors.New("task is not running")
	}

	task.currentProgress = progress
	task.inactiveCycles = 0

	p.tasks[nodeSn] = task
	return nil
}

func (p *taskMgr) AddTask(nodeSn, fileBaseName string) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	task, ok := p.tasks[nodeSn]
	if ok && task.isRunning() {
		return errors.New("task is running")
	}

	p.tasks[nodeSn] = edgeTaskInfo{
		currentProgress: logcollect.TaskProgress{Status: common.Success},
		fileBaseName:    fileBaseName,
	}
	return nil
}

func (p *taskMgr) GetTaskProgress(nodeSn string) (logcollect.TaskProgress, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	task, ok := p.tasks[nodeSn]
	if !ok {
		return logcollect.TaskProgress{}, errors.New("task not found")
	}
	return task.currentProgress, nil

}

func (p *taskMgr) GetTaskPath(nodeSn string) (string, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	task, ok := p.tasks[nodeSn]
	if !ok {
		return "", errors.New("task not found")
	}
	return task.fileBaseName, nil

}

func (t *edgeTaskInfo) isRunning() bool {
	return t.currentProgress.Status == common.Success && t.currentProgress.Progress != logcollect.ProgressMax
}
