// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certupdater cert update control module
package certupdater

import (
	"fmt"
	"reflect"
	"sync"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// table names and state definition for cert update operation
const (
	TableEdgeCaCertStatus  = "edge_ca_cert_status"
	TableEdgeSvcCertStatus = "edge_svc_cert_status"
	UpdateStatusInit       = 1
	UpdateStatusSuccess    = 2
	UpdateStatusFail       = 3
)

var dbRebuildLock sync.Mutex

// EdgeSvcCertStatusMod db model instance for edge svc cert state
var EdgeSvcCertStatusMod edgeSvcCertStatus

// EdgeCaCertStatusMod db model instance for edge root ca cert state
var EdgeCaCertStatusMod edgeCaCertStatus

// edgeCaCertStatus track each edge node status during root ca cert update
type edgeCaCertStatus struct {
	Id              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Sn              string `gorm:"column:sn;size:64;not null"`
	Ip              string `gorm:"column:ip;size:16;not null"`
	Status          int64  `gorm:"column:status"`
	NotifyTimestamp int64  `gorm:"column:notify_timestamp"`
}

// edgeSvcCertStatus track each edge node status during service cert update
type edgeSvcCertStatus struct {
	Id              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Sn              string `gorm:"column:sn;size:64;not null"`
	Ip              string `gorm:"column:ip;size:16;not null"`
	Status          int64  `gorm:"column:status"`
	NotifyTimestamp int64  `gorm:"column:notify_timestamp"`
}

// TableName define database table name for edge root ca state
func (*edgeCaCertStatus) TableName() string {
	return TableEdgeCaCertStatus
}

// TableName define database table name for edge service cert state
func (*edgeSvcCertStatus) TableName() string {
	return TableEdgeSvcCertStatus
}

// getEdgeSvcCertStatusModInstance get a model instance for edge service cert state
func getEdgeSvcCertStatusModInstance() *edgeSvcCertStatus {
	return &EdgeSvcCertStatusMod
}

// getEdgeCaCertStatusDbModInstance get a model instance for edge root ca cert state
func getEdgeCaCertStatusDbModInstance() *edgeCaCertStatus {
	return &EdgeCaCertStatusMod
}
func isValidModel(model interface{}) bool {
	if model == nil {
		return false
	}
	modType := reflect.TypeOf(model)
	if modType == reflect.TypeOf(edgeCaCertStatus{}) || modType == reflect.TypeOf(edgeSvcCertStatus{}) {
		return true
	}
	return false
}

// RebuildDBTable drop and create new db table when cert update starts
func RebuildDBTable(model interface{}) error {
	modTypeName := reflect.TypeOf(model).Name()
	if !isValidModel(model) {
		return fmt.Errorf("invalid database model: %v", modTypeName)
	}
	dbRebuildLock.Lock()
	defer dbRebuildLock.Unlock()
	if err := database.DropTableIfExist(model); err != nil {
		optErr := fmt.Errorf("drop table [%v] failed: %v", modTypeName, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	if err := database.CreateTableIfNotExist(model); err != nil {
		optErr := fmt.Errorf("create table [%v] failed: %v", modTypeName, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// DeleteDBTable delete db table when cert update is finished
func DeleteDBTable(model interface{}) error {
	modTypeName := reflect.TypeOf(model).Name()
	if !isValidModel(model) {
		return fmt.Errorf("invalid database model: %v", modTypeName)
	}
	if err := database.DropTableIfExist(model); err != nil {
		optErr := fmt.Errorf("drop table [%v] failed: %v", modTypeName, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// CreateOneRecord insert one record at one time
func (updater *edgeCaCertStatus) CreateOneRecord() error {
	if updater == nil {
		return fmt.Errorf("invalid edgeCaCertStatus instance")
	}
	if err := database.GetDb().Create(updater).Error; err != nil {
		optErr := fmt.Errorf("insert record(s) to table [%v] error: %v", TableEdgeCaCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// CreateMultipleRecords efficiently insert multiple data at one time
func (updater *edgeCaCertStatus) CreateMultipleRecords(data []edgeCaCertStatus) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeCaCertStatus instance")
	}
	if len(data) == 0 {
		return nil
	}
	if err := database.GetDb().Create(data).Error; err != nil {
		optErr := fmt.Errorf("insert record(s) to table [%v] error: %v", TableEdgeCaCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// UpdateRecordsBySns update multiple records by sn, new data is passed by edgeSvcCertStatus instance
func (updater *edgeCaCertStatus) UpdateRecordsBySns(sns []string, newData map[string]interface{}) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeCaCertStatus instance")
	}
	if len(newData) == 0 || len(sns) == 0 {
		return fmt.Errorf("empty data need to be updated")
	}
	if err := database.GetDb().Model(updater).Where("sn IN ?", sns).Updates(newData).Error; err != nil {
		optErr := fmt.Errorf("update record(s) in table [%v] error: %v", TableEdgeCaCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// DeleteRecordsBySns delete multiple records by sn at one time
func (updater *edgeCaCertStatus) DeleteRecordsBySns(sns []string) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeCaCertStatus instance")
	}
	if len(sns) == 0 {
		return nil
	}
	var deletedRecords []edgeCaCertStatus
	if err := database.GetDb().Model(updater).Where("sn IN ?", sns).Delete(&deletedRecords).Error; err != nil {
		optErr := fmt.Errorf("delete record(s) from table [%v] error: %v", TableEdgeCaCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// queryRecords query records by condition
func (updater *edgeCaCertStatus) queryRecords(cond, notCond map[string]interface{}) ([]edgeCaCertStatus, error) {
	if updater == nil {
		return nil, fmt.Errorf("invalid edgeCaCertStatus instance")
	}
	var records []edgeCaCertStatus
	db := database.GetDb().Limit(common.MaxNode)
	if len(cond) > 0 {
		db = db.Where(cond)
	}
	if len(notCond) > 0 {
		db = db.Not(notCond)
	}
	if err := db.Find(&records).Error; err != nil {
		optErr := fmt.Errorf("query record(s) from table [%v] error: %v", TableEdgeCaCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return nil, optErr
	}
	return records, nil
}

// QueryInitRecords query init state records, wrap function queryRecords
func (updater *edgeCaCertStatus) QueryInitRecords() ([]edgeCaCertStatus, error) {
	queryCond := map[string]interface{}{
		"status":           UpdateStatusInit,
		"notify_timestamp": 0,
	}
	return updater.queryRecords(queryCond, nil)
}

// QueryUnsuccessfulRecords query unsuccessful records, including init and failed records
func (updater *edgeCaCertStatus) QueryUnsuccessfulRecords() ([]edgeCaCertStatus, error) {
	notCond := map[string]interface{}{
		"status": UpdateStatusSuccess,
	}
	return updater.queryRecords(nil, notCond)
}

// CreateOneRecord insert one record at one time
func (updater *edgeSvcCertStatus) CreateOneRecord() error {
	if updater == nil {
		return fmt.Errorf("invalid edgeSvcCertStatus instance")
	}
	if err := database.GetDb().Create(updater).Error; err != nil {
		optErr := fmt.Errorf("insert record(s) to table [%v] error: %v", TableEdgeSvcCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// CreateMultipleRecords efficiently insert multiple data at one time
func (updater *edgeSvcCertStatus) CreateMultipleRecords(data []edgeSvcCertStatus) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeSvcCertStatus instance")
	}
	if len(data) == 0 {
		return nil
	}
	if err := database.GetDb().Create(data).Error; err != nil {
		optErr := fmt.Errorf("insert record(s) to table [%v] error: %v", TableEdgeSvcCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// UpdateRecordsBySns update multiple records by sn, new data is passed by edgeSvcCertStatus instance
func (updater *edgeSvcCertStatus) UpdateRecordsBySns(sns []string, newData map[string]interface{}) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeSvcCertStatus instance")
	}
	if len(newData) == 0 || len(sns) == 0 {
		return fmt.Errorf("empty data need to be updated")
	}
	if err := database.GetDb().Model(updater).Where("sn IN ?", sns).Updates(newData).Error; err != nil {
		optErr := fmt.Errorf("update record(s) in table [%v] error: %v", TableEdgeSvcCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// DeleteRecordsBySns delete multiple records by sn at one time
func (updater *edgeSvcCertStatus) DeleteRecordsBySns(sns []string) error {
	if updater == nil {
		return fmt.Errorf("invalid edgeSvcCertStatus instance")
	}
	if len(sns) == 0 {
		return nil
	}
	if err := database.GetDb().Model(updater).Where("sn IN ?", sns).Delete(nil).Error; err != nil {
		optErr := fmt.Errorf("delete record(s) from table [%v] error: %v", TableEdgeSvcCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return optErr
	}
	return nil
}

// queryRecords query records by condition
func (updater *edgeSvcCertStatus) queryRecords(cond, notCond map[string]interface{}) ([]edgeSvcCertStatus, error) {
	if updater == nil {
		return nil, fmt.Errorf("invalid edgeSvcCertStatus instance")
	}
	var records []edgeSvcCertStatus
	db := database.GetDb().Limit(common.MaxNode)
	if len(cond) > 0 {
		db = db.Where(cond)
	}
	if len(notCond) > 0 {
		db = db.Not(notCond)
	}
	if err := db.Find(&records).Error; err != nil {
		optErr := fmt.Errorf("query record(s) from table [%v] error: %v", TableEdgeSvcCertStatus, err)
		hwlog.RunLog.Error(optErr)
		return nil, optErr
	}
	return records, nil
}

// QueryInitRecords query init state records, wrap function queryRecords
func (updater *edgeSvcCertStatus) QueryInitRecords() ([]edgeSvcCertStatus, error) {
	queryCond := map[string]interface{}{
		"status":           UpdateStatusInit,
		"notify_timestamp": 0,
	}
	return updater.queryRecords(queryCond, nil)
}

// QueryUnsuccessfulRecords query unsuccessful records
func (updater *edgeSvcCertStatus) QueryUnsuccessfulRecords() ([]edgeSvcCertStatus, error) {
	notCond := map[string]interface{}{
		"status":           UpdateStatusSuccess,
		"notify_timestamp": 0,
	}
	return updater.queryRecords(nil, notCond)
}
