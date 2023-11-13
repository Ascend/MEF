// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
)

type schedulerImpl struct {
	SchedulerSpec

	ctx  context.Context
	repo taskRepository

	executorFactories  sync.Map
	goroutinePools     sync.Map
	activeTaskContexts sync.Map
	activeTaskCount    int64
	taskDone           chan struct{}
}

func (s *schedulerImpl) RegisterExecutorFactory(factory TaskExecutorFactory) bool {
	_, loaded := s.executorFactories.LoadOrStore(factory.GetID(), factory)
	return !loaded
}

func (s *schedulerImpl) RegisterGoroutinePool(poolSpec GoroutinePoolSpec) bool {
	pool := &goroutinePoolController{
		GoroutinePool:           GoroutinePool{Spec: poolSpec},
		ctx:                     s.ctx,
		executorFactoryRegistry: &s.executorFactories,
	}
	_, loaded := s.goroutinePools.LoadOrStore(poolSpec.Id, pool)
	if !loaded {
		pool.start()
	}
	return !loaded
}

func (s *schedulerImpl) SubmitTask(taskSpec *TaskSpec) error {
	if taskSpec == nil {
		return ErrNilPointer
	}

	var (
		parent *taskContextImpl
		err    error
	)
	if taskSpec.ParentId != "" {
		parent, err = s.getActiveTaskContext(taskSpec.ParentId)
		if err != nil {
			return err
		}
	}
	taskCtx, err := createTaskContext(taskSpec, parent, s.ctx, s.taskDone, s.repo)
	if err != nil {
		return err
	}

	if !atomicIncreaseInt64(&s.activeTaskCount, s.MaxActiveTasks) {
		return ErrTooManyTask
	}
	submitSpec, err := s.trySubmitTask(taskCtx)
	if err != nil {
		taskCtx.destroy()
		atomic.AddInt64(&s.activeTaskCount, -1)
		return err
	}
	*taskSpec = submitSpec
	taskCtx.startLifeCycleEventsMonitoring()
	return nil
}

func (s *schedulerImpl) GetTaskContext(taskId string) (TaskContext, error) {
	tc, err := s.getActiveTaskContext(taskId)
	if err == nil {
		return tc, nil
	}
	if err != nil && err != ErrTaskNotFound {
		return nil, err
	}
	return getHistoryTaskContext(s.repo, taskId)
}

func (s *schedulerImpl) NewSubTaskSelector(taskId string) SubTaskSelector {
	return &subTaskSelectorImpl{
		taskId:               taskId,
		repo:                 s.repo,
		scheduler:            s,
		allowedMaxTasksCount: s.AllowedMaxTasksInDb,
	}
}

func (s *schedulerImpl) getActiveTaskContext(taskId string) (*taskContextImpl, error) {
	value, ok := s.activeTaskContexts.Load(taskId)
	if !ok {
		return nil, ErrTaskNotFound
	}
	tc, ok := value.(*taskContextImpl)
	if !ok {
		return nil, ErrTypeInvalid
	}
	return tc, nil
}

func (s *schedulerImpl) trySubmitTask(taskCtx *taskContextImpl) (TaskSpec, error) {
	task := Task{
		Spec: taskCtx.Spec(),
		Status: TaskStatus{
			Phase:     Waiting,
			CreatedAt: time.Now(),
		},
	}
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		total, err := newTaskRepo(tx).countTask()
		if err != nil || total >= s.AllowedMaxTasksInDb {
			return errors.New("count task failed or has reached maximum number")
		}
		if err := newTaskRepo(tx).createTask(task); err != nil {
			return err
		}
		s.activeTaskContexts.Store(taskCtx.Spec().Id, taskCtx)

		if err := s.dispatchTaskToGoroutinePool(taskCtx); err != nil {
			s.activeTaskContexts.Delete(taskCtx.spec.Id)
			return err
		}
		return nil
	})
	return task.Spec, err
}

func (s *schedulerImpl) dispatchTaskToGoroutinePool(taskCtx *taskContextImpl) error {
	value, ok := s.goroutinePools.Load(taskCtx.spec.GoroutinePool)
	if !ok {
		return ErrGoroutinePoolNotFound
	}
	pool, ok := value.(*goroutinePoolController)
	if !ok {
		return ErrGoroutinePoolNotFound
	}
	select {
	case pool.waitingQueue <- taskCtx:
	default:
		return ErrFullQueue
	}
	return nil
}

func (s *schedulerImpl) start() error {
	if err := s.repo.updateUnfinishedTasksToFailed(); err != nil {
		return err
	}
	go s.removeHistoryTasks()
	return nil
}

func (s *schedulerImpl) removeHistoryTasks() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case _, _ = <-s.taskDone:
		}
		s.doRemoveHistoryTasks()
	}
}

