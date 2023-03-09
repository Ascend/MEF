// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager module main test
package softwaremanager

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
)

var (
	gormInstance *gorm.DB
	dbPath       = "./test.db"
)

func setup() {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}

	if err = os.Remove(dbPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		hwlog.RunLog.Errorf("cleanup db failed, error: %v", err)
	}
	gormInstance, err = gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		hwlog.RunLog.Errorf("failed to init test db, %v", err)
	}

	if err = gormInstance.AutoMigrate(&SoftwareInfo{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}

}

func teardown() {

}

func mockGetDb() *gorm.DB {
	return gormInstance
}

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFunc(database.GetDb, mockGetDb)
	defer patches.Reset()

	var encPatch = gomonkey.ApplyFunc(common.EncryptContent, func(content []byte, kmcCfg *common.KmcCfg) ([]byte,
		error) {
		return []byte{1, 2, 3}, nil
	})
	defer encPatch.Reset()

	var decryptPatch = gomonkey.ApplyFunc(common.DecryptContent, func(encryptByte []byte, kmcCfg *common.KmcCfg) ([]byte,
		error) {
		return []byte{1, 2, 3}, nil
	})
	defer decryptPatch.Reset()

	setup()
	code := m.Run()
	teardown()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
}
