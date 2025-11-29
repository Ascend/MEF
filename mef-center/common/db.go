// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common for database
package common

import (
	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
)

// Paginate slice page
func Paginate(page, pageSize uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = DefaultPage
		}
		if pageSize > DefaultMaxPageSize {
			pageSize = DefaultMaxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}

// GetItemCount get item count in table
func GetItemCount(table interface{}) (int, error) {
	var total int64
	if err := database.GetDb().Model(table).Count(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}
