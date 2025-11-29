// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package test for utils
package test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"huawei.com/mindx/common/hwlog"
)

var (
	dbInstance *gorm.DB
	// ErrTest common error for test
	ErrTest = errors.New("test error")
)

// RunWithPatches test main with patches, can be invoked in TestMain.
func RunWithPatches(tm TcModule, m *testing.M, patches *gomonkey.Patches) {
	if patches != nil {
		defer patches.Reset()
	}
	if err := tm.Setup(); err != nil {
		panic(err)
	}
	code := m.Run()
	tm.Teardown()
	fmt.Printf("exit_code = %d\n", code)
}

// InitLog init log
func InitLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitHwLogger(logConfig, logConfig); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

// InitDb init db
func InitDb(dbPath string, tables ...interface{}) error {
	var err error
	if err = os.Remove(dbPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("cleanup db failed, %v\n", err)
		return errors.New("cleanup db failed")
	}
	dbInstance, err = gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		fmt.Printf("init test db failed, %v\n", err)
		return errors.New("init test db failed")
	}

	if err = dbInstance.AutoMigrate(tables...); err != nil {
		fmt.Printf("setup table failed, %v\n", err)
		return errors.New("setup table failed")
	}
	return nil
}

// MockGetDb mock for func GetDb
func MockGetDb() *gorm.DB {
	return dbInstance
}

// RemoveDb remove db
func RemoveDb(dbPath string) error {
	if err := os.Remove(dbPath); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup [%s] failed, %v\n", dbPath, err)
		return fmt.Errorf("cleanup [%s] failed", dbPath)
	}
	return nil
}

// CloseDb close db connection
func CloseDb() error {
	if dbInstance == nil {
		return nil
	}
	sqlDb, err := dbInstance.DB()
	if err != nil {
		fmt.Printf("get sql db failed, %v\n", err)
		return errors.New("get sql db failed")
	}

	if err = sqlDb.Close(); err != nil {
		fmt.Printf("close db handle failed, %v\n", err)
		return errors.New("close db handle failed")
	}
	return nil
}
