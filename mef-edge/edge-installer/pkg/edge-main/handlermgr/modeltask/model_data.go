// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package modeltask

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/common/database"

	"huawei.com/mindx/common/hwlog"
)

// ModelEventType the type of model event
type ModelEventType int

// the actual type for model event
const (
	TypeProgress ModelEventType = iota + 1
	TypeDownloadFinish
)

const maxFailReportCount = 5
const maxModelFileSize = 4 * 1024 * constants.MB
const reserveModelFileSize = 10 * constants.MB
const modelFileKey = "modelfile"
const notActiveModelFileKey = "model_file_download"

// SyncList a list send to edge-om to sync files
type SyncList struct {
	FileList []types.ModelBrief `json:"fileList"`
}

// ModelProgressResp the struct to FD which show the model file info
type ModelProgressResp struct {
	ModelFiles []*ModelProgress `json:"modelfiles,omitempty"`
}

// ModelProgress the struct to FD which show the model file info
type ModelProgress struct {
	Uuid      string          `json:"uuid,omitempty"`
	PodUid    string          `json:"pod_uid"`
	Modelfile []*DownloadInfo `json:"modelfile,omitempty"`
}

// DownloadInfo the struct to FD which show the model file info
type DownloadInfo struct {
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	Reason         string `json:"reason"`
	DownloadedTime string `json:"downloaded_time"`
	Status         string `json:"status,omitempty"`
	DownloadStatus string `json:"download_status,omitempty"`
	Percentage     string `json:"percentage,omitempty"`
}

// ModelDBRecord the struct which serialize to database
type ModelDBRecord struct {
	PodUid         string `json:"pod_uid,omitempty"`
	Uuid           string `json:"uuid, omitempty"`
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	DownloadedTime string `json:"downloaded_time"`
	Status         string `json:"status,omitempty"`
	CheckType      string `json:"check_type,omitempty"`
	CheckCode      string `json:"check_code,omitempty"`
}

// ModelFileTask represent a model file task
type ModelFileTask struct {
	Uuid           string
	ModelFile      types.ModelFile
	CurrentStatus  StatusIntf
	ReportCount    int
	DownloadedTime string
	DownloadedSize int
	RemainSize     int
	Size           int
	PodUid         string
}

// StatusIntf the model status interface, which indicates the status of task
// all the statuses compose the life circle of a task. a normal success life circle of a task is as below:
// Downloading-->InActive-->Active
type StatusIntf interface {
	start() error
	getStatusType() types.ModelStatusType
	buildReportData(m ModelFileTask) *DownloadInfo
	buildRecord(m ModelFileTask) *ModelDBRecord
	onEvent(event IModelEvent)
	cancel()
	canUpgrade(m types.ModelFile) bool
	setBaseStatus(b *BaseTaskStatus)
}

// IModelEvent for multiple thread consideration
// we pass an event to model manager's main thread
type IModelEvent interface {
	GetUuid() string
	GetEventType() ModelEventType
	GetKey() string
}

// ProgressEvent an event consists of a downloading job's progress
type ProgressEvent struct {
	uuid     string
	key      string
	Progress int
}

// GetEventType get the type of an event
func (e ProgressEvent) GetEventType() ModelEventType {
	return TypeProgress
}

// GetKey get the key of the task
func (e ProgressEvent) GetKey() string {
	return e.key
}

// GetUuid get the key of the task
func (e ProgressEvent) GetUuid() string {
	return e.uuid
}

// DownloadFinishEvent an event indicates download operation finish
type DownloadFinishEvent struct {
	uuid    string
	key     string
	Reason  string
	Success bool
}

// GetEventType get the type of an event
func (e DownloadFinishEvent) GetEventType() ModelEventType {
	return TypeDownloadFinish
}

// GetUuid get the key of the task
func (e DownloadFinishEvent) GetUuid() string {
	return e.uuid
}

// GetKey get the key of the task
func (e DownloadFinishEvent) GetKey() string {
	return e.key
}

// NewProgressEvent the event when download progress change
func NewProgressEvent(uuid, key string, p int) ProgressEvent {
	return ProgressEvent{
		uuid:     uuid,
		key:      key,
		Progress: p,
	}
}

