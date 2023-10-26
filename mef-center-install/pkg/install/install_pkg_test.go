// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

var ErrTest = errors.New("test Err")

func ResetAndClearDir(p *Patches, clearPath string) {
	if p != nil {
		p.Reset()
	}
	err := fileutils.DeleteAllFileWithConfusion(clearPath)
	So(err, ShouldBeNil)
}

// Environment is a struct used on UT to control the environment
type Environment struct {
}

// Setup is used to init the UT environment
func (e *Environment) Setup() error {
	logFile := "./test_log"
	logConfig := &hwlog.LogConfig{
		OnlyToFile:  true,
		LogFileName: logFile,
		MaxBackups:  hwlog.DefaultMaxBackups,
		MaxAge:      hwlog.DefaultMinSaveAge,
	}
	if err := common.InitHwlogger(logConfig, logConfig); err != nil {
		return err
	}
	return nil
}

// Teardown is used to clear the UT environment
func (e *Environment) Teardown() {}

func TestMain(m *testing.M) {
	env := Environment{}
	if err := env.Setup(); err != nil {
		fmt.Printf("failed to setup test environment, reason: %v", err)
		return
	}
	defer env.Teardown()
	code := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", code)
}

func TestInstallPkg(t *testing.T) {
	Convey("test install pkg", t, func() {
		Convey("test install mgr file", DoInstallMgrTest)
		Convey("test cert mgr file", CertMgrTest)
		Convey("test working dir mgr file", WorkingDirMgrTest)
		Convey("test yaml mgr file", YamlMgrTest)
	})
}
