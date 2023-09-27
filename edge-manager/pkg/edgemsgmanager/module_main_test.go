// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager module main test
package edgemsgmanager

import (
	"errors"
	"testing"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

var testErr = errors.New("test error")

func setup() {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
}

func teardown() {

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
}
