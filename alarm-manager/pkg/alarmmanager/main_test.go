// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for package main test
package alarmmanager

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcAlarmMgr := &TcAlarmMgr{
		tcBaseWithDb: &test.TcBaseWithDb{
			Tables: append(tables, &AlarmInfo{}),
		},
	}

	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFunc(modulemgr.SendSyncMessage, func(m *model.Message, duration time.Duration) (*model.Message,
			error) {
			resp := common.RespMsg{
				Status: common.Success,
				Data:   []string{testSn1},
			}
			return &model.Message{
				Content: resp,
			}, nil
		}).
		ApplyMethod(&httpsmgr.HttpsRequest{}, "GetWithTimeout",
			func(req *httpsmgr.HttpsRequest, body io.Reader, timeout time.Duration) ([]byte, error) {
				var resp common.RespMsg
				nodeGroup := []string{testSn1}
				resp.Data = nodeGroup
				resp.Status = common.Success
				bytes, err := json.Marshal(resp)
				if err != nil {
					fmt.Println("error marshalling")
					return nil, err
				}
				return bytes, nil
			})

	test.RunWithPatches(tcAlarmMgr, m, patches)
}

// TcBase struct for test case
type TcAlarmMgr struct {
	tcBaseWithDb *test.TcBaseWithDb
}

// Setup pre-processing
func (tc *TcAlarmMgr) Setup() error {
	if err := tc.tcBaseWithDb.Setup(); err != nil {
		return err
	}
	createInitialData()
	alarmSlice, err := AlarmDbInstance().listAllAlarmsOrEventsDb(firstPageNum, firstPageSize, alarms.AlarmType)
	if err != nil {
		return err
	}
	defaultAlarmID = alarmSlice[0].Id
	eventSlice, err := AlarmDbInstance().listAllAlarmsOrEventsDb(firstPageNum, firstPageSize, alarms.EventType)
	if err != nil {
		return err
	}
	defaultEventID = eventSlice[0].Id
	return nil
}

// Teardown post-processing
func (tc *TcAlarmMgr) Teardown() {
	tc.tcBaseWithDb.Teardown()
}

func createInitialData() {
	const InitialRecordNums = 100
	res := make([]AlarmInfo, InitialRecordNums)

	for idx, alarm := range res {
		randSetOneAlarm(&alarm)
		alarm.SerialNumber = testSns[idx%len(testSns)]
		if err := AlarmDbInstance().addAlarmInfo(&alarm); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
}

func randIntn(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -1, err
	}
	randNum := int((*n).Int64())
	return randNum, nil
}
