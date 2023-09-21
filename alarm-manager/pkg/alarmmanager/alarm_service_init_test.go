// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for init packages testing
package alarmmanager

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

var (
	gormInstance  *gorm.DB
	dbPath        = "./test.db"
	pachers       = make([]*gomonkey.Patches, 0)
	groupNodesMap = map[string]bool{testSn1: true}
	// ensure testSn is in db
	testSns = []string{testSn1, testSn2, ""}
)

const (
	InitDbFlag        = true
	NodeNums          = 4
	InitialRecordNums = 100
	MaxAlarmNum       = 1000
	MinAlarmNum       = 3
	TypesOfSevirity   = 3
	DefaultAlarmID    = uint64(1)
	DefaultEventID    = uint64(2)
	Possibility       = 10
	HalfPossibility   = 5
	defaultAlarmId    = 0
	defaultEventID    = 1
)

func setup() {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
		return
	}
	if InitDbFlag {
		if err = os.Remove(dbPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			hwlog.RunLog.Errorf("cleanup db failed, error: %v", err)
			return
		}
	}
	gormInstance, err = gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		hwlog.RunLog.Errorf("failed to init test db, %v", err)
		return
	}
	if err = gormInstance.AutoMigrate(&AlarmInfo{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
		return
	}
	if InitDbFlag {
		createInitialData()
	}
}

func teardown() {
	if err := os.Remove(dbPath); err != nil && errors.Is(err, os.ErrExist) {
		hwlog.RunLog.Errorf("cleanup [%s] failed, error: %s", dbPath, err.Error())
	}
}

func setupPachers() {
	p1 := gomonkey.ApplyFunc(database.GetDb, mockGetDb)
	p2 := gomonkey.ApplyFunc(modulemgr.SendSyncMessage, func(m *model.Message, duration time.Duration) (*model.Message,
		error) {
		resp := common.RespMsg{
			Status: common.Success,
			Data:   []string{testSn1},
		}
		return &model.Message{
			Content: resp,
		}, nil
	})
	p3 := gomonkey.ApplyMethod(&httpsmgr.HttpsRequest{}, "GetWithTimeout",
		func(req *httpsmgr.HttpsRequest, body io.Reader, timeout time.Duration) ([]byte, error) {
			var resp common.RespMsg
			nodeGroup := genNodeGroup()
			resp.Data = nodeGroup
			resp.Status = common.Success
			bytes, err := json.Marshal(resp)
			if err != nil {
				hwlog.RunLog.Error("error marshalling")
				return nil, err
			}
			return bytes, nil
		})
	pachers = append(pachers, p1, p2, p3)
}

func mockGetDb() *gorm.DB {
	return gormInstance
}

func TestMain(m *testing.M) {
	setupPachers()
	setup()
	code := m.Run()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
	defer func() {
		teardown()
		for _, p := range pachers {
			p.Reset()
		}
	}()
}

func createInitialData() {
	genRandomAlarmStaticInfo(InitialRecordNums)
}

func genRandomAlarmStaticInfo(num int) {
	if num <= 0 || num >= MaxAlarmNum {
		hwlog.RunLog.Error("the number of records is exceeded")
		return
	}
	res := make([]AlarmInfo, num)

	for idx, alarm := range res {
		if num < MinAlarmNum {
			hwlog.RunLog.Errorf("testing db alarms should be more than %d records", MinAlarmNum)
			return
		}
		randSetOneAlarm(&alarm, idx)

		alarm.SerialNumber = testSns[idx%len(testSns)]

		err := AlarmDbInstance().addAlarmInfo(&alarm)
		if err != nil {
			hwlog.RunLog.Error(err.Error())
			continue
		}
	}
}

func randSetOneAlarm(alarm *AlarmInfo, idx int) {
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
	randTypeSe, err2 := randIntn(TypesOfSevirity - 1)
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
	alarmId, err := randIntn(math.MaxInt64)
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
	// make sure first record is an alarm and second record is an event for testing get user interface
	if idx == defaultAlarmId {
		alarm.AlarmType = alarms.AlarmType
	}
	if idx == defaultEventID {
		alarm.AlarmType = alarms.EventType
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

func genNodeGroup() []string {
	return []string{testSn1}
}
