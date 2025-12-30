// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package handlermgr test for get net config
package handlermgr

import (
	"encoding/json"
	"fmt"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/kmc"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

const testToken = "ABCDEFG"

func testGetNetConfig() {
	netConfig := &config.NetManager{
		NetType: constants.MEF,
		IP:      "127.0.0.1",
		Port:    10000,
		WithOm:  false,
		Token:   []byte(testToken),
	}
	var p1 = gomonkey.ApplyFuncReturn(config.GetNetManager, netConfig, nil).
		ApplyFuncReturn(util.GetKmcConfig, nil, nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil)
	defer p1.Reset()

	bytes, err := json.Marshal(netConfig)
	if err != nil {
		fmt.Printf("marshal data failed: %v\n", err)
		return
	}

	mgr := config.NewDbMgr("./", "test.db")
	res := getNetConfig(mgr)
	convey.So(res, convey.ShouldResemble, string(bytes))

	var p2 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p2.Reset()
	res = getNetConfig(mgr)
	convey.So(res, convey.ShouldResemble, constants.Failed)
}

func testDecryptTokenForSdk() {
	testErrNetType()
	testErrGetInstallRootDir()
	testErrGetKmcCfg()
	testErrDecryptContent()
}

func testErrNetType() {
	netConfig := &config.NetManager{
		NetType: constants.FD,
	}
	err := decryptToken(netConfig)
	convey.So(err, convey.ShouldBeNil)
}

func testErrGetInstallRootDir() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr)
	defer p1.Reset()

	netConfig := &config.NetManager{
		NetType: constants.MEF,
	}
	err := decryptToken(netConfig)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testErrGetKmcCfg() {
	var p1 = gomonkey.ApplyFunc(util.GetKmcConfig,
		func(kmcDir string) (*kmc.SubConfig, error) {
			return nil, testErr
		})
	defer p1.Reset()

	netConfig := &config.NetManager{
		NetType: constants.MEF,
	}
	err := decryptToken(netConfig)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testErrDecryptContent() {
	var p1 = gomonkey.ApplyFuncReturn(util.GetKmcConfig, nil, nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), testErr)
	defer p1.Reset()

	netConfig := &config.NetManager{
		NetType: constants.MEF,
	}
	err := decryptToken(netConfig)
	convey.So(err, convey.ShouldResemble, testErr)
}
