//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package x509 password test file
package x509

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/terminal"
)

func TestGetRandomPass(t *testing.T) {
	convey.Convey("normal situation", t, func() {
		r1 := gomonkey.ApplyFunc(rand.Read, func(b []byte) (int, error) {
			for i := range b {
				b[i] = byte(i)
			}
			return len(b), nil
		})
		defer r1.Reset()
		res, err := GetRandomPass()
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(res), convey.ShouldNotEqual, 0)
	})
	convey.Convey("simple passwd situation", t, func() {
		r2 := gomonkey.ApplyFunc(rand.Read, func(b []byte) (int, error) {
			for i := range b {
				b[i] = 1
			}
			return len(b), nil
		})
		defer r2.Reset()
		_, err := GetRandomPass()
		convey.So(err.Error(), convey.ShouldEqual, "the password is to simple,please retry")
	})
}

// TestOverridePassWdFile test OverridePassWdFile
func TestOverridePassWdFile(t *testing.T) {
	convey.Convey("override padding test", t, func() {
		var path = "./testdata/test.key"
		data, err := fileutils.ReadLimitBytes(getAbsPath("./testdata/client.key", t), fileutils.Size10M)
		convey.So(err, convey.ShouldBeEmpty)
		err = OverridePassWdFile(path, data, fileutils.Mode600)
		convey.So(err, convey.ShouldBeEmpty)
		data2, err := fileutils.ReadLimitBytes(path, fileutils.Size10M)
		convey.So(err, convey.ShouldBeEmpty)
		convey.So(reflect.DeepEqual(data, data2), convey.ShouldBeTrue)
	})
}

// TestDecryptPrivateKeyWithPd test DecryptPrivateKeyWithPd
func TestDecryptPrivateKeyWithPd(t *testing.T) {
	convey.Convey("test for DecryptPrivateKey", t, func() {
		convey.Convey("private key is not encrypt", func() {
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			var ins *BackUpInstance
			mock := gomonkey.ApplyMethod(reflect.TypeOf(ins), "ReadFromDisk",
				func(_ *BackUpInstance, mode os.FileMode, needPadding bool) ([]byte, error) {
					return keyByte, nil
				})
			defer mock.Reset()
			block, err := DecryptPrivateKeyWithPd(getAbsPath("./testdata/client.key", t),
				getAbsPath("./testdata/backup/client.key", t), nil)
			convey.So(err, convey.ShouldEqual, nil)
			_, ok := block.Headers["DEK-Info"]
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("private key is  encrypted", func() {
			keyByte, err := fileutils.ReadLimitBytes(getAbsPath("./testdata/server-aes.key", t), fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			var ins *BackUpInstance
			mock := gomonkey.ApplyMethod(reflect.TypeOf(ins), "ReadFromDisk",
				func(_ *BackUpInstance, mode os.FileMode, needPadding bool) ([]byte, error) {
					return keyByte, nil
				})
			defer mock.Reset()
			block, err := DecryptPrivateKeyWithPd(getAbsPath("./testdata/server-aes.key", t),
				getAbsPath("./testdata/backup/server-aes.key", t), []byte("111111"))
			convey.So(err, convey.ShouldEqual, nil)
			_, ok := block.Headers["DEK-Info"]
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

// TestReadPassWd test ReadPassWd
func TestReadPassWd(t *testing.T) {
	convey.Convey("test for ReadPassWd", t, func() {
		convey.Convey("read passwd failed", func() {
			gomonkey.ApplyFunc(terminal.ReadPassword, func(fd, maxReadLength int) ([]byte, error) {
				return nil, errors.New("read passwd failed")
			})

			bytePassword, err := ReadPassWd()
			convey.So(bytePassword, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("read passwd succeed but too long", func() {
			gomonkey.ApplyFunc(terminal.ReadPassword, func(fd, maxReadLength int) ([]byte, error) {
				const longPass = 3000
				return make([]byte, longPass), nil
			})

			bytePassword, err := ReadPassWd()
			convey.So(bytePassword, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "input too long")
		})

		convey.Convey("read passwd succeed with valid length", func() {
			gomonkey.ApplyFunc(terminal.ReadPassword, func(fd, maxReadLength int) ([]byte, error) {
				const shortPass = 20
				return make([]byte, shortPass), nil
			})

			bytePassword, err := ReadPassWd()
			convey.So(bytePassword, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
