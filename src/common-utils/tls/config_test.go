// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tls config test file
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	cx509 "huawei.com/mindx/common/x509"
)

// TestNewTLSConfig test for new tls
func TestNewTLSConfig(t *testing.T) {
	convey.Convey("test for NewTLSConfig", t, func() {
		c := tls.Certificate{}
		convey.Convey("One-way HTTPS", func() {
			conf, err := NewTLSConfig([]byte{}, c, []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256})
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(conf, convey.ShouldNotBeEmpty)
		})
		convey.Convey("Two-way HTTPS,but ca check failed", func() {
			conf, err := NewTLSConfig([]byte("sdsddd"), c, []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256})
			convey.So(conf, convey.ShouldEqual, nil)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("Two-way HTTPS", func() {
			ca, err := cx509.CheckCaCert("./testdata/ca.crt", cx509.InvalidNum)
			conf, err := NewTLSConfig(ca, c, []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256})
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(conf, convey.ShouldNotBeEmpty)
		})

	})
}

// TestGetTLSConfigForClient test for GetTLSConfigForClient
func TestGetTLSConfigForClient(t *testing.T) {
	convey.Convey("get tlsconfig", t, func() {
		mock := gomonkey.ApplyFunc(cx509.LoadCertPairByte, func(pathMap map[string]string, encryptAlgorithm int,
			mode os.FileMode) ([]byte, []byte, error) {
			return nil, nil, errors.New("error")
		})
		defer mock.Reset()
		cfg, err := GetTLSConfigForClient("npu-exporter", 1)
		convey.So(err, convey.ShouldNotBeEmpty)
		convey.So(cfg, convey.ShouldNotBeEmpty)
		convey.So(cfg, convey.ShouldEqual, nil)
	})

	convey.Convey("get tlsconfig succuss", t, func() {
		var pool *x509.CertPool
		mock := gomonkey.ApplyFunc(LoadCertPair, func(pathMap map[string]string,
			encryptAlgorithm int) (*tls.Certificate,
			error) {
			return &tls.Certificate{}, nil
		})
		defer mock.Reset()
		mock2 := gomonkey.ApplyFunc(cx509.VerifyCaCert, func(caBytes []byte, overdueTime int) error {
			return nil
		})
		defer mock2.Reset()
		mock3 := gomonkey.ApplyMethod(reflect.TypeOf(pool), "AppendCertsFromPEM",
			func(_ *x509.CertPool, _ []byte) bool {
				return true
			})
		defer mock3.Reset()
		var ins *cx509.BackUpInstance
		mock4 := gomonkey.ApplyMethod(reflect.TypeOf(ins), "ReadFromDisk",
			func(_ *cx509.BackUpInstance, mode os.FileMode, needPadding bool) ([]byte, error) {
				return []byte{1, 1, 1}, nil
			})
		defer mock4.Reset()
		cfg, err := GetTLSConfigForClient("npu-exporter", 1)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cfg, convey.ShouldNotBeEmpty)
	})
}
