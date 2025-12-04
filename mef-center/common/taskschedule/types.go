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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"huawei.com/mindx/common/utils"
)

// TaskPhase phase type
type TaskPhase string

// phase constants
const (
	Waiting         TaskPhase = "waiting"
	Processing      TaskPhase = "processing"
	Aborting        TaskPhase = "aborting"
	Succeed         TaskPhase = "succeed"
	Failed          TaskPhase = "failed"
	PartiallyFailed TaskPhase = "partiallyFailed"
)

// JsonObject map object
type JsonObject map[string]interface{}

// Task struct
type Task struct {
	Spec   TaskSpec   `json:"spec"   gorm:"embedded"`
	Status TaskStatus `json:"status" gorm:"embedded"`
}

// TaskStatus task status
type TaskStatus struct {
	Phase      TaskPhase  `json:"phase"     gorm:"type:text; not null"`
	Reason     string     `json:"reason"`
	Message    string     `json:"message"`
	Progress   uint       `json:"progress"  gorm:"not null"`
	Data       JsonObject `json:"data"      gorm:"type:json"`
	StartedAt  time.Time  `json:"startedAt"`
	CreatedAt  time.Time  `json:"createdAt" gorm:"not null"`
	FinishedAt time.Time  `json:"finishedAt"`
}

// TaskSpec task spec
type TaskSpec struct {
	Id                      string        `json:"id"            gorm:"primaryKey; not null"`
	Name                    string        `json:"name"`
	ParentId                string        `json:"parentId"`
	GoroutinePool           string        `json:"goroutinePool" gorm:"not null"`
	Command                 string        `json:"executor"      gorm:"not null"`
	Args                    JsonObject    `json:"args"          gorm:"type:json"`
	WaitTimeout             time.Duration `json:"waitTimeout"`
	HeartbeatTimeout        time.Duration `json:"heartbeatTimeout"`
	ExecuteTimeout          time.Duration `json:"executeTimeout"`
	GracefulShutdownTimeout time.Duration `json:"gracefulShutdownTimeout"`
}

// TaskTreeNode struct
type TaskTreeNode struct {
	Current  *Task
	Children []TaskTreeNode
}

// GoroutinePool struct
type GoroutinePool struct {
	Spec GoroutinePoolSpec
}

// GoroutinePoolSpec struct
type GoroutinePoolSpec struct {
	Id             string
	MaxConcurrency uint
	MaxCapacity    uint
}

// SchedulerSpec struct
type SchedulerSpec struct {
	MaxHistoryMasterTasks int64
	MaxActiveTasks        int64
	AllowedMaxTasksInDb   int
}

// IsFinished tests phase is finished
func (p TaskPhase) IsFinished() bool {
	return p == PartiallyFailed || p == Failed || p == Succeed
}

// Scan scans bytes into JsonObject
func (j *JsonObject) Scan(data interface{}) error {
	bytes, ok := data.([]byte)
	if !ok {
		return errors.New("data type is invalid")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// Value converts JsonObject to bytes
func (j JsonObject) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Get load specific field to dst
func (j JsonObject) Get(field string, dst interface{}) error {
	if j == nil {
		return errors.New("nil map")
	}

	fieldObj, ok := j[field]
	if !ok {
		return errors.New("field not found")
	}
	return utils.ObjectConvert(fieldObj, dst)
}
