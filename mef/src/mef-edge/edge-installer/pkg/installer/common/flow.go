// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common this file for flow interface
package common

import "huawei.com/mindx/common/hwlog"

// ProgressSuccess success progress
const ProgressSuccess = 100

// Flow process flow
type Flow interface {
	RunTasks() error
}

// PostFunc post-processing func
type PostFunc func(*FlowItem) error

// FinalFunc final processing func
type FinalFunc func()

// ExceptionFunc exception func
type ExceptionFunc func()

// FlowBase process flow base
type FlowBase struct {
	items       []*FlowItem
	posts       []PostFunc
	finals      []*finalItem
	exceptions  []ExceptionFunc
	CurProgress uint64
}

// Task process task
type Task interface {
	Run() error
}

// FlowItem process flow item
type FlowItem struct {
	Description string
	Progress    uint64
	Task        Task
	Error       error
}

type finalItem struct {
	method          FinalFunc
	progressReached uint64
}

// RunTasks run flow
func (f *FlowBase) RunTasks() error {
	defer func() {
		for _, final := range f.finals {
			if final != nil && f.CurProgress >= final.progressReached {
				final.method()
			}
		}
	}()
	for _, item := range f.items {
		if item != nil && item.Task != nil {
			if err := item.Task.Run(); err != nil {
				hwlog.RunLog.Errorf("task[%s] failed,error:%v", item.Description, err)
				item.Error = err
			}
		}
		for _, post := range f.posts {
			if post == nil {
				continue
			}
			if err := post(item); err != nil {
				hwlog.RunLog.Errorf("task[%s] post-processing failed,error:%v", item.Description, err)
				return err
			}
		}
		if item.Error != nil {
			for _, exceptionFunc := range f.exceptions {
				exceptionFunc()
			}
			return item.Error
		}
		f.CurProgress = item.Progress
	}
	return nil
}

// AddTask add flow task
func (f *FlowBase) AddTask(task Task, description string, progress uint64) {
	if f.items == nil {
		f.items = make([]*FlowItem, 0)
	}
	f.items = append(f.items, &FlowItem{
		Description: description,
		Progress:    progress,
		Task:        task,
	})
}

// AddPost add post-processing func
func (f *FlowBase) AddPost(post PostFunc) {
	if f.posts == nil {
		f.posts = make([]PostFunc, 0)
	}
	f.posts = append(f.posts, post)
}

// AddFinal add final processing func
func (f *FlowBase) AddFinal(final FinalFunc, progressReached uint64) {
	if f.finals == nil {
		f.finals = make([]*finalItem, 0)
	}
	f.finals = append(f.finals, &finalItem{method: final, progressReached: progressReached})
}

// AddException add exception func
func (f *FlowBase) AddException(exception ExceptionFunc) {
	if f.exceptions == nil {
		f.exceptions = make([]ExceptionFunc, 0)
	}
	f.exceptions = append(f.exceptions, exception)
}
