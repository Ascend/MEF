// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package restful test for init.go
package restful

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
)

var service *Service

const (
	testIp             = "127.0.0.1"
	testPort           = 30001
	defaultConnection  = 3
	defaultConcurrency = 3
	defaultDataLimit   = 1024 * 1024
	defaultLimitIPReq  = "2/1"
	defaultCacheSize   = 1024 * 1024 * 10
)

func TestAlarmManager(t *testing.T) {
	httpsServerConf := &httpsmgr.HttpsServer{
		IP:          testIp,
		Port:        testPort,
		SwitchLimit: true,
		ServerParam: getLimitParam(),
		TlsCertPath: certutils.TlsCertInfo{
			RootCaPath: utils.RootCaPath,
			CertPath:   utils.ServerCertPath,
			KeyPath:    utils.ServerKeyPath,
			SvrFlag:    true,
			KmcCfg:     nil,
			WithBackup: true,
		},
	}
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(NewRestfulService(true, httpsServerConf)); err != nil {
		panic(err)
	}
	service = &Service{
		enable:   true,
		httpsSvr: httpsServerConf,
	}
	convey.Convey("test Service method 'NewRestfulService', 'Name', 'Enable'", t, testService)
	convey.Convey("test Service method 'Start'", t, testServiceStart)
}

func getLimitParam() httpsmgr.ServerParam {
	return httpsmgr.ServerParam{
		Concurrency:    defaultConcurrency,
		BodySizeLimit:  defaultDataLimit,
		LimitIPReq:     defaultLimitIPReq,
		LimitIPConn:    defaultConcurrency,
		LimitTotalConn: defaultConnection,
		CacheSize:      defaultCacheSize,
	}
}

func testService() {
	if service == nil {
		panic("restful service is nil")
	}
	convey.So(service.Name(), convey.ShouldEqual, common.RestfulServiceName)
	convey.So(service.Enable(), convey.ShouldBeTrue)
}

func testServiceStart() {
	var p1 = gomonkey.ApplyMethodSeq(&httpsmgr.HttpsServer{}, "Init",
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}, Times: 1},
			{Values: gomonkey.Params{test.ErrTest}},
			{Values: gomonkey.Params{nil}, Times: 2},
		}).
		ApplyMethodSeq(&httpsmgr.HttpsServer{}, "RegisterRoutes",
			[]gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}, Times: 1},
				{Values: gomonkey.Params{test.ErrTest}},
				{Values: gomonkey.Params{nil}, Times: 1},
			}).
		ApplyMethodSeq(&httpsmgr.HttpsServer{}, "Start",
			[]gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}, Times: 1},
				{Values: gomonkey.Params{test.ErrTest}},
			})
	defer p1.Reset()

	if service == nil {
		panic("restful service is nil")
	}
	service.Start()
	service.Start()
	service.Start()
	service.Start()
}
