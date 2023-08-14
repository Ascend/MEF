// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import (
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
)

var (
	alarmSingleton    sync.Once
	alarmDbHandlerIns *AlarmDbHandler
)

// AlarmDbHandler is the struct to deal with alarm db
type AlarmDbHandler struct {
	db *gorm.DB
}

// AlarmDbInstance is a singleton instance
func AlarmDbInstance() *AlarmDbHandler {
	alarmSingleton.Do(func() {
		alarmDbHandlerIns = &AlarmDbHandler{
			db: database.GetDb(),
		}
	})
	return alarmDbHandlerIns
}

func (adh *AlarmDbHandler) addAlarmStaticInfo(data *AlarmStaticInfo) error {
	return adh.db.Model(AlarmStaticInfo{}).Create(data).Error
}

func (adh *AlarmDbHandler) addAlarmInfo(data *AlarmInfo) error {
	return adh.db.Model(AlarmInfo{}).Create(data).Error
}

func (adh *AlarmDbHandler) getAlarmInfo(alarmId string, nodeId int) (*AlarmInfo, error) {
	var ret AlarmInfo
	return &ret, adh.db.Model(AlarmInfo{}).Where("alarm_id = ? and node_id = ?", alarmId, nodeId).First(&ret).Error
}

func (adh *AlarmDbHandler) getAlarmStaticInfo(alarmId string) (*AlarmStaticInfo, error) {
	var ret AlarmStaticInfo
	return &ret, adh.db.Model(AlarmStaticInfo{}).Where("alarm_id = ?", alarmId).First(&ret).Error
}

func (adh *AlarmDbHandler) deleteAlarmInfo(data *AlarmInfo) error {
	return adh.db.Model(AlarmInfo{}).Where("alarm_id = ? and node_id = ?", data.AlarmId, data.NodeId).Delete(&data).Error
}