func (s *schedulerImpl) doRemoveHistoryTasks() {
	tasks, err := s.repo.getFinishedMasterTasks()
	if err != nil {
		hwlog.RunLog.Error("failed to query tasks from database")
		return
	}

	deleteTaskInfo := func(tn TaskTreeNode) {
		_, loaded := s.activeTaskContexts.LoadAndDelete(tn.Current.Spec.Id)
		if loaded {
			atomic.AddInt64(&s.activeTaskCount, -1)
		}
	}
	for _, task := range tasks {
		if !task.Status.Phase.IsFinished() {
			continue
		}
		if _, err := s.getActiveTaskContext(task.Spec.Id); err != nil {
			continue
		}
		if err := s.walkTaskTree(task.Spec.Id, deleteTaskInfo); err != nil {
			hwlog.RunLog.Errorf("failed to delete active task context, %v", err)
		}
	}

	if int64(len(tasks)) <= s.MaxHistoryMasterTasks {
		return
	}

	less := func(i, j int) bool { return tasks[i].Status.FinishedAt.Before(tasks[j].Status.FinishedAt) }
	sort.Slice(tasks, less)

	var deletedCount int
	deleteTaskFromDb := func(tn TaskTreeNode) {
		if err := s.repo.deleteTask(tn.Current.Spec.Id); err != nil {
			hwlog.RunLog.Errorf("delete task failed, error: %v", err)
		} else {
			deletedCount++
		}
	}
	for i := 0; deletedCount < len(tasks)-int(s.MaxHistoryMasterTasks) && i < len(tasks); i++ {
		if err := s.walkTaskTree(tasks[i].Spec.Id, deleteTaskFromDb); err != nil {
			hwlog.RunLog.Errorf("failed to clean tasks from database, %v", err)
		}
	}
}

func (s *schedulerImpl) walkTaskTree(taskId string, fn func(node TaskTreeNode)) error {
	taskTree, err := s.repo.getTaskTree(taskId)
	if err != nil {
		return err
	}

	func(tn TaskTreeNode, fn func(tn TaskTreeNode)) {
		for _, child := range tn.Children {
			fn(child)
		}
		fn(tn)
	}(taskTree, fn)
	return nil
}

type subTaskSelectorImpl struct {
	taskId               string
	scheduler            Scheduler
	repo                 taskRepository
	finishedTaskIds      sync.Map
	allowedMaxTasksCount int
}

func (s *subTaskSelectorImpl) Select(cancel ...<-chan struct{}) (TaskContext, error) {
	taskTree, err := s.repo.getTaskTree(s.taskId)
	if err != nil {
		return nil, err
	}
	var childrenCtx []TaskContext
	for _, child := range taskTree.Children {
		if _, ok := s.finishedTaskIds.Load(child.Current.Spec.Id); ok {
			continue
		}
		childCtx, err := s.scheduler.GetTaskContext(child.Current.Spec.Id)
		if err != nil {
			return nil, err
		}
		childrenCtx = append(childrenCtx, childCtx)
	}

	var chosenChildCtx TaskContext
	for i := 0; i < s.allowedMaxTasksCount; i++ {
		if len(childrenCtx) == 0 {
			return nil, ErrNoRunningSubTask
		}

		selectCases := make([]reflect.SelectCase, 0, len(childrenCtx)+1)
		for _, subTask := range childrenCtx {
			selectCases = append(selectCases, reflect.SelectCase{
				Dir: reflect.SelectRecv, Chan: reflect.ValueOf(subTask.Done())})
		}
		for _, ch := range cancel {
			selectCases = append(selectCases, reflect.SelectCase{
				Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
		}

		chosen, _, _ := reflect.Select(selectCases)
		if chosen >= len(childrenCtx) || chosen < 0 {
			return nil, ErrCancelled
		}
		chosenChildCtx = childrenCtx[chosen]
		if _, loaded := s.finishedTaskIds.LoadOrStore(chosenChildCtx.Spec().Id, struct{}{}); !loaded {
			break
		}
		childrenCtx = append(childrenCtx[:chosen], childrenCtx[chosen+1:]...)
	}

	return chosenChildCtx, nil
}

func startScheduler(ctx context.Context, db *gorm.DB, spec SchedulerSpec) (Scheduler, error) {
	if err := db.AutoMigrate(&Task{}); err != nil {
		return nil, fmt.Errorf("init task table failed, %v", err)
	}
	scheduler := &schedulerImpl{
		SchedulerSpec: spec,
		ctx:           ctx,
		repo:          newTaskRepo(db),
		taskDone:      make(chan struct{}, 1),
	}
	if err := scheduler.start(); err != nil {
		return nil, fmt.Errorf("start sheduler failed, %v", err)
	}
	return scheduler, nil
}

func newRandomID() (string, error) {
	const idLen = 12
	randomData := make([]byte, idLen)
	readLen, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}
	if readLen != idLen {
		return "", errors.New("no enough random")
	}
	return fmt.Sprintf("%x", randomData), nil
}

func atomicIncreaseInt64(int64Ptr *int64, max int64) bool {
	const increaseRetry = 10
	for i := 0; i < increaseRetry; i++ {
		value := atomic.LoadInt64(int64Ptr)
		if value >= max {
			break
		}
		if atomic.CompareAndSwapInt64(int64Ptr, value, value+1) {
			return true
		}
	}
	return false
}
