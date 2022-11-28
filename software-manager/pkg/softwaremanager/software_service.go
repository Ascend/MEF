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
	defaultFilesPath         = "/etc/mindx-edge/software-manager/"
	dBFileMode               = 0640
)

// SoftwareRecord is to define the struct of software record table
type softwareRecord struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	ContentType string    `gorm:"type:varchar(64);not null" json:"contentType"`
	Version     string    `gorm:"unique;type:varchar(64);not null" json:"version"`
	FileSize    float64   `gorm:"type:float(64);not null" json:"fileSize"`
}
type downloadData struct {
	URL      string `json:"url"`
	UserName string `json:"userName"`
	Password string `json:"password"`
	NodeID   string `json:"node_id"`
}
type queryResult struct {
	SoftwareRecords []softwareRecord `json:"softwareRecords"`
	Total           int64            `json:"total"`
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
	if !downloadRight(info.UserName, info.Password, info.NodeID) {
		hwlog.RunLog.Error("Wrong user or password")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Wrong user or password"}
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
	info, ok := input.(restfulservice.SoftwareInfo)
	if !ok {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	var softwareRecords []softwareRecord
	page, err := strconv.Atoi(info.Page)
	if err != nil {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	pageSize, err := strconv.Atoi(info.PageSize)
	if err != nil {
		hwlog.RunLog.Error("Class convert error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Class convert error"}
	}
	if page == 0 {
		page = common.DefaultPage
	}
	if pageSize > common.DefaultMaxPageSize {
		pageSize = common.DefaultMaxPageSize
	}
	offset := (page - 1) * pageSize
	if gormDB == nil {
		hwlog.RunLog.Error("gormDB is nil")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "gormDB is nil"}
	}
	var total int64
	if err := gormDB.Model(&softwareRecord{}).Count(&total).Error; err != nil {
		hwlog.RunLog.Error("Database query exception")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Database query exception"}
	}
	db := gormDB.Model(&softwareRecord{}).Offset(offset).Limit(pageSize).Find(&softwareRecords)
	if db.Error != nil {
		hwlog.RunLog.Error("Database error")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Database error"}
	}
	return common.RespMsg{Status: common.Success, Data: queryResult{softwareRecords, total}}
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
	if !checkNodeID(info.NodeID) {
		hwlog.RunLog.Error("Incorrect node_id")
		return common.RespMsg{Status: common.ErrorGetResponse, Msg: "Incorrect node_id"}
	}
	path := checkSoftwareExist(info.ContentType, info.Version)
	if path == "" {
		hwlog.RunLog.Error("Software dose not exist. Need to import software first")
		return common.RespMsg{Status: common.ErrorGetResponse,
			Msg: "Software dose not exist. Need to import software first"}
	}
	url := "Get " + "http://" + IP + ":" + strconv.Itoa(Port) + "/software-manager/v1/softwaremanager/?" +
		"contentType=" + info.ContentType + "&version=" + info.Version
	if userInfo := restfulservice.UserInfoMap[info.NodeID]; userInfo == nil {
		userName := usrgenerate()
		password := psdgenerate()
		restfulservice.UserInfoMap[info.NodeID] =
			map[string][]byte{restfulservice.UserName: userName, restfulservice.Password: password}
		downloadInfo := downloadData{url, string(userName), string(password), info.NodeID}
		return common.RespMsg{Status: common.Success, Data: downloadInfo}
	} else {
		downloadInfo := downloadData{url, string(userInfo[restfulservice.UserName]),
			string(userInfo[restfulservice.Password]), info.NodeID}
		return common.RespMsg{Status: common.Success, Data: downloadInfo}
	}
}
