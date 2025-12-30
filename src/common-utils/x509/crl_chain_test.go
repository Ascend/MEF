// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package x509

import (
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCheckCrlsChainReturnContent(t *testing.T) {
	convey.Convey("test for CheckCrlsChainReturnContent func", t, func() {
		convey.Convey("test for CheckCrlsChainReturnContent success", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/crl/chain.crl", t),
				getAbsPath("./testdata/ca_chain.crt", t))
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("test for CheckCrlsChainReturnContent crl does not exist", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/notexists", t),
				getAbsPath("./testdata/ca_chain.crt", t))
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})

		convey.Convey("test for CheckCrlsChainReturnContent no PEM crl", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/ca_chain.crt", t),
				getAbsPath("./testdata/ca_chain.crt", t))
			convey.So(err, convey.ShouldResemble, errors.New("new crl Chain Mgr failed: no PEM crl contains"))
		})

		convey.Convey("test for CheckCrlsChainReturnContent wrong cert", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/crl/chain.crl", t),
				getAbsPath("./testdata/crl/chain.crl", t))
			convey.So(err, convey.ShouldResemble,
				errors.New("check crl chain failed: init ca chain manager instance failed: no PEM cert contains"))
		})

		convey.Convey("test for CheckCrlsChainReturnContent that crt and chain does not match", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/crl/chain.crl", t),
				getAbsPath("./testdata/ca.crt", t))
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check crl chain failed: %v", ErrCrlCertNotMatch.Error()))
		})

		convey.Convey("test for CheckCrlsChainReturnContent that crl is invalid", func() {
			_, err := CheckCrlsChainReturnContent(getAbsPath("./testdata/crl/wrong.crl", t),
				getAbsPath("./testdata/ca.crt", t))
			convey.So(err, convey.ShouldResemble,
				errors.New("new crl Chain Mgr failed: parse single crl failed: "+
					"asn1: syntax error: truncated tag or length"))
		})
	})
}

func TestParseCrls(t *testing.T) {
	convey.Convey("test for ParseCrls func", t, func() {
		convey.Convey("test for ParseCrls success", func() {
			_, err := ParseCrls(&CrlData{CrlPath: getAbsPath("./testdata/crl/chain.crl", t)})
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("test for CheckCrlsChainReturnContent crl does not exist", func() {
			_, err := ParseCrls(&CrlData{CrlPath: getAbsPath("./testdata/notexists", t)})
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})

		convey.Convey("test for CheckCrlsChainReturnContent no PEM crl", func() {
			_, err := ParseCrls(&CrlData{CrlPath: getAbsPath("./testdata/ca_chain.crt", t)})
			convey.So(err, convey.ShouldResemble, errors.New("new crl Chain Mgr failed: no PEM crl contains"))
		})
	})
}
