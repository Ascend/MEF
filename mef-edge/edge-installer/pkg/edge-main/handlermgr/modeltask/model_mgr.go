// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package modeltask handle the model file stuff
package modeltask

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"edge-installer/pkg/common/types"

	"huawei.com/mindx/common/hwlog"
)

const concurrentDownCount = 1
const modelEventBufferSize = 1024
const globalLock = "Global"

// ModelMgr model mgr which manage all the model activeTasks
type ModelMgr struct {
	resLockMap map[string]bool
	taskLock   sync.Mutex
	taskCache  *CacheTask
	eventChan  chan IModelEvent
}

var modelMgr *ModelMgr
var once sync.Once

// GetModelMgr to get the instance of ModelMgr
func GetModelMgr() *ModelMgr {
	once.Do(func() {
		modelMgr = &ModelMgr{
			taskCache:  NewTaskCache(),
			eventChan:  make(chan IModelEvent, modelEventBufferSize),
			resLockMap: make(map[string]bool),
		}
		modelMgr.loadTasksFromDB()
		go modelMgr.startListenEvent(context.Background())
	})
	return modelMgr
}

// AddTask add a task to download or update model file
func (mo *ModelMgr) AddTask(uuid string, m types.ModelFile, ca []byte) error {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()

	oldActiveTask := mo.taskCache.getActiveTask(uuid, m.Name)
	if oldActiveTask != nil && !oldActiveTask.canUpgrade(m) {
		hwlog.RunLog.Infof("has same active task: %s_%s, skip this upgrade", uuid, m.Name)
		return nil
	}
	oldNotActiveTask := mo.taskCache.getNotActiveTask(uuid, m.Name)
	if oldNotActiveTask != nil && !oldNotActiveTask.canUpgrade(m) {
		hwlog.RunLog.Infof("has same not active task: %s_%s, skip this upgrade", uuid, m.Name)
		return nil
	}
	size, err := strconv.Atoi(m.Size)
	if err != nil {
		hwlog.RunLog.Errorf("model file size %s not correct, cannot add", m.Size)
		return fmt.Errorf("model file size %s not correct, cannot add", m.Size)
	}

	downloadCount := mo.taskCache.getDownloadingCount()
	if downloadCount >= concurrentDownCount {
		hwlog.RunLog.Errorf("task count:%d max:%d check", downloadCount, concurrentDownCount)
		return fmt.Errorf("download task reach limit count")
	}

	remainSize := mo.getOtherTaskRemainDownloadSize(uuid, m.Name)
	task := &ModelFileTask{
		Uuid:       uuid,
		ModelFile:  m,
		RemainSize: remainSize,
		Size:       size,
	}
	status := NewDownloadingStatus(task, ca)
	task.setCurrentStatus(status)
	mo.taskCache.addNotActiveTask(task)
	err = task.start()
	if err != nil {
		mo.taskCache.delNotActiveTask(uuid, m.Name)
	}
	return err
}

func (mo *ModelMgr) getOtherTaskRemainDownloadSize(uuid, name string) int {
	allTasks := mo.taskCache.getNotActiveTasksExclude(uuid, name)
	total := 0
	for _, task := range allTasks {
		if task.GetStatusType() == types.StatusDownloading {
			total += task.Size - task.DownloadedSize
		}
	}
	return total
}

// AddFailTask add a fail task which will terminated after 5 times report to FD
func (mo *ModelMgr) AddFailTask(uuid string, m types.ModelFile, reason string) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	ret := &ModelFileTask{
		Uuid:      uuid,
		ModelFile: m,
	}
	status := BuildModelStatus(&FailStatus{reason: reason}, ret)
	ret.setCurrentStatus(status)
	mo.taskCache.addPreFailTask(ret)
}

// GetActiveTask get a active task
func (mo *ModelMgr) GetActiveTask(uuid, name string) *ModelFileTask {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	return mo.taskCache.getActiveTask(uuid, name)
}

// GetFileList get the task related file list
func (mo *ModelMgr) GetFileList() []types.ModelBrief {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	return mo.taskCache.getBriefList()
}

// GetNotActiveTask get a not active task
func (mo *ModelMgr) GetNotActiveTask(uuid, name string) *ModelFileTask {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	return mo.taskCache.getNotActiveTask(uuid, name)
}

// DelNotActiveTasks delete a not active task
func (mo *ModelMgr) DelNotActiveTasks(uuid string, mFiles []types.ModelFile) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	for _, mFile := range mFiles {
		task := mo.taskCache.getNotActiveTask(uuid, mFile.Name)
		if task == nil {
			hwlog.RunLog.Warnf("not found del task %s_%s", uuid, mFile.Name)
			continue
		}
		task.cancel()
		mo.taskCache.delNotActiveTask(uuid, mFile.Name)
	}
}

// DelTaskByBriefs delete tasks in the file list
func (mo *ModelMgr) DelTaskByBriefs(briefs []types.ModelBrief) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	for _, brief := range briefs {
		if brief.Status == types.StatusActive.String() {
			mo.taskCache.delActiveTask(brief.Uuid, brief.Name)
			continue
		}
		mo.taskCache.delNotActiveTask(brief.Uuid, brief.Name)
	}
}

