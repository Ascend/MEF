// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"context"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
)

type taskContextImpl struct {
	context.Context
	shutdown            context.CancelFunc
	mainCtx             context.Context
	mainCtxCancel       context.CancelFunc
	gracefulShutdownCtx context.Context
	gracefulShutdown    context.CancelFunc

	spec TaskSpec

	repo       taskRepository
	phase      TaskPhase
	updates    chan taskUpdateRequest
	heartbeats chan struct{}
	doneEvents chan<- struct{}
}

func (t *taskContextImpl) Spec() TaskSpec {
	return t.spec
}

func (t *taskContextImpl) GracefulShutdown() <-chan struct{} {
	if t.gracefulShutdownCtx == nil {
		return closedEmptyChannel
	}
	return t.gracefulShutdownCtx.Done()
}

func (t *taskContextImpl) UpdateStatus(status TaskStatus) error {
	return t.updateStatus(status, true)
}

func (t *taskContextImpl) UpdateLiveness() error {
	var err error
	func() {
		defer func() {
			if data := recover(); data != nil {
				err = fmt.Errorf("update_liveness_error(%v)", data)
			}
		}()
		select {
		case t.heartbeats <- struct{}{}:
		default:
		}
	}()
	return err
}

func (t *taskContextImpl) GetStatus() (TaskStatus, error) {
	task, err := t.repo.getTask(t.spec.Id)
	return task.Status, err
}

func (t *taskContextImpl) GetSubTaskTree() (TaskTreeNode, error) {
	return t.repo.getTaskTree(t.spec.Id)
}

func (t *taskContextImpl) Cancel() {
	if t.mainCtxCancel == nil {
		return
	}
	t.mainCtxCancel()
}

func (t *taskContextImpl) updateStatus(status TaskStatus, byUser bool) error {
	if t.spec.HeartbeatTimeout > 0 && !t.phase.IsFinished() {
		if err := t.UpdateLiveness(); err != nil {
			hwlog.RunLog.Errorf("(taskId=%s)failed to update liveness, %v", t.spec.Id, err)
		}
	}

	var err error
	func() {
		defer func() {
			if data := recover(); data != nil {
				err = fmt.Errorf("update_status_error(%v)", data)
			}
		}()
		err = t.tryUpdateStatus(status, byUser)
	}()
	return err
}

func (t *taskContextImpl) tryUpdateStatus(status TaskStatus, byUser bool) error {
	if t.updates == nil {
		return ErrTaskAlreadyFinished
	}
	const sendSyncMessageTimeout = 5 * time.Second
	timer := time.NewTimer(sendSyncMessageTimeout)
	defer timer.Stop()

	var resp taskUpdateResponse
	respCh := make(chan taskUpdateResponse, 1)
	select {
	case t.updates <- taskUpdateRequest{newStatus: status, respCh: respCh, byUser: byUser}:
	case <-timer.C:
		return ErrTimeout
	}

	select {
	case resp = <-respCh:
	case <-timer.C:
		return ErrTimeout
	}

	if resp.rowsAffected <= 0 {
		return ErrNoRowsAffected
	}
	return nil
}

type taskUpdateRequest struct {
	newStatus TaskStatus
	respCh    chan taskUpdateResponse
	byUser    bool
}

type taskUpdateResponse struct {
	updatedStatus TaskStatus
	rowsAffected  int64
	err           error
}

var (
	forcedShutdownReq   = taskUpdateRequest{newStatus: TaskStatus{Phase: Failed}}
	gracefulShutdownReq = taskUpdateRequest{newStatus: TaskStatus{Phase: Aborting}}
)