// NewDownloadFinishEvent the event when download finish
func NewDownloadFinishEvent(uuid, key, reason string, success bool) DownloadFinishEvent {
	return DownloadFinishEvent{
		uuid:    uuid,
		key:     key,
		Success: success,
		Reason:  reason,
	}
}

// CacheTask a mem cache to store all kinds of task
type CacheTask struct {
	// [uuid][name]
	activeTasks    map[string]map[string]*ModelFileTask
	notActiveTasks map[string]map[string]*ModelFileTask
	preFailTasks   map[string]map[string]*ModelFileTask
}

var buildFromDataFuncMap map[string]func(t *CacheTask, records []*ModelDBRecord)

// NewTaskCache create the task cache
func NewTaskCache() *CacheTask {
	buildFromDataFuncMap = make(map[string]func(t *CacheTask, records []*ModelDBRecord))
	buildFromDataFuncMap[modelFileKey] = buildActiveTaskFromData
	buildFromDataFuncMap[notActiveModelFileKey] = buildNotActiveTaskFromData
	return &CacheTask{
		activeTasks:    make(map[string]map[string]*ModelFileTask),
		notActiveTasks: make(map[string]map[string]*ModelFileTask),
		preFailTasks:   make(map[string]map[string]*ModelFileTask),
	}
}

func (t *CacheTask) addNotActiveTask(task *ModelFileTask) {
	tMap, ok := t.notActiveTasks[task.Uuid]
	if !ok {
		tMap = make(map[string]*ModelFileTask)
	}
	oldTask, ok := tMap[task.ModelFile.Name]
	if ok {
		oldTask.cancel()
	}

	tMap[task.ModelFile.Name] = task
	t.notActiveTasks[task.Uuid] = tMap

	// clear the pre fail notification if task is added
	fMap, ok := t.preFailTasks[task.Uuid]
	if !ok {
		fMap = make(map[string]*ModelFileTask)
		t.preFailTasks[task.Uuid] = fMap
	}
	delete(fMap, task.ModelFile.Name)
}

func (t *CacheTask) addActiveTask(task *ModelFileTask) {
	tMap, ok := t.activeTasks[task.Uuid]
	if !ok {
		tMap = make(map[string]*ModelFileTask)
	}
	tMap[task.ModelFile.Name] = task
	t.activeTasks[task.Uuid] = tMap
}

func (t *CacheTask) addPreFailTask(task *ModelFileTask) {
	tMap, ok := t.preFailTasks[task.Uuid]
	if !ok {
		tMap = make(map[string]*ModelFileTask)
	}
	tMap[task.ModelFile.Name] = task
	t.preFailTasks[task.Uuid] = tMap
}

func (t *CacheTask) newTaskFromDB(data *ModelDBRecord) *ModelFileTask {
	if data.Status != types.StatusInactive.String() && data.Status != types.StatusActive.String() {
		return nil
	}
	modelFile := types.ModelFile{
		Name:      data.Name,
		Version:   data.Version,
		CheckType: data.CheckType,
		CheckCode: data.CheckCode,
	}
	task := ModelFileTask{
		Uuid:           data.Uuid,
		ModelFile:      modelFile,
		DownloadedTime: data.DownloadedTime,
		PodUid:         data.PodUid,
	}
	if data.Status == types.StatusInactive.String() {
		inactiveStatus := BuildModelStatus(&InactiveStatus{}, &task)
		task.CurrentStatus = inactiveStatus
		return &task
	}
	activeStatus := BuildModelStatus(&ActiveStatus{}, &task)
	task.CurrentStatus = activeStatus
	return &task
}

func (t *CacheTask) getTask(uuid, name string) *ModelFileTask {
	if t.getActiveTask(uuid, name) != nil {
		return t.getActiveTask(uuid, name)
	}
	return t.getNotActiveTask(uuid, name)
}

func (t *CacheTask) getActiveTask(uuid, name string) *ModelFileTask {
	v, ok := t.activeTasks[uuid]
	if !ok {
		return nil
	}
	task, ok := v[name]
	if !ok {
		return nil
	}
	return task
}

