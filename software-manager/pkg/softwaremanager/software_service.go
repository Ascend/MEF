// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager this file is for business logic
package softwaremanager

import (
	"flag"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"

	"software-manager/pkg/restfulservice"
)

var (
	maximumSize float64
	// RepositoryFilesPath is to define the root path of imported software
	RepositoryFilesPath string
	gormDB              *gorm.DB
	// IP is the address of service
	IP string
	// Port is to define the service entrance
	Port int
)

const (
	kbToMB           float64 = 1048576
	defaultFilesPath         = "/etc/mindx-edge/software-manager/software-manager.db"
	dBFileMode               = 0640
)

// SoftwareRecord is to define the struct of software record table
type softwareRecord struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	ContentType string  `gorm:"type:varchar(64);not null"`
	Version     string  `gorm:"unique;type:varchar(64);not null"`
	FileSize    float64 `gorm:"type:float(64);not null"`
}

func init() {
	flag.StringVar(&RepositoryFilesPath, "repositoryFilesPath", defaultFilesPath,
		"The path of repository")
	flag.Float64Var(&maximumSize, "maximumSize", 0,
		"The path of repository")
	flag.StringVar(&IP, "ip", "",
		"The listen ip of the service,0.0.0.0 is not recommended when install on Multi-NIC host")
	flag.IntVar(&Port, "port", 0, "The server port of the http service,range[1025-40000]")
}

// InitDatabase is to init the database table for repository
func InitDatabase(repositoryFilesPath string) error {
	relPath, err := filepath.Abs(filepath.Join(repositoryFilesPath, "/repository.db"))
	if err != nil {
		hwlog.RunLog.Error("File path standardization fail")
		return err
	}
	db, err := gorm.Open(sqlite.Open(relPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		hwlog.RunLog.Error("Init database connection failed")
		return err
	}
	if err = os.Chmod(relPath, dBFileMode); err != nil {
		hwlog.RunLog.Error("Chmod for database file error")
		return err
	}
	err = db.AutoMigrate(&softwareRecord{})
	if err != nil {
		hwlog.RunLog.Error("Migrate database error")
		return err
	}
	gormDB = db
	return nil
}
func deleteSoftware(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	hwlog.RunLog.Info("Start delete software")
	path := checkSoftwareExist(info.ContentType, info.Version)
	if path == "" {
		hwlog.RunLog.Error("Software dose not exist")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Software dose not exist"}
	}
	err := os.Remove(path)
	if err != nil {
		hwlog.RunLog.Error("Delete file error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Delete file error"}
	} else {
		if gormDB == nil {
			hwlog.RunLog.Error("gormDB is nil")
			return common.RespMsg{Status: common.ErrorGetResponse, Msg: "gormDB is nil"}
		}
		gormDB.Where("content_type=? and version=?", info.ContentType, info.Version).
			Unscoped().Delete(&softwareRecord{})
		return common.RespMsg{Status: common.Success}

	}
}

func downloadSoftware(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	if !checkFields(info.ContentType, info.Version) {
		hwlog.RunLog.Error("Incorrect fields")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Incorrect fields"}
	}
	if path := checkSoftwareExist(info.ContentType, info.Version); path != "" {
		return common.RespMsg{Status: common.Success, Data: path}
	} else {
		hwlog.RunLog.Error("Software dose not exist")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Software dose not exist"}
	}
}

func uploadSoftware(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	if !checkFields(info.ContentType, info.Version) {
		hwlog.RunLog.Error("Incorrect fields")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Incorrect fields"}
	}
	if path := checkSoftwareExist(info.ContentType, info.Version); path != "" {
		hwlog.RunLog.Error("Software already exists")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Software already exists"}
	}
	file := info.File
	if !checkFile(file) {
		hwlog.RunLog.Error("Wrong file format")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Wrong file format"}
	}
	dst := creatDir(info.ContentType, info.Version)
	err := saveUploadedFile(file, dst+"/"+info.ContentType+".zip")
	if err != nil {
		hwlog.RunLog.Error("Save file error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Save file error"}
	} else {
		if gormDB == nil {
			hwlog.RunLog.Error("gormDB is nil")
			return common.RespMsg{Status: common.ErrorGetResponse, Msg: "gormDB is nil"}
		}
		gormDB.Create(&softwareRecord{
			ContentType: info.ContentType,
			Version:     info.Version,
			FileSize:    float64(file.Size) / kbToMB,
		})
		return common.RespMsg{Status: common.Success}
	}
}

func getRepository(input interface{}) common.RespMsg {
	var softwareRecords []softwareRecord
	if gormDB == nil {
		hwlog.RunLog.Error("gormDB is nil")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "gormDB is nil"}
	}
	db := gormDB.Find(&softwareRecords)
	if db.Error != nil {
		hwlog.RunLog.Error("Database error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Database error"}
	}
	return common.RespMsg{Status: common.Success, Data: softwareRecords}
}

func getURL(input interface{}) common.RespMsg {
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	if !checkFields(info.ContentType, info.Version) {
		hwlog.RunLog.Error("Incorrect fields")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Incorrect fields"}
	}
	url := IP + ":" + strconv.Itoa(Port) + "/software-manager/v1/softwaremanager/?" +
		"contentType=" + info.ContentType + "&version=" + info.Version
	return common.RespMsg{Status: common.Success, Data: url}
}
