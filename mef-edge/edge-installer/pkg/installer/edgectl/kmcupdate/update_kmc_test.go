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
	"huawei.com/mindx/common/test"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/kmc"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

const defaultDomainId = 0

var configPathMgr *pathmgr.ConfigPathMgr

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestUpdateKmc(t *testing.T) {
	configPathMgr = pathmgr.NewConfigPathMgr("./")

	convey.Convey("test UpdateKmcFlow", t, func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsSoftLink, nil)
		convey.Convey("test entire Flow no key", doUpdateKmcEntireTest)
		convey.Convey("test entire Flow with key", doUpdateKmcWithKeyTest)
		defer p1.Reset()
	})
}

func doUpdateKmcEntireTest() {
	flow := NewUpdateKmcFlow(configPathMgr)
	err := flow.RunFlow()
	convey.So(err, convey.ShouldBeNil)

	if configPathMgr == nil {
		panic("config path manager is nil")
	}
	err = os.RemoveAll(configPathMgr.GetConfigDir())
	convey.So(err, convey.ShouldBeNil)
}

func doUpdateKmcWithKeyTest() {
	if configPathMgr == nil {
		panic("config path manager is nil")
	}
	coreKmcPath := configPathMgr.GetCompKmcDir(constants.EdgeCore)
	fmt.Println(coreKmcPath)
	ctx, err := initKmcCtx(coreKmcPath)
	convey.So(err, convey.ShouldBeNil)

	keyPath := configPathMgr.GetCompInnerSvrKeyPath(constants.EdgeCore)

	err = prepareKey(keyPath, ctx)
	convey.So(err, convey.ShouldBeNil)

	err = ctx.KeFinalizeEx()
	convey.So(err, convey.ShouldBeNil)

	preSha256Sum, err := getSha256Sum(keyPath)
	convey.So(err, convey.ShouldBeNil)

	flow := NewUpdateKmcFlow(configPathMgr)
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

	if err = os.WriteFile(keyPath, cipherData, fileutils.Mode600); err != nil {
		return fmt.Errorf("write cipher data failed: %s", err.Error())
	}

	return nil
}

func initKmcCtx(curPath string) (kmc.Context, error) {
	kmcCfg, err := util.GetKmcConfig(curPath)
	if err != nil {
		fmt.Printf("get kmc cfg failed: %s\n", err.Error())
		return kmc.Context{}, errors.New("get kmc cfg failed")
	}

	config := kmc.NewKmcInitConfig()
	config.PrimaryKeyStoreFile = kmcCfg.PrimaryKeyPath
	config.StandbyKeyStoreFile = kmcCfg.StandbyKeyPath
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