func (t *CacheTask) getBriefList() []types.ModelBrief {
	var fileList []types.ModelBrief
	for _, tMap := range t.activeTasks {
		for _, task := range tMap {
			fileList = append(fileList, task.buildBrief())
		}
	}
	for _, tMap := range t.notActiveTasks {
		for _, task := range tMap {
			fileList = append(fileList, task.buildBrief())
		}
	}
	return fileList
}

func (t *CacheTask) getNotActiveTask(uuid, name string) *ModelFileTask {
	v, ok := t.notActiveTasks[uuid]
	if !ok {
		return nil
	}
	task, ok := v[name]
	if !ok {
		return nil
	}
	return task
}

func (t *CacheTask) active(task *ModelFileTask) {
	if _, ok := t.notActiveTasks[task.Uuid]; ok {
		delete(t.notActiveTasks[task.Uuid], task.ModelFile.Name)
	}
	t.addActiveTask(task)
}

func (t *CacheTask) getNotActiveTasksExclude(uuid, name string) []*ModelFileTask {
	var ret []*ModelFileTask
	for uuidKey, tMap := range t.notActiveTasks {
		for nameKey, task := range tMap {
			if uuidKey == uuid && nameKey == name {
				continue
			}
			ret = append(ret, task)
		}
	}
	return ret
}

func (t *CacheTask) delNotActiveTask(uuid string, name string) {
	v, ok := t.notActiveTasks[uuid]
	if !ok {
		return
	}
	delete(v, name)
	if len(v) == 0 {
		delete(t.notActiveTasks, uuid)
	}
	return
}

func (t *CacheTask) delActiveTask(uuid string, name string) {
	v, ok := t.activeTasks[uuid]
	if !ok {
		return
	}
	delete(v, name)
	if len(v) == 0 {
		delete(t.activeTasks, uuid)
	}
	return
}

func (t *CacheTask) delAllTasksByUuid(uuid string) {
	tMap, ok := t.activeTasks[uuid]
	if ok {
		for _, task := range tMap {
			task.cancel()
		}
	}
	delete(t.activeTasks, uuid)

	tMap, ok = t.notActiveTasks[uuid]
	if ok {
		for _, task := range tMap {
			task.cancel()
		}
	}
	delete(t.notActiveTasks, uuid)
}

func (t *CacheTask) clear() {
	t.activeTasks = make(map[string]map[string]*ModelFileTask)
	t.notActiveTasks = make(map[string]map[string]*ModelFileTask)
	t.preFailTasks = make(map[string]map[string]*ModelFileTask)
}

func (t *CacheTask) cancelTasks() {
	t.cancelTasksByMap(t.activeTasks)
	t.cancelTasksByMap(t.notActiveTasks)
	t.cancelTasksByMap(t.preFailTasks)
}

func (t *CacheTask) cancelTasksByMap(tasks map[string]map[string]*ModelFileTask) {
	for _, tMap := range tasks {
		for _, task := range tMap {
			task.cancel()
		}
	}
}

func (t *CacheTask) onEvent(event IModelEvent) {
	task := t.getNotActiveTask(event.GetUuid(), event.GetKey())
	if task == nil {
		hwlog.RunLog.Errorf("not found task key : %s_%s", event.GetUuid(), event.GetKey())
		return
	}
	task.onEvent(event)
}

func (t *CacheTask) getDownloadingCount() int {
	count := 0
	for _, taskMap := range t.notActiveTasks {
		for _, task := range taskMap {
			if task.GetStatusType() == types.StatusDownloading {
				count++
			}
		}
	}
	return count
}

func (t *CacheTask) cleanFailTasks() {
	for key, tMap := range t.notActiveTasks {
		t.cleanTaskMap(tMap)
		if len(tMap) == 0 {
			delete(t.notActiveTasks, key)
		}
	}
	for key, tMap := range t.preFailTasks {
		t.cleanTaskMap(tMap)
		if len(tMap) == 0 {
			delete(t.preFailTasks, key)
		}
	}
}