func (t *taskContextImpl) onWaiting() {
	var (
		waitTimer   *time.Timer
		waitTimerCh <-chan time.Time = alwaysOpenTimeChannel
	)
	if t.spec.WaitTimeout > 0 {
		waitTimer = time.NewTimer(t.spec.WaitTimeout)
		waitTimerCh = waitTimer.C
	}
	if waitTimerCh == nil {
		waitTimerCh = make(chan time.Time)
	}

	for {
		if t.phase != Waiting {
			break
		}
		select {
		case updateReq, ok := <-t.updates:
			if !ok {
				hwlog.RunLog.Info("task update request chan has closed")
				break
			}
			t.handleUpdateRequest(updateReq)
		case _, _ = <-t.mainCtx.Done():
			t.handleUpdateRequest(forcedShutdownReq)
		case _, _ = <-waitTimerCh:
			t.handleUpdateRequest(forcedShutdownReq)
		}
	}

	if waitTimer != nil {
		waitTimer.Stop()
	}
	hwlog.RunLog.Debugf("(taskId=%s)phase transform: waiting=>%s", t.spec.Id, t.phase)
}

func (t *taskContextImpl) onProcessing() {
	var (
		heartbeatMonitoring = alwaysOpenEmptyChannel
		executeTimer        *time.Timer
		executeTimerCh      <-chan time.Time
	)
	if t.spec.HeartbeatTimeout > 0 {
		heartbeatMonitoring = make(chan struct{})
		go doHeartbeatMonitoring(t, t.spec.HeartbeatTimeout, t.heartbeats, heartbeatMonitoring)
	}
	if t.spec.ExecuteTimeout > 0 {
		executeTimer = time.NewTimer(t.spec.ExecuteTimeout)
		executeTimerCh = executeTimer.C
	}
	if executeTimerCh == nil {
		executeTimerCh = make(chan time.Time)
	}
	if heartbeatMonitoring == nil {
		heartbeatMonitoring = make(chan struct{})
	}

	for {
		if t.phase != Processing {
			break
		}

		select {
		case updateReq, ok := <-t.updates:
			if !ok {
				hwlog.RunLog.Info("task update request chan has closed")
				break
			}
			t.handleUpdateRequest(updateReq)
		case _, _ = <-heartbeatMonitoring:
			t.handleUpdateRequest(gracefulShutdownReq)
		case _, _ = <-executeTimerCh:
			t.handleUpdateRequest(gracefulShutdownReq)
		case _, _ = <-t.mainCtx.Done():
			t.handleUpdateRequest(gracefulShutdownReq)
		}
	}

	if executeTimer != nil {
		executeTimer.Stop()
	}
	t.mainCtxCancel()
	hwlog.RunLog.Debugf("(taskId=%s)phase transform: processing=>%s", t.spec.Id, t.phase)
}

func (t *taskContextImpl) onAborting() {
	defer func() {
		close(t.heartbeats)
		close(t.updates)
		t.shutdown()
		hwlog.RunLog.Debugf("(taskId=%s)phase transform: aborting=>%s", t.spec.Id, t.phase)

		if t.spec.ParentId != "" {
			return
		}
		select {
		case t.doneEvents <- struct{}{}:
		default:
		}
	}()
	if t.spec.GracefulShutdownTimeout <= 0 {
		t.handleUpdateRequest(forcedShutdownReq)
		t.gracefulShutdown()
		return
	}

	t.gracefulShutdown()
	gracefulShutdownTimer := time.NewTimer(t.spec.GracefulShutdownTimeout)
	gracefulShutdownTimerCh := gracefulShutdownTimer.C
	if gracefulShutdownTimerCh == nil {
		gracefulShutdownTimerCh = make(chan time.Time)
	}
	for {
		if t.phase != Aborting {
			break
		}

		select {
		case updateReq, ok := <-t.updates:
			if !ok {
				hwlog.RunLog.Info("task update request chan has closed")
				break
			}
			t.handleUpdateRequest(updateReq)
		case _, _ = <-gracefulShutdownTimerCh:
			t.handleUpdateRequest(forcedShutdownReq)
		}
	}
	gracefulShutdownTimer.Stop()
}

