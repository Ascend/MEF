// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK || MEFEdge_A500

// Package handlermgr test for get net config
package handlermgr

import (
	"encoding/json"
	"fmt"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

func testGetNetConfig() {
	netConfig := &config.NetManager{
		NetType: constants.FD,
		WithOm:  true,
	}
	var p1 = gomonkey.ApplyFunc(config.GetNetManager,
		func(dbMgr *config.DbMgr) (*config.NetManager, error) {
			return netConfig, nil
		})
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
	return
}