func (t *CacheTask) cleanTaskMap(tMap map[string]*ModelFileTask) {
	for key, value := range tMap {
		if value.GetStatusType() == types.StatusFail && value.ReportCount >= maxFailReportCount {
			delete(tMap, key)
		}
	}
}

func (t *CacheTask) buildReportData() []*ModelProgress {
	var mFiles []*ModelProgress
	mFiles = t.buildReportDataByMap(mFiles, t.activeTasks)
	mFiles = t.buildReportDataByMap(mFiles, t.notActiveTasks)
	mFiles = t.buildReportDataByMap(mFiles, t.preFailTasks)
	return mFiles
}

// buildReportDataByMap in every slice of the same uuid, we should report to FD with orders.
// the orders is: active, inactive, downloading, fail.
// the reason for the order is FD search the downloading or fail status in the slice first,
// then the active or inactive status
func (t *CacheTask) buildReportDataByMap(mFiles []*ModelProgress,
	tasks map[string]map[string]*ModelFileTask) []*ModelProgress {
	for uuid, tMap := range tasks {
		var uuidReports []*DownloadInfo
		podUid := ""
		for _, value := range tMap {
			data := value.buildReportData()
			if len(value.PodUid) > 0 {
				podUid = value.PodUid
			}
			if data != nil {
				uuidReports = append(uuidReports, data)
			}
		}
		if len(uuidReports) == 0 {
			continue
		}
		mFile := t.findMFile(uuid, mFiles)
		if mFile == nil {
			mFile = &ModelProgress{
				Uuid:      uuid,
				PodUid:    podUid,
				Modelfile: uuidReports,
			}
			mFiles = append(mFiles, mFile)
		} else {
			mFile.Modelfile = append(mFile.Modelfile, uuidReports...)
		}
	}
	return mFiles
}

func (t *CacheTask) findMFile(uuid string, mFiles []*ModelProgress) *ModelProgress {
	for _, mFile := range mFiles {
		if mFile.Uuid == uuid {
			return mFile
		}
	}
	return nil
}

func (t *CacheTask) buildActiveRecords() []*ModelDBRecord {
	return t.buildRecordsByMap(t.activeTasks)
}

func (t *CacheTask) buildNotActiveRecords() []*ModelDBRecord {
	return t.buildRecordsByMap(t.notActiveTasks)
}

func (t *CacheTask) buildRecordsByMap(tasks map[string]map[string]*ModelFileTask) []*ModelDBRecord {
	var records []*ModelDBRecord
	for _, tMap := range tasks {
		for _, value := range tMap {
			record := value.buildRecord()
			if record == nil {
				continue
			}
			records = append(records, record)
		}
	}
	return records
}

func (t *CacheTask) loadTasksFromDB(key string) {
	meta, err := database.GetMetaRepository().GetByKey(key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	if err != nil {
		hwlog.RunLog.Errorf("load task from db error: %s", err.Error())
		return
	}
	var records []*ModelDBRecord
	if err = json.Unmarshal([]byte(meta.Value), &records); err != nil {
		hwlog.RunLog.Errorf("load task unmarshal error: %s", err.Error())
		return
	}
	if _, ok := buildFromDataFuncMap[key]; !ok {
		hwlog.RunLog.Error("load task no handle func")
		return
	}
	buildFromDataFuncMap[key](t, records)
}

func buildActiveTaskFromData(t *CacheTask, records []*ModelDBRecord) {
	for _, record := range records {
		t.buildTaskFromData(t.activeTasks, record)
	}
}

func buildNotActiveTaskFromData(t *CacheTask, records []*ModelDBRecord) {
	for _, record := range records {
		t.buildTaskFromData(t.notActiveTasks, record)
	}
}

func (t *CacheTask) buildTaskFromData(tasks map[string]map[string]*ModelFileTask, data *ModelDBRecord) {
	task := t.newTaskFromDB(data)
	if task == nil {
		return
	}
	if tasks == nil {
		return
	}
	tMap, ok := tasks[data.Uuid]
	if !ok {
		tMap = make(map[string]*ModelFileTask)
	}
	tMap[data.Name] = task
	tasks[data.Uuid] = tMap
}
