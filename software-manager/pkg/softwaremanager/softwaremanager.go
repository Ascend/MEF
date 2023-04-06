// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"software-manager/pkg/restfulservice"

	"gorm.io/gorm"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

const (
	edgeCore            = "edgecore"
	edgeInstaller       = "edge-installer"
	mefEdge             = "MEFEdge"
	userLength          = 8
	maxLength           = 20
	randomSet           = "0123456789abcdefghijklkmnopqrstuvwxyzABCDEFGHIJKLKMNOPQRSTUVWXYZ!@#$%^&*()-.+=`~"
	regexExp            = "^[\\w]+"
	stringLength        = 2
	maxByteLength       = 100
	maxExtractFileCount = 100
	hexTag              = 0xFF
)

const (
	kbToMB           float64 = 1048576
	defaultFilesPath         = "/home/data/config/"
	dBFileMode               = 0640
	floatByteSize            = 64
	zipFileHeader            = "504b0304"
)

// SoftwareRecord is to define the struct of software record table
type softwareRecord struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	CreatedAt   string  `gorm:"not null" json:"createdAt"`
	ContentType string  `gorm:"type:varchar(64);not null" json:"contentType"`
	Version     string  `gorm:"type:varchar(64);not null" json:"version"`
	FileSize    float64 `gorm:"type:float(64);not null" json:"fileSize"`
	Description string  `gorm:"type:varchar(64);not null" json:"description"`
}

type downloadData struct {
	URL      string `json:"url"`
	UserName string `json:"userName"`
	Password string `json:"password"`
	NodeID   string `json:"nodeID"`
}
type queryResult struct {
	SoftwareRecords []softwareRecord `json:"softwareRecords"`
	Total           int64            `json:"total"`
}
type batchDeleteResult struct {
	NotDeleteID []int `json:"deleteFail"`
}

// SoftwareDbCtl is the interface of database operation
type SoftwareDbCtl interface {
	addSoftware(info *restfulservice.SoftwareInfo) error
	listSoftware(info *restfulservice.SoftwareInfo) (*[]softwareRecord, int64, error)
	deleteSoftware(id int, notDeleteId *[]int) error
	querySoftware(contentType string, version string) (*softwareRecord, error)
	queryLaSoftware(contentType string) (*softwareRecord, error)
	querySoftwareByID(ID int) (*softwareRecord, error)
}

type softwareDbCtlImpl struct {
	db *gorm.DB
}

func (dbCtl *softwareDbCtlImpl) addSoftware(info *restfulservice.SoftwareInfo) error {
	fileSize, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(info.File.Size)/kbToMB), floatByteSize)
	if err != nil {
		hwlog.RunLog.Error("float truncate error in adAdd func")
		return errors.New("float truncate error")
	}
	db := dbCtl.db.Create(&softwareRecord{
		ContentType: info.ContentType,
		Version:     info.Version,
		FileSize:    fileSize,
		CreatedAt:   time.Now().Format(common.TimeFormat),
		Description: info.Description,
	})
	if db.Error != nil {
		return errors.New("database create error")
	}
	return nil
}

func (dbCtl *softwareDbCtlImpl) listSoftware(info *restfulservice.SoftwareInfo) (*[]softwareRecord, int64, error) {
	var total int64
	var softwareRecords []softwareRecord
	if info.Page == 0 {
		info.Page = common.DefaultPage
	}
	if info.PageSize > common.DefaultMaxPageSize {
		info.PageSize = common.DefaultMaxPageSize
	}
	offset := (info.Page - 1) * info.PageSize
	if err := dbCtl.db.Model(&softwareRecord{}).Count(&total).Error; err != nil {
		return &softwareRecords, total, errors.New("database query error")
	}
	db := dbCtl.db.Model(&softwareRecord{}).Offset(offset).Limit(info.PageSize).Find(&softwareRecords)
	if db.Error != nil {
		return &softwareRecords, total, errors.New("database query error")
	}
	return &softwareRecords, total, nil
}

func (dbCtl *softwareDbCtlImpl) deleteSoftware(id int, notDeleteId *[]int) error {
	db := dbCtl.db.Where("id=?", id).Unscoped().Delete(&softwareRecord{})
	if db.Error != nil {
		*notDeleteId = append(*notDeleteId, id)
		return errors.New("database delete error")
	}
	return nil
}

func (dbCtl *softwareDbCtlImpl) querySoftware(contentType string, version string) (*softwareRecord, error) {
	var records []softwareRecord
	db := dbCtl.db.Where("content_type=? and version=?", contentType, version).Find(&records)
	if db.Error != nil {
		return nil, errors.New("query database error in querySoftware func")
	}
	if len(records) == 0 {
		return nil, nil
	}
	if len(records) > 1 {
		return nil, fmt.Errorf("%s%s has %d records", contentType, version, len(records))
	}
	return &records[0], nil
}

func (dbCtl *softwareDbCtlImpl) queryLaSoftware(contentType string) (*softwareRecord, error) {
	var records []softwareRecord
	db := dbCtl.db.Where("content_type=?", contentType).Order("id desc").Limit(1).Find(&records)
	if db.Error != nil {
		return nil, errors.New("query database error in returnLatestVer func")
	}
	if len(records) != 1 {
		return nil, nil
	}
	return &records[0], nil
}

func (dbCtl *softwareDbCtlImpl) querySoftwareByID(id int) (*softwareRecord, error) {
	var records []softwareRecord
	db := dbCtl.db.Where("id=?", id).Find(&records)
	if db.Error != nil {
		return nil, errors.New("query database error in returnLatestVer func")
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}
