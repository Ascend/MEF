// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package kmcupdate

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const defaultDomainId = 0

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

func TestUpdateKmc(t *testing.T) {
	convey.Convey("test UpdateKmcFlow", t, func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsSoftLink, nil)
		convey.Convey("test entire Flow no key", doUpdateKmcEntireTest)
		convey.Convey("test entire Flow with key", doUpdateKmcWithKeyTest)
		defer p1.Reset()
	})
}

func doUpdateKmcEntireTest() {
	pathMgr, err := util.InitInstallDirPathMgr()
	convey.So(err, convey.ShouldBeNil)
	flow := NewUpdateKmcFlow(pathMgr)
	err = flow.RunFlow()
	convey.So(err, convey.ShouldBeNil)
}

func doUpdateKmcWithKeyTest() {
	pathMgr, err := util.InitInstallDirPathMgr()
	convey.So(err, convey.ShouldBeNil)
	ctx, err := initKmcCtx(pathMgr)
	convey.So(err, convey.ShouldBeNil)

	keyPath := pathMgr.ConfigPathMgr.GetRootCaKeyPath()
	err = prepareKey(keyPath, ctx)
	convey.So(err, convey.ShouldBeNil)

	preSha256Sum, err := getSha256Sum(keyPath)
	convey.So(err, convey.ShouldBeNil)

	flow := NewUpdateKmcFlow(pathMgr)
	err = flow.RunFlow()
	convey.So(err, convey.ShouldBeNil)

	postSha256Sum, err := getSha256Sum(keyPath)
	convey.So(err, convey.ShouldBeNil)

	convey.So(postSha256Sum, convey.ShouldNotEqual, preSha256Sum)
}

func prepareKey(keyPath string, ctx kmc.Context) error {

	err := fileutils.MakeSureDir(keyPath)
	if err != nil {
		return err
	}

	plainData := []byte("key")
	cipherData, err := ctx.KeEncryptByDomainEx(defaultDomainId, plainData)
	if err != nil {
		return fmt.Errorf("encrypt data failed: %s", err.Error())
	}

	if err := fileutils.WriteData(keyPath, cipherData); err != nil {
		return fmt.Errorf("write cipher data failed: %s", err.Error())
	}

	return nil
}

func initKmcCtx(pathMgr *util.InstallDirPathMgr) (kmc.Context, error) {
	kmcKeyPath := pathMgr.ConfigPathMgr.GetRootMasterKmcPath()
	kmcBackKeyPath := pathMgr.ConfigPathMgr.GetRootBackKmcPath()
	kmcCfg := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)

	config := kmc.NewKmcInitConfig()
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.SdpAlgId = kmcCfg.SdpAlgID
	c, err := kmc.KeInitializeEx(config)
	if err != nil {
		fmt.Printf("Init kmc failed: %s\n", err.Error())
		return kmc.Context{}, errors.New("init kmc failed")
	}

	return c, nil
}

func getSha256Sum(keyFile string) ([]byte, error) {
	file, err := os.Open(keyFile)
	if err != nil {
		return nil, fmt.Errorf("open key file failed: %s", err.Error())
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("get file hash failed: %s", err.Error())
	}

	return hash.Sum(nil), nil
}
