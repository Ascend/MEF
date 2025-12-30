// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils provides the util func
package utils

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const byteLength = 32

func TestReplacePrefix(t *testing.T) {
	convey.Convey("relative path", t, func() {
		path := ReplacePrefix("./testdata/cert/ca.crt", "****")
		convey.So(path, convey.ShouldEqual, "****testdata/cert/ca.crt")
	})
	convey.Convey("abconvey.Solute path", t, func() {
		path := ReplacePrefix("/testdata/cert/ca.crt", "****")
		convey.So(path, convey.ShouldEqual, "****estdata/cert/ca.crt")
	})
	convey.Convey("path length less than 2", t, func() {
		path := ReplacePrefix("/", "****")
		convey.So(path, convey.ShouldEqual, "****")
	})
	convey.Convey("empty string", t, func() {
		path := ReplacePrefix("", "****")
		convey.So(path, convey.ShouldEqual, "****")
	})

}

func TestMaskPrefix(t *testing.T) {
	convey.Convey("relative path", t, func() {
		path := MaskPrefix("./testdata/cert/ca.crt")
		convey.So(path, convey.ShouldEqual, "****testdata/cert/ca.crt")
	})
	convey.Convey("abconvey.Solute path", t, func() {
		path := MaskPrefix("/testdata/cert/ca.crt")
		convey.So(path, convey.ShouldEqual, "****estdata/cert/ca.crt")
	})
	convey.Convey("path length less than 2", t, func() {
		path := MaskPrefix("/")
		convey.So(path, convey.ShouldEqual, "****")
	})
	convey.Convey("empty string", t, func() {
		path := MaskPrefix("")
		convey.So(path, convey.ShouldEqual, "****")
	})

}

func TestGetSha256Code(t *testing.T) {
	convey.Convey("test sha256", t, func() {
		hashs := GetSha256Code([]byte("this is a test sentence"))
		convey.So(len(hashs), convey.ShouldEqual, byteLength)
	})
}

func TestBinaryFormat(t *testing.T) {
	convey.Convey("test binary format", t, func() {
		const (
			num         = 1234
			bitSize32   = 32
			base16      = 16
			byteLength4 = 4
			allOccurs   = -1
			strLen11    = 11
		)
		numStr := BinaryFormat(big.NewInt(num).Bytes(), byteLength4)
		convey.So(len(numStr), convey.ShouldEqual, strLen11)
		hexStr := strings.Replace(numStr, ":", "", allOccurs)
		newNum, err := strconv.ParseInt(hexStr, base16, bitSize32)
		convey.So(err, convey.ShouldBeNil)
		convey.So(newNum, convey.ShouldEqual, int64(num))
	})
}
