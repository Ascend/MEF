// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configchecker
package configchecker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	validPort   = 443
	invalidPort = 65536
)

type testImageConfig struct {
	Domain   string `json:"domain"`
	IP       string `json:"ip"`
	Port     int64  `json:"port"`
	Account  string `json:"account"`
	Password []byte `json:"password"`
}

func getImageCfgTemplate() testImageConfig {
	return testImageConfig{
		Domain:   "fd.Test-123.com",
		IP:       "127.0.0.1",
		Port:     validPort,
		Account:  "test-Account_123",
		Password: []byte{116, 101, 115, 116, 49, 50, 51, 239, 188, 154, 38, 124, 52, 53, 54, 194, 183, 96},
	}
}

func TestImageConfig(t *testing.T) {
	convey.Convey("test valid image config", t, testValidImageCfg)
	convey.Convey("test invalid domain", t, testInvalidDomain)
	convey.Convey("test invalid ip", t, testInvalidIp)
	convey.Convey("test invalid port", t, testInvalidPort)
	convey.Convey("test invalid account", t, testInvalidAccount)
	convey.Convey("test invalid password", t, testInvalidPasswd)
}

func testValidImageCfg() {
	imageCfg := getImageCfgTemplate()
	resp := NewConfigChecker().Check(imageCfg)
	convey.So(resp.Result, convey.ShouldEqual, true)
}

func testInvalidDomain() {
	imageCfg := getImageCfgTemplate()
	testData := []string{
		"fd",
		"fd.test.com.",
		"fd.test.com-",
		"-fd.test.com",
		".fd.test.com",
	}
	for _, data := range testData {
		imageCfg.Domain = data
		resp := NewConfigChecker().Check(imageCfg)
		convey.So(resp.Result, convey.ShouldEqual, false)
	}
}

func testInvalidIp() {
	imageCfg := getImageCfgTemplate()
	testData := []string{
		"0.0.0.0",
		"255.255.255.255",
	}
	imageCfg.Domain = ""
	for _, data := range testData {
		imageCfg.IP = data
		resp := NewConfigChecker().Check(imageCfg)
		convey.So(resp.Result, convey.ShouldEqual, false)
	}
}

func testInvalidPort() {
	imageCfg := getImageCfgTemplate()
	testData := []int64{0, invalidPort}
	for _, data := range testData {
		imageCfg.Port = data
		resp := NewConfigChecker().Check(imageCfg)
		convey.So(resp.Result, convey.ShouldEqual, false)
	}
}

func testInvalidAccount() {
	imageCfg := getImageCfgTemplate()
	testData := []string{
		"",
		"testAccount~",
		"testAccount-",
		"_testAccount",
	}
	for _, data := range testData {
		imageCfg.Account = data
		resp := NewConfigChecker().Check(imageCfg)
		convey.So(resp.Result, convey.ShouldEqual, false)
	}
}

func testInvalidPasswd() {
	imageCfg := getImageCfgTemplate()
	testData := [][]byte{
		{},
		{116, 101, 115, 116, 80, 97, 115, 115, 119, 100, 58, 49, 50, 37, 94},
		{116, 101, 115, 116, 49, 50, 51, 58, 38, 124, 52, 53, 54, 194, 183, 96},
		{116, 101, 115, 116, 49, 50, 51, 38, 124, 52, 53, 54, 194, 183, 96, 59, 60, 64, 65, 80, 81},
	}
	for _, data := range testData {
		imageCfg.Password = data
		resp := NewConfigChecker().Check(imageCfg)
		convey.So(resp.Result, convey.ShouldEqual, false)
	}
}
