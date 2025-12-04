// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package domainconfig
package domainconfig

import (
	"bufio"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
)

const (
	testDomain        = "fd.fusion.huawei.com"
	testInvalidDomain = "fd.fusion_director.huawei.com"
	testIp            = "192.168.0.1"
	testExistIp       = "192.168.0.2"
	testInvalidIp     = "255.255.255.255"
	testHosts         = `127.0.0.1   Euler localhost localhost.localdomain localhost4 localhost4.localdomain4
						::1         Euler localhost localhost.localdomain localhost6 localhost6.localdomain6
						192.168.0.2 fd.fusion.huawei.com`
)

var expectErr = errors.New("import domain config failed")

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func TestDomainCfgFlow(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{}, nil)
	defer p.Reset()

	convey.Convey("test domain config successful", t, domainCfgSuccess)
	convey.Convey("test domain config failed", t, func() {
		convey.Convey("check param domain failed", checkParamDomainFailed)
		convey.Convey("check param ip failed", checkParamIPFailed)
		convey.Convey("import domain config failed", func() {
			convey.Convey("get domain config failed", getDomainCfgFailed)
			convey.Convey("create domain config failed", createDomainCfgFailed)
			convey.Convey("delete domain config in file failed", DeleteDomainCfgInFileFailed)
			convey.Convey("get modified hosts file failed", getModifiedHostsFileFailed)
			convey.Convey("overwrite hosts file failed", overwriteHostsFileFailed)
			convey.Convey("add domain config to file failed", addDomainCfgToFileFailed)
			convey.Convey("update domain config failed", updateDomainCfgFailed)
		})
	})
}

func domainCfgSuccess() {
	convey.Convey("duplicate domain and ip, not need to import", func() {
		p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, &config.DomainConfigs{
			Configs: []config.DomainConfig{
				{Domain: testDomain, IP: testIp},
			}}, nil)
		defer p.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("update domain config success", func() {
		p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, &config.DomainConfigs{
			Configs: []config.DomainConfig{
				{Domain: testDomain, IP: testExistIp},
			}}, nil).
			ApplyFuncSeq(DeleteDomainCfgInFile, []gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}},
				{Values: gomonkey.Params{nil}},
			}).
			ApplyFuncReturn(addDomainCfgToFile, nil).
			ApplyFuncReturn(config.SetDomainCfg, nil)
		defer p.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldBeNil)
	})
}

func checkParamDomainFailed() {
	domainCfg := NewDomainCfgFlow(testInvalidDomain, testIp)
	err := domainCfg.RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func checkParamIPFailed() {
	domainCfg := NewDomainCfgFlow(testDomain, testInvalidIp)
	err := domainCfg.RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func getDomainCfgFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, test.ErrTest)
	defer p.Reset()
	err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func createDomainCfgFailed() {
	convey.Convey("set domain config to db failed", func() {
		p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
			ApplyFuncReturn(config.SetDomainCfg, test.ErrTest)
		defer p.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create hosts file failed", func() {
		p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
			ApplyFuncReturn(config.SetDomainCfg, nil).
			ApplyFuncReturn(fileutils.IsExist, false).
			ApplyFuncReturn(os.OpenFile, nil, test.ErrTest)
		defer p.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func DeleteDomainCfgInFileFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
		ApplyFuncReturn(config.SetDomainCfg, nil).
		ApplyFuncSeq(fileutils.IsExist, []gomonkey.OutputCell{
			{Values: gomonkey.Params{true}},
			{Values: gomonkey.Params{true}},
		})
	defer p.Reset()

	convey.Convey("check hosts file permission failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p1.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check hosts file symlink failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", nil)
		defer p2.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("load hosts file failed", func() {
		p3 := gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, hostsFilePath, nil).
			ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p3.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func getModifiedHostsFileFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
		ApplyFuncReturn(config.SetDomainCfg, nil).
		ApplyFuncSeq(fileutils.IsExist, []gomonkey.OutputCell{
			{Values: gomonkey.Params{true}},
			{Values: gomonkey.Params{true}},
		}).
		ApplyFuncReturn(fileutils.CheckOwnerAndPermission, hostsFilePath, nil).
		ApplyFuncReturn(fileutils.LoadFile, []byte(testHosts), nil).
		ApplyMethodReturn(&bufio.Reader{}, "ReadString", "", test.ErrTest)
	defer p.Reset()
	err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}

func overwriteHostsFileFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
		ApplyFuncReturn(config.SetDomainCfg, nil).
		ApplyFuncSeq(fileutils.IsExist, []gomonkey.OutputCell{
			{Values: gomonkey.Params{true}},
			{Values: gomonkey.Params{true}},
		}).
		ApplyFuncReturn(fileutils.CheckOwnerAndPermission, hostsFilePath, nil).
		ApplyFuncReturn(fileutils.LoadFile, []byte(testHosts), nil)
	defer p.Reset()

	convey.Convey("open swap file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(os.OpenFile, nil, test.ErrTest)
		defer p1.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("lock swap file failed", func() {
		p2 := gomonkey.ApplyFuncReturn(syscall.Flock, test.ErrTest)
		defer p2.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func addDomainCfgToFileFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, nil, gorm.ErrRecordNotFound).
		ApplyFuncReturn(config.SetDomainCfg, nil).
		ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(createHostsFile, nil)
	defer p.Reset()

	convey.Convey("check hosts file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(checkHostsFile, test.ErrTest)
		defer p1.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("load /etc/hosts failed", func() {
		p2 := gomonkey.ApplyFuncReturn(checkHostsFile, nil).
			ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p2.Reset()
		err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func updateDomainCfgFailed() {
	p := gomonkey.ApplyFuncReturn(config.GetDomainCfg, &config.DomainConfigs{
		Configs: []config.DomainConfig{
			{Domain: testDomain, IP: testExistIp},
		}}, nil).
		ApplyFuncSeq(DeleteDomainCfgInFile, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{nil}},
		}).
		ApplyFuncReturn(addDomainCfgToFile, test.ErrTest)
	defer p.Reset()
	err := NewDomainCfgFlow(testDomain, testIp).RunTasks()
	convey.So(err, convey.ShouldResemble, expectErr)
}
