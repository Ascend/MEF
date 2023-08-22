// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func newTaskRepo(tx *gorm.DB) taskRepository {
	return taskRepository{DB: tx}
}

type taskRepository struct {
	*gorm.DB
}

func (r taskRepository) createTask(task Task) error {
	stmt := r.DB.Model(Task{}).Create(&task)
	if stmt.Error != nil {
		return errors.New("failed to create task")
	}
	return nil
}

func (r taskRepository) deleteTask(id string) error {
	if id == "" {
		return errors.New("id is empty")
	}
	stmt := r.DB.Model(Task{}).Delete(Task{Spec: TaskSpec{Id: id}})
	if stmt.Error != nil {
		return errors.New("failed to updates task updates")
	}
	return nil
}

func (r taskRepository) updateTaskStatus(id string, status TaskStatus) (Task, int64, error) {
	if id == "" {
		return Task{}, 0, errors.New("id is empty")
	}
	var task Task
	stmt := r.DB.Model(&task).Clauses(clause.Returning{}).Where("id = ?", id).Updates(Task{Status: status})
	if stmt.Error != nil {
		return Task{}, 0, errors.New("failed to updates task status")
	}
	return task, stmt.RowsAffected, nil
}

func (r taskRepository) updateUnfinishedTasksToFailed() error {
	return r.DB.Model(Task{}).
		Where("phase IN (?)", []TaskPhase{Waiting, Progressing, Aborting}).
		Updates(Task{Status: TaskStatus{Phase: Failed}}).Error
}

func (r taskRepository) getTask(id string) (Task, error) {
	if id == "" {
		return Task{}, errors.New("id is empty")
	}
	var tasks Task
	stmt := r.DB.Model(Task{}).First(&tasks, Task{Spec: TaskSpec{Id: id}})
	if stmt.Error != nil {
		return Task{}, errors.New("failed to find task")
	}
	return tasks, nil
}

func (r taskRepository) getSubTasks(parentId string) ([]Task, error) {
	if parentId == "" {
		return nil, errors.New("id is empty")
	}
	var tasks []Task
	stmt := r.DB.Model(Task{}).Find(&tasks, Task{Spec: TaskSpec{ParentId: parentId}})
	if stmt.Error != nil {
		return nil, errors.New("failed to find subtasks")
	}
	return tasks, nil
}

func (r taskRepository) getFinishedMasterTasks() ([]Task, error) {
	var (
		tasks        []Task
		finishPhases = []TaskPhase{Failed, Succeed, PartiallyFailed}
	)
	stmt := r.DB.Model(Task{}).Where(`parent_id = "" AND phase IN (?)`, finishPhases).Find(&tasks)
	if stmt.Error != nil {
		return nil, errors.New("failed to find finished master tasks")
	}
	return tasks, nil
}

func (r taskRepository) getTaskTree(taskId string) (TaskTreeNode, error) {
	task, err := r.getTask(taskId)
	if err != nil {
		return TaskTreeNode{}, err
	}
	return r.getTaskTreeInternal(&task)
}

func (r taskRepository) getTaskTreeInternal(task *Task) (TaskTreeNode, error) {
	subTasks, err := r.getSubTasks(task.Spec.Id)
	if err != nil {
		return TaskTreeNode{}, err
	}
	var children []TaskTreeNode
	for index := range subTasks {
		childTask := subTasks[index]
		childNode, err := r.getTaskTreeInternal(&childTask)
		if err != nil {
			return TaskTreeNode{}, err
		}
		children = append(children, childNode)
	}
	return TaskTreeNode{Current: task, Children: children}, nil
}
