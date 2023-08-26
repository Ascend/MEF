// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module db opreation
package alarmmanager

import (
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"

	"huawei.com/mindxedge/base/common"
)

var (
	alarmSingleton sync.Once
	alarmDb        *AlarmDbHandler
)

// AlarmDbHandler is the struct to deal with alarm db
type AlarmDbHandler struct {
	db *gorm.DB
}

// AlarmDbInstance is a singleton instance
func AlarmDbInstance() *AlarmDbHandler {
	alarmSingleton.Do(func() {
		alarmDb = &AlarmDbHandler{
			db: database.GetDb(),
		}
	})
	return alarmDb
}

func (adh *AlarmDbHandler) addAlarmInfo(data *AlarmInfo) error {
	return adh.db.Model(AlarmInfo{}).Create(data).Error
}

func (adh *AlarmDbHandler) getNodeAlarmCount(nodeId int) (int, error) {
	var count int64
	return int(count), adh.db.Model(AlarmInfo{}).Where("node_id = ? and alarm_type = ?", nodeId,
		AlarmType).Count(&count).Error
}

func (adh *AlarmDbHandler) getAlarmInfo(alarmId string, nodeId int) (*[]AlarmInfo, error) {
	var ret []AlarmInfo
	return &ret, adh.db.Model(AlarmInfo{}).Where("alarm_id = ? and node_id = ?",
		alarmId, nodeId).Find(&ret).Error
}

func (adh *AlarmDbHandler) deleteAlarmInfo(data *AlarmInfo) error {
	return adh.db.Model(AlarmInfo{}).Where("alarm_id = ? and node_id = ?", data.AlarmId,
		data.NodeId).Delete(&data).Error
}

func (adh *AlarmDbHandler) listCenterAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) (
	*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db.Scopes(getAlarmNodeScopes(pageNum, pageSize, CenterNodeID, queryType)).
		Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listEdgeAlarmsOrEventsDb(pageNum, pageSize uint64, nodeId uint64, queryType string) (
	*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db.Scopes(getAlarmNodeScopes(pageNum, pageSize, nodeId, queryType)).Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listAllAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) (*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db.Scopes(getPagedScopes(pageNum, pageSize, queryType)).Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listAllEdgeAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) (
	*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db.Scopes(getPagedEdgeScopes(pageNum, pageSize, queryType)).Find(&alarmInfo).Error
}

func getPagedEdgeScopes(pageNum, pageSize uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(pageNum, pageSize)).Where("alarm_type=? and node_id <> ?", queryType, CenterNodeID)
	}
}

func getPagedScopes(pageNum, pageSize uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(pageNum, pageSize)).Where("alarm_type=?", queryType)
	}
}

func getAlarmNodeScopes(page, pageSize uint64, nodeId uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("alarm_type=? and node_id=?", queryType, nodeId)
	}
}

func (adh *AlarmDbHandler) getAlarmOrEventInfoByAlarmInfoId(Id uint64) (*AlarmInfo, error) {
	var alarm AlarmInfo
	return &alarm, adh.db.Model(AlarmInfo{}).Where("id=?", Id).First(&alarm).Error
}

func (adh *AlarmDbHandler) listGroupAlarmsOrEventsDb(pageNum, pageSize uint64, nodeIds []uint64, queryType string) (
	*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db.Scopes(getAlarmGroupScopes(pageNum, pageSize, nodeIds, queryType)).Find(&alarmInfo).Error
}

func getAlarmGroupScopes(page, pageSize uint64, nodeIds []uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("alarm_type=? and node_id in (?)", queryType, nodeIds)
	}
}