// DelActiveAndNotActiveTasks both delete active and not active task
func (mo *ModelMgr) DelActiveAndNotActiveTasks(uuid string, mFiles []types.ModelFile) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	for _, mFile := range mFiles {
		task := mo.taskCache.getTask(uuid, mFile.Name)
		if task == nil {
			hwlog.RunLog.Warnf("not found del task %s_%s", uuid, mFile.Name)
			continue
		}
		task.cancel()
		mo.taskCache.delNotActiveTask(uuid, mFile.Name)
		mo.taskCache.delActiveTask(uuid, mFile.Name)
	}
}

// DelTasksByUuid delete all tasks with specific
func (mo *ModelMgr) DelTasksByUuid(uuid string) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	mo.taskCache.delAllTasksByUuid(uuid)
}

// CancelTasks cancel all the tasks exist in memory
func (mo *ModelMgr) CancelTasks() {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	mo.taskCache.cancelTasks()
}

// Clear clear all tasks
func (mo *ModelMgr) Clear() {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	mo.taskCache.clear()
}

// Lock lock the resource of uuid+name
func (mo *ModelMgr) Lock(uuid, name string) bool {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	if _, ok := mo.resLockMap[globalLock]; ok {
		return false
	}
	if _, ok := mo.resLockMap[uuid]; ok {
		return false
	}
	if _, ok := mo.resLockMap[uuid+name]; ok {
		return false
	}
	mo.resLockMap[uuid+name] = true
	return true
}

// LockUuid lock the resource of uuid
func (mo *ModelMgr) LockUuid(uuid string) bool {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	if _, ok := mo.resLockMap[globalLock]; ok {
		return false
	}
	for k, _ := range mo.resLockMap {
		if strings.HasPrefix(k, uuid) {
			return false
		}
	}
	mo.resLockMap[uuid] = true
	return true
}

// LockGlobal lock the unique res, all other action cannot perform
func (mo *ModelMgr) LockGlobal() bool {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	if len(mo.resLockMap) > 1 {
		return false
	}
	mo.resLockMap[globalLock] = true
	return true
}

// UnLock unlock the resource of uuid and name
func (mo *ModelMgr) UnLock(uuid, name string) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	delete(mo.resLockMap, uuid+name)
}

// UnLockUuid unlock the resource of uuid
func (mo *ModelMgr) UnLockUuid(uuid string) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	delete(mo.resLockMap, uuid)
}

// UnLockGlobal unlock the unique res
func (mo *ModelMgr) UnLockGlobal() {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	delete(mo.resLockMap, globalLock)
}

// Active to take effect of a model file
func (mo *ModelMgr) Active(uuid, name, podUid string) error {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	task := mo.taskCache.getNotActiveTask(uuid, name)
	if task == nil {
		return fmt.Errorf("cannot active, not find task: %s_%s", uuid, name)
	}
	if task.CurrentStatus.getStatusType() != types.StatusInactive {
		return fmt.Errorf("cannot active when task in %s status", task.CurrentStatus.getStatusType())
	}
	task.PodUid = podUid
	activeStatus := BuildModelStatus(&ActiveStatus{}, task)
	task.setCurrentStatus(activeStatus)
	mo.taskCache.active(task)
	GetModelReporter().Notify()
	return nil
}

// Notify notify task mgr some event happen
func (mo *ModelMgr) Notify(event IModelEvent) {
	mo.eventChan <- event
}

func (mo *ModelMgr) buildReport() ([]*ModelProgress, int) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	mo.taskCache.cleanFailTasks()
	reportData := mo.taskCache.buildReportData()
	return reportData, mo.taskCache.getDownloadingCount()
}

func (mo *ModelMgr) buildToDbData() ([]*ModelDBRecord, []*ModelDBRecord) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	activeRecords := mo.taskCache.buildActiveRecords()
	notActiveRecords := mo.taskCache.buildNotActiveRecords()
	return activeRecords, notActiveRecords
}

func (mo *ModelMgr) startListenEvent(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("model mgr event job stop")
			return
		case e, ok := <-mo.eventChan:
			if !ok {
				continue
			}
			mo.handleModelEvent(e)
		}
	}
}

func (mo *ModelMgr) handleModelEvent(event IModelEvent) {
	mo.taskLock.Lock()
	defer mo.taskLock.Unlock()
	task := mo.taskCache.getNotActiveTask(event.GetUuid(), event.GetKey())
	if task == nil {
		hwlog.RunLog.Errorf("not found task : %s_%s", event.GetUuid(), event.GetKey())
		return
	}
	task.onEvent(event)
}

func (mo *ModelMgr) loadTasksFromDB() {
	mo.taskCache.loadTasksFromDB(modelFileKey)
	mo.taskCache.loadTasksFromDB(notActiveModelFileKey)
	GetModelReporter().Notify()
}
