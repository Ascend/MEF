// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const defaultDomainId = 0

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFuncReturn(fileutils.IsSoftLink, nil).
		ApplyMethodReturn(&kmc.ManualUpdateKmcTask{}, "RunTask", nil).
		ApplyFuncReturn(kmc.KeInitializeEx, kmc.Context{}, nil).
		ApplyMethodReturn(kmc.Context{}, "KeEncryptByDomainEx", []byte("test"), nil)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patches)
}

func TestUpdateKmc(t *testing.T) {
	convey.Convey("test UpdateKmcFlow", t, func() {
		convey.Convey("test entire Flow no key", doUpdateKmcEntireTest)
		convey.Convey("test entire Flow with key", doUpdateKmcWithKeyTest)
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