func (t *taskContextImpl) handleUpdateRequest(req taskUpdateRequest) {
	if !allowPhaseTrans(t.phase, req.newStatus.Phase, req.byUser) {
		if req.respCh != nil {
			req.respCh <- taskUpdateResponse{rowsAffected: 0, err: ErrTaskAlreadyFinished}
		}
		return
	}
	if req.newStatus.Phase.IsFinished() && req.newStatus.FinishedAt.IsZero() {
		req.newStatus.FinishedAt = time.Now()
	}

	task, rowsAffected, err := t.repo.updateTaskStatus(t.spec.Id, req.newStatus)
	if err != nil {
		hwlog.RunLog.Errorf("(taskId=%s)failed to update task status, %v", t.spec.Id, err)
	}
	if err == nil || !req.byUser {
		t.phase = task.Status.Phase
	}
	if req.respCh != nil {
		req.respCh <- taskUpdateResponse{updatedStatus: task.Status, err: err, rowsAffected: rowsAffected}
	}
}

const (
	weight0 = 0
	weight1 = 1
	weight2 = 2
	weight3 = 3
)

var phaseWeights = map[TaskPhase]int{
	Waiting:         weight0,
	Processing:      weight1,
	Aborting:        weight2,
	Succeed:         weight3,
	Failed:          weight3,
	PartiallyFailed: weight3,
}

func allowPhaseTrans(from, to TaskPhase, byUser bool) bool {
	if to == "" {
		to = from
	}

	if from.IsFinished() {
		return false
	}
	if to.IsFinished() || to == from {
		return true
	}
	fromWeight, ok := phaseWeights[from]
	if !ok {
		return false
	}
	toWeight, ok := phaseWeights[to]
	if !ok {
		return false
	}
	return !byUser && fromWeight < toWeight
}

func doHeartbeatMonitoring(
	ctx context.Context, duration time.Duration, heartbeat <-chan struct{}, notifyCh chan<- struct{}) {
	timer := time.NewTimer(duration)
forLoop:
	for {
		if heartbeat == nil {
			break
		}
		select {
		case <-timer.C:
			break forLoop
		case _, ok := <-heartbeat:
			if !ok {
				break forLoop
			}
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(duration)
		case _, _ = <-ctx.Done():
			break forLoop
		}
	}

	close(notifyCh)
}

func createTaskContext(taskSpec *TaskSpec, parent *taskContextImpl, rootCtx context.Context,
	doneEvents chan<- struct{}, repo taskRepository) (*taskContextImpl, error) {
	if taskSpec.Id == "" {
		uuid, err := newRandomID()
		if err != nil {
			return nil, err
		}
		if taskSpec.Name == "" {
			taskSpec.Id = uuid
		} else {
			taskSpec.Id = fmt.Sprintf("%s.%s", taskSpec.Name, uuid)
		}
	}
	parentCtx := rootCtx
	if parent != nil {
		parentCtx = parent.mainCtx
	}
	mainCtx, mainCtxCancel := context.WithCancel(parentCtx)
	shutdownCtx, shutdownCtxCancel := context.WithCancel(context.Background())
	gracefulShutdownCtx, gracefulShutdownCancel := context.WithCancel(context.Background())
	return &taskContextImpl{
		Context:             shutdownCtx,
		shutdown:            shutdownCtxCancel,
		mainCtx:             mainCtx,
		mainCtxCancel:       mainCtxCancel,
		gracefulShutdownCtx: gracefulShutdownCtx,
		gracefulShutdown:    gracefulShutdownCancel,
		spec:                *taskSpec,
		repo:                repo,
		phase:               Waiting,
		updates:             make(chan taskUpdateRequest),
		heartbeats:          make(chan struct{}),
		doneEvents:          doneEvents,
	}, nil
}

func (t *taskContextImpl) destroy() {
	t.mainCtxCancel()
	t.gracefulShutdown()
	t.shutdown()
}

func (t *taskContextImpl) startLifeCycleEventsMonitoring() {
	go func() {
		t.onWaiting()
		t.onProcessing()
		t.onAborting()
	}()
}

func getHistoryTaskContext(repo taskRepository, taskId string) (TaskContext, error) {
	task, err := repo.getTask(taskId)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	return &taskContextImpl{
		Context: closedContext,
		spec:    task.Spec,
		repo:    repo,
		phase:   task.Status.Phase,
	}, nil
}
