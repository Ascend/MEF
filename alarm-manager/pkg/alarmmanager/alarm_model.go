// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager for alarm-manager module db opreation
package alarmmanager

import (
	"sync"

	"gorm.io/gorm"

	"huawei.com/mindx/common/database"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

var (
	alarmSingleton sync.Once
	alarmDb        *AlarmDbHandler
)

// AlarmDbHandler is the struct to deal with alarm db
type AlarmDbHandler struct {
}

// AlarmDbInstance is a singleton instance
func AlarmDbInstance() *AlarmDbHandler {
	alarmSingleton.Do(func() {
		alarmDb = &AlarmDbHandler{}
	})
	return alarmDb
}

func (adh *AlarmDbHandler) db() *gorm.DB {
	return database.GetDb()
}

func (adh *AlarmDbHandler) addAlarmInfo(data *AlarmInfo) error {
	return adh.db().Model(AlarmInfo{}).Create(data).Error
}

func (adh *AlarmDbHandler) getNodeAlarmCount(sn string) (int, error) {
	var count int64
	return int(count), adh.db().Model(AlarmInfo{}).Where("serial_number = ? and alarm_type = ?", sn, alarms.AlarmType).
		Count(&count).Error
}

func (adh *AlarmDbHandler) getNodeEventCount(sn string) (int, error) {
	var count int64
	return int(count), adh.db().Model(AlarmInfo{}).Where("serial_number = ? and alarm_type = ?", sn, alarms.EventType).
		Count(&count).Error
}

func (adh *AlarmDbHandler) getNodeOldEvent(sn string, offset int) ([]AlarmInfo, error) {
	var ret []AlarmInfo
	return ret, adh.db().Model(AlarmInfo{}).Where("serial_number = ? and alarm_type = ?",
		sn, alarms.EventType).Order("created_at").Offset(offset).Find(&ret).Error
}

func (adh *AlarmDbHandler) getAlarmInfo(alarmId string, sn string) ([]AlarmInfo, error) {
	var ret []AlarmInfo
	return ret, adh.db().Model(AlarmInfo{}).Where("alarm_id = ? and serial_number = ?", alarmId, sn).Find(&ret).Error
}

func (adh *AlarmDbHandler) deleteAlarmInfo(data []AlarmInfo) error {
	return adh.db().Model(AlarmInfo{}).Delete(&data).Error
}

func (adh *AlarmDbHandler) deleteBySn(sn string) error {
	return adh.db().Model(AlarmInfo{}).Where("serial_number = ?", sn).Delete(AlarmInfo{}).Error
}

// DeleteAlarmTable is the func to delete all alarm table
func (adh *AlarmDbHandler) DeleteAlarmTable() error {
	return database.DropTableIfExist(&AlarmInfo{})
}

// DeleteEdgeAlarm is the func to delete all alarm from MEFEdge
func (adh *AlarmDbHandler) DeleteEdgeAlarm() error {
	return adh.db().Model(AlarmInfo{}).Where("serial_number != ?", alarms.CenterSn).Delete(AlarmInfo{}).Error
}

func (adh *AlarmDbHandler) listCenterAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) (
	[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return alarmInfo, adh.db().Scopes(getAlarmNodeScopes(pageNum, pageSize, alarms.CenterSn, queryType)).
		Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listEdgeAlarmsOrEventsDb(pageNum, pageSize uint64, sn string, queryType string) (
	[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return alarmInfo, adh.db().Scopes(getAlarmNodeScopes(pageNum, pageSize, sn, queryType)).Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listAllAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) ([]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return alarmInfo, adh.db().Scopes(getPagedScopes(pageNum, pageSize, queryType)).Find(&alarmInfo).Error
}

func (adh *AlarmDbHandler) listAllEdgeAlarmsOrEventsDb(pageNum, pageSize uint64, queryType string) (
	[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return alarmInfo, adh.db().Scopes(getPagedEdgeScopes(pageNum, pageSize, queryType)).Find(&alarmInfo).Error
}

func getPagedEdgeScopes(pageNum, pageSize uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(pageNum, pageSize)).Where("alarm_type=? and serial_number <> ?",
			queryType, alarms.CenterSn)
	}
}

func getPagedScopes(pageNum, pageSize uint64, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(pageNum, pageSize)).Where("alarm_type=?", queryType)
	}
}

func getAlarmNodeScopes(page, pageSize uint64, sn string, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("alarm_type=? and serial_number=?", queryType, sn)
	}
}

func (adh *AlarmDbHandler) getAlarmOrEventInfoByAlarmInfoId(Id uint64) (*AlarmInfo, error) {
	var alarm AlarmInfo
	return &alarm, adh.db().Model(AlarmInfo{}).Where("id=?", Id).First(&alarm).Error
}

func (adh *AlarmDbHandler) listGroupAlarmsOrEventsDb(pageNum, pageSize uint64, sns []string, queryType string) (
	*[]AlarmInfo, error) {
	var alarmInfo []AlarmInfo
	return &alarmInfo, adh.db().Scopes(getAlarmGroupScopes(pageNum, pageSize, sns, queryType)).Find(&alarmInfo).Error
}

func getAlarmGroupScopes(page, pageSize uint64, sns []string, queryType string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("alarm_type=? and sn in (?)", queryType, sns)
	}
}
