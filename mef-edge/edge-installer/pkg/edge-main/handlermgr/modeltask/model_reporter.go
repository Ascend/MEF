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
	"context"
	"encoding/json"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/edge-main/common/database"
)

const (
	reportIdleInterval = time.Second * 60
	reportInterval     = time.Second * 5
)

var (
	modelReporterOnce sync.Once
	instance          *ModelReporter
)

// ModelReporter struct for report model file info to FD
type ModelReporter struct {
	eventChan  chan interface{}
	reportTick *time.Ticker
}

// GetModelReporter get the singleton of a model reporter
func GetModelReporter() *ModelReporter {
	modelReporterOnce.Do(func() {
		const eventChanSize = 1024
		instance = &ModelReporter{
			eventChan:  make(chan interface{}, eventChanSize),
			reportTick: time.NewTicker(reportInterval),
		}
	})
	return instance
}

// Notify notify a thread to report to fd
func (m *ModelReporter) Notify() {
	m.eventChan <- struct{}{}
}

// StartReportJob start the model file info report job
func (m *ModelReporter) StartReportJob(ctx context.Context) {
	hwlog.RunLog.Info("model report job start")
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("model report job stop")
			return
		case _, ok := <-m.eventChan:
			if ok {
				reportData, downloadingCount := GetModelMgr().buildReport()
				m.report(reportData, downloadingCount)
				activeRecords, notActiveRecords := GetModelMgr().buildToDbData()
				m.syncDataToDB(modelFileKey, activeRecords)
				m.syncDataToDB(notActiveModelFileKey, notActiveRecords)
			}
		case <-m.reportTick.C:
			reportData, downloadingCount := GetModelMgr().buildReport()
			m.report(reportData, downloadingCount)
			activeRecords, notActiveRecords := GetModelMgr().buildToDbData()
			m.syncDataToDB(modelFileKey, activeRecords)
			m.syncDataToDB(notActiveModelFileKey, notActiveRecords)
		}
	}
}

func (m *ModelReporter) report(data []*ModelProgress, downloadingCount int) {
	if downloadingCount > 0 {
		m.reportTick.Reset(reportInterval)
	} else {
		m.reportTick.Reset(reportIdleInterval)
	}
	m.doReport(data)
}

func (m *ModelReporter) doReport(data []*ModelProgress) {
	if len(data) > 0 {
		resp := ModelProgressResp{
			ModelFiles: data,
		}
		SendReport(resp)
		return
	}
}

func (m *ModelReporter) syncDataToDB(key string, records []*ModelDBRecord) {
	dataBytes, err := json.Marshal(records)
	if err != nil {
		hwlog.RunLog.Errorf("sync data to db error: %s", err.Error())
		return
	}
	err = database.GetMetaRepository().CreateOrUpdate(database.Meta{
		Key:   key,
		Type:  key,
		Value: string(dataBytes),
	})
	if err != nil {
		hwlog.RunLog.Errorf("save records to db failed, please check the environment")
	}
}
