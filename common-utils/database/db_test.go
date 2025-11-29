// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package database for
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"huawei.com/mindx/common/hwlog"
)

var (
	logConfig *hwlog.LogConfig
	dbPath    string
	testErr   = errors.New("test error")
)

type testTable struct {
	ID   uint64 `gorm:"type:Integer;primaryKey"`
	Name string `gorm:"type:char(128);unique;not null"`
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	fmt.Printf("exit_code = %d\n", code)
}

func setup() {
	logFile, err := filepath.Abs("./test_log")
	if err != nil {
		fmt.Printf("get log abs path failed: %v", err)
		return
	}
	logConfig = &hwlog.LogConfig{
		OnlyToFile:  true,
		LogFileName: logFile,
		MaxBackups:  hwlog.DefaultMaxBackups,
		MaxAge:      hwlog.DefaultMinSaveAge,
	}
	if err := hwlog.InitHwLogger(logConfig, logConfig); err != nil {
		fmt.Printf("init hwlog failed, error: %v\n", err)
		return
	}
	dbPathFile, err := os.CreateTemp("", "test.db")
	if err != nil {
		fmt.Printf("create file in temp dir failed, error: %v\n", err)
		return
	}
	dbPath, err = filepath.Abs(dbPathFile.Name())
	if err != nil {
		fmt.Printf("get abs dbPath failed: %s", err.Error())
		return
	}
}

func teardown() {
	if err := os.Remove(dbPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("cleanup db failed, error: %v\n", err)
		return
	}
}

func TestInitDB(t *testing.T) {
	convey.Convey("init db should be success", t, testInitDB)
	convey.Convey("init db should be failed, dbPath is invalid", t, testInitDBErrDbPath)
}

func testInitDB() {
	err := InitDB(dbPath)
	convey.So(err, convey.ShouldBeNil)
}

func testInitDBErrDbPath() {
	err := InitDB("./test.db")
	convey.So(err, convey.ShouldResemble, errors.New("check db path error"))
}

func TestInitDBConn(t *testing.T) {
	convey.Convey("init db conn should be failed, initialize db session failed", t, testInitDBConnErrInitDBSession)
	convey.Convey("init db conn should be failed, set max conn num failed", t, testInitDBConnErrSetMaxConnNum)
}

func testInitDBConnErrInitDBSession() {
	var p1 = gomonkey.ApplyFuncReturn(gorm.Open, nil, testErr)
	defer p1.Reset()
	dbConn := InitDbConnection(dbPath)
	convey.So(dbConn, convey.ShouldBeNil)
}

func testInitDBConnErrSetMaxConnNum() {
	var db *gorm.DB
	var p1 = gomonkey.ApplyFuncReturn(gorm.Open, db, nil)
	defer p1.Reset()
	var p2 = gomonkey.ApplyMethod(reflect.TypeOf(db), "DB",
		func(*gorm.DB) (*sql.DB, error) {
			return nil, testErr
		})
	defer p2.Reset()
	dbConn := InitDbConnection(dbPath)
	convey.So(dbConn, convey.ShouldEqual, db)
}

func TestCreateTableIfNotExist(t *testing.T) {
	convey.Convey("create table should be success", t, testCreateTable)
	convey.Convey("create table should be failed, gorm db is nil", t, testCreateTableReturnErrGormDBIsNil)
}

func testCreateTable() {
	var err error
	gormDB, err = gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		fmt.Printf("init test db failed, error: %v\n", err)
		return
	}
	err = CreateTableIfNotExist(testTable{})
	convey.So(err, convey.ShouldBeNil)
	convey.So(GetDb(), convey.ShouldEqual, gormDB)
}

func testCreateTableReturnErrGormDBIsNil() {
	gormDB = nil
	err := CreateTableIfNotExist(testTable{})
	convey.So(err, convey.ShouldResemble, errors.New("create table failed"))
}
