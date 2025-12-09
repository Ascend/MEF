// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package database operation for meta db
package database

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"huawei.com/mindx/common/database"
)

var (
	repository     metaRepositoryImpl
	repositoryOnce sync.Once
)

// Meta metadata object
type Meta struct {
	Key   string `gorm:"column:key; size:256; primaryKey"`
	Type  string `gorm:"column:type; size:32"`
	Value string `gorm:"column:value; type:text"`
}

// MetaRepository operate meta table
type MetaRepository interface {
	GetByKey(key string) (Meta, error)
	GetByType(typ string) ([]Meta, error)
	GetKeyByType(typ string) ([]string, error)
	DeleteByKey(key string) error
	CreateOrUpdate(meta Meta) error
	CountByType(typ string) (int64, error)
}

// GetMetaRepository get meta repository instance
func GetMetaRepository() MetaRepository {
	repositoryOnce.Do(func() {
		repository = metaRepositoryImpl{db: database.GetDb()}
	})
	return &repository
}

// InitMetaRepository init meta repository
func InitMetaRepository() error {
	return database.CreateTableIfNotExist(&Meta{})
}

type metaRepositoryImpl struct {
	db *gorm.DB
}

// GetByKey get by key
func (m *metaRepositoryImpl) GetByKey(key string) (Meta, error) {
	var meta Meta
	return meta, m.db.Model(&Meta{}).Where(&Meta{Key: key}).First(&meta).Error
}

// GetByType get by type
func (m *metaRepositoryImpl) GetByType(typ string) ([]Meta, error) {
	var metas []Meta
	return metas, m.db.Model(&Meta{}).Where(&Meta{Type: typ}).Find(&metas).Error
}

// GetKeyByType get key by type
func (m *metaRepositoryImpl) GetKeyByType(typ string) ([]string, error) {
	var keys []string
	return keys, m.db.Model(&Meta{}).Select("key").Where(&Meta{Type: typ}).Find(&keys).Error
}

// DeleteByKey delete key
func (m *metaRepositoryImpl) DeleteByKey(key string) error {
	return m.db.Model(&Meta{}).Delete(&Meta{Key: key}).Error
}

// CreateOrUpdate create or update
func (m *metaRepositoryImpl) CreateOrUpdate(meta Meta) error {
	return m.db.Model(&Meta{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(&meta).Error
}

// CountByType get count of records by type
func (m *metaRepositoryImpl) CountByType(typ string) (int64, error) {
	var count int64
	return count, m.db.Model(&Meta{}).Where(&Meta{Type: typ}).Count(&count).Error
}
