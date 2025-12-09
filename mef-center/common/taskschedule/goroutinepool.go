// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package taskschedule
package taskschedule

import (
	"context"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
)

type goroutinePoolController struct {
	GoroutinePool
	ctx                     context.Context
	executorFactoryRegistry *sync.Map
	waitingQueue            chan *taskContextImpl
	startOnce               sync.Once
}

func (p *goroutinePoolController) start() {
	p.startOnce.Do(func() {
		p.waitingQueue = make(chan *taskContextImpl, p.Spec.MaxCapacity)
		for i := uint(0); i < p.Spec.MaxConcurrency; i++ {
			go p.runWorker()
		}
	})
}

func (p *goroutinePoolController) runWorker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.waitingQueue:
			if err := p.handleTask(task); err != nil {
				hwlog.RunLog.Errorf("(taskId=%s)failed to run task, %v", task.Spec().Id, err)
			}
		}
	}
}

func (p *goroutinePoolController) handleTask(t *taskContextImpl) error {
	// entering processing phase
	status := TaskStatus{Phase: Processing, StartedAt: time.Now()}
	if err := t.updateStatus(status, false); err != nil {
		if err == ErrNoRowsAffected {
			return nil
		}
		return err
	}

	executorFactory, err := p.getExecutorFactory(t.Spec().Command)
	if err != nil {
		if updateErr := t.updateStatus(status, false); updateErr != nil {
			hwlog.RunLog.Errorf("(taskId=%s)failed to update task, %v", t.Spec().Id, updateErr)
		}
		return err
	}
	// make executor context and launch a new go routine to run the task
	if executorFactory != nil {
		executor := executorFactory.CreateExecutor()
		if executor != nil {
			go runTask(t, executor)
		}
	}

	<-t.Done()
	return nil
}

func (p *goroutinePoolController) getExecutorFactory(factoryId string) (TaskExecutorFactory, error) {
	if factoryId == "" {
		return nil, nil
	}
	value, ok := p.executorFactoryRegistry.Load(factoryId)
	if !ok {
		return nil, ErrFactoryNotFound
	}
	factory, ok := value.(TaskExecutorFactory)
	if !ok {
		return nil, ErrTypeInvalid
	}
	return factory, nil
}

func runTask(t TaskContext, executor TaskExecutor) {
	defer func() {
		if data := recover(); data != nil {
			hwlog.RunLog.Errorf("(taskId=%s)crash, %v", t.Spec().Id, data)
		}
	}()
	executor(t)
}
