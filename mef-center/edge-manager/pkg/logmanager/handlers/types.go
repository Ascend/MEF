// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers
package handlers

import (
	"encoding/json"
	"time"

	"huawei.com/mindxedge/base/common/taskschedule"
)

// CreateTaskReq defines request for creating task
type CreateTaskReq struct {
	Module    string   `json:"module"`
	EdgeNodes []uint64 `json:"edgeNodes"`
}

// CreateTaskResp defines response for creating task
type CreateTaskResp struct {
	TaskId string `json:"taskId"`
}

// QueryProgressResp defines response for querying task progress
type QueryProgressResp struct {
	TaskId     string                  `json:"taskId"`
	Status     taskschedule.TaskPhase  `json:"status"`
	Reason     string                  `json:"reason"`
	Progress   uint                    `json:"progress"`
	Data       taskschedule.JsonObject `json:"data"`
	StartedAt  NullableTime            `json:"startedAt"`
	CreatedAt  NullableTime            `json:"createdAt"`
	FinishedAt NullableTime            `json:"finishedAt"`
}

// TaskErrorInfo defines task error info
type TaskErrorInfo struct {
	Id      string `json:"id"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// NullableTime enhanced time format
type NullableTime time.Time

// MarshalJSON marshal json
func (nt NullableTime) MarshalJSON() ([]byte, error) {
	t := time.Time(nt)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t)
}
