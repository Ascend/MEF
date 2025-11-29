// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker for test domain checker
package checker

import (
	"errors"
	"net"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestDomainChecker(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{}, nil)
	defer p.Reset()

	checker := GetDomainChecker("", true, true, true)
	convey.Convey("test normal case", t, func() {
		domain := "fd.Test-123.com"
		checkResult := checker.Check(domain)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})

	convey.Convey("parsing error can be ignored", t, func() {
		domain := "fd.Test-123.com"
		p1 := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{}, errors.New("test error"))
		defer p1.Reset()
		checkResult := checker.Check(domain)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})

	convey.Convey("domain does not match regex", t, func() {
		domain := "fd.Test~123.com"
		checkResult := checker.Check(domain)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "domain does not match allowed regex")
	})

	convey.Convey("domain can not be all digits", t, func() {
		domain := "123456"
		checkResult := checker.Check(domain)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "domain can not be all digits")
	})

	convey.Convey("domain can not contain localhost", t, func() {
		domain := "localhost"
		checkResult := checker.Check(domain)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "domain can not contain localhost")
	})

	convey.Convey("domain is not allowed to be a loop back address", t, func() {
		domain := "Euler"
		p1 := gomonkey.ApplyMethodReturn(net.DefaultResolver, "LookupIP", []net.IP{net.ParseIP("127.0.0.1")}, nil)
		defer p1.Reset()
		checkResult := checker.Check(domain)
		convey.So(checkResult.Reason, convey.ShouldContainSubstring, "domain is not allowed to be a loop back address")
	})
}
