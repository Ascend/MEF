// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgeproxy common job definition
package edgeproxy

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

const (
	neverStop = 0
	doOneTime = -1
)

// ConnLoopJobFunc job content and run interval
type ConnLoopJobFunc struct {
	// define interval to do the job
	interval time.Duration
	// define what you want to do.
	// ***CAUTION: any error will cancel all other jobs and close the connection.
	// if your job doesn't want to disturb other jobs, return nil
	do func() error
}

// JobProxy all jobs and related websocket connection
type JobProxy struct {
	conn *websocket.Conn
	jobs []ConnLoopJobFunc
	ctx  context.Context    // for sync STOP info across all jobs
	cf   context.CancelFunc // trigger STOP info to all jobs
}

// JobController abstract job and related websocket connection
type JobController interface {
	GetJobProxy() *JobProxy
}

// ProcessJob process each job concurrently
func ProcessJob(jc JobController) {
	jobProxy := jc.GetJobProxy()
	for _, eachJob := range jobProxy.jobs {
		switch eachJob.interval {
		case neverStop:
			go infiniteJob(jobProxy, eachJob)
		case doOneTime:
			go oneTimeJob(eachJob)
		default:
			go intervalJob(jobProxy, eachJob)
		}
	}
}

func infiniteJob(jobProxy *JobProxy, job ConnLoopJobFunc) {
	for {
		select {
		case <-jobProxy.ctx.Done():
			return
		default:
		}
		if err := job.do(); err != nil {
			jobProxy.cf()
			return
		}
	}
}

func oneTimeJob(job ConnLoopJobFunc) {
	if err := job.do(); err != nil {
		return
	}
}

func intervalJob(jobProxy *JobProxy, job ConnLoopJobFunc) {
	for {
		select {
		case <-jobProxy.ctx.Done():
			return
		default:
		}
		if err := job.do(); err != nil {
			jobProxy.cf()
			return
		}
		time.Sleep(job.interval)
	}
}
