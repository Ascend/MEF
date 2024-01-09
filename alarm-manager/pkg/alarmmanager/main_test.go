// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for package main test
package alarmmanager

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

var (
	groupNodesMap = map[string]bool{testSn1: true}
	// ensure testSn is in db
	testSns        = []string{testSn1, testSn2, ""}
	defaultAlarmID uint64
	defaultEventID uint64
)

const (
	NodeNums        = 4
	TypesOfSeverity = 3
	Possibility     = 10
	HalfPossibility = 5
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
			bytes, err := json.Marshal(resp)
			convey.So(err, convey.ShouldBeNil)
			return &model.Message{
				Content: bytes,
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

func randSetOneAlarm(alarm *AlarmInfo) {
	severities := []string{"MINOR", "MAJOR", "CRITICAL"}
	defaultInfo := "ALARM DEFAULT INFO"
	defaultSuggest := "ALARM DEFAULT SUGGESTION"
	defaultReason := "ALARM DEFAULT Reason"
	defaultImpact := "ALARM DEFAULT Impact"
	defaultAlarmName := "ALARM DEFAULT NAME"
	defaultAlarmResource := "ALARM DEFAULT RESOURCE"
	randType := alarms.AlarmType
	var randNum, randNodeId, randTypeSe int
	randNodeId, err1 := randIntn(NodeNums)
	randTypeSe, err2 := randIntn(TypesOfSeverity - 1)
	randNum, err3 := randIntn(Possibility)
	if err1 != nil || err2 != nil || err3 != nil {
		hwlog.RunLog.Error("failed to generate random id")
		return
	}
	if randNum < HalfPossibility {
		randType = alarms.EventType
	}
	alarm.AlarmType = randType
	alarm.CreatedAt = time.Now()
	alarm.SerialNumber = strconv.Itoa(randNodeId)
	alarmId, err := randIntn(math.MaxUint32)
	if err != nil {
		hwlog.RunLog.Error("failed to generate alarm id")
		return
	}
	alarm.AlarmId = strconv.Itoa(alarmId)
	alarm.AlarmName = defaultAlarmName
	typeSe := randTypeSe
	if typeSe >= len(severities) {
		return
	}
	alarm.PerceivedSeverity = severities[typeSe]
	alarm.DetailedInformation = defaultInfo
	alarm.Suggestion = defaultSuggest
	alarm.Reason = defaultReason
	alarm.Impact = defaultImpact
	alarm.Resource = defaultAlarmResource
}
