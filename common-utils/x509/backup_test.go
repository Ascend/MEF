// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 provides the capability of x509.
package x509

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
)

func getAbsPath(relPath string, t *testing.T) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("get client key abs path failed")
	}
	return absPath
}

func TestNewBKPInstance(t *testing.T) {

	convey.Convey("normal situation,no error returned", t, func() {
		_, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/client.keybkp", t))
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("no provide path,error returned", t, func() {
		_, err := NewBKPInstance(nil, "", "")
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("provide path,but check failed,error returned", t, func() {
		exsitStub := gomonkey.ApplyFunc(fileutils.IsExist, func(filePath string) bool {
			return true
		})
		defer exsitStub.Reset()
		mockStub := gomonkey.ApplyFunc(fileutils.CheckOriginPath, func(path string) (string, error) {
			return "", errors.New("mock error")
		})
		defer mockStub.Reset()
		_, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/client.keybkp", t))
		convey.So(err.Error(), convey.ShouldEqual, "mock error")
	})
}

func TestWriteToDisk(t *testing.T) {
	convey.Convey("normal situation without padding,no error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(ioutil.WriteFile, func(filename string, data []byte,
			perm fs.FileMode) error {
			return nil
		})
		defer existStub.Reset()
		err = data.WriteToDisk(fileutils.Mode600, false)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("normal situation without padding,error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(ioutil.WriteFile, func(filename string, data []byte,
			perm fs.FileMode) error {
			return errors.New("mock error")
		})
		defer existStub.Reset()
		err = data.WriteToDisk(fileutils.Mode600, false)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("normal situation with padding,no error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(OverridePassWdFile, func(path string, data []byte, mode os.FileMode) error {
			return nil
		})
		defer existStub.Reset()
		err = data.WriteToDisk(fileutils.Mode600, true)
		convey.So(err, convey.ShouldEqual, nil)
	})

	convey.Convey("normal situation with padding, error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(OverridePassWdFile, func(path string, data []byte, mode os.FileMode) error {
			return errors.New("mock error")
		})
		defer existStub.Reset()
		err = data.WriteToDisk(fileutils.Mode600, true)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestCommonValid(t *testing.T) {
	convey.Convey("normal situation , no error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		err = commonValid(data)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("no data  ,  error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		err = commonValid(data)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("no path  ,  error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		data.mainPath = ""
		err = commonValid(data)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("same path  ,  error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		err = commonValid(data)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestReadFromDisk(t *testing.T) {
	convey.Convey("normal situation, read from main file , no error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(readFromFile, func(b *BackUpInstance, isMain, needPadding bool,
			mode os.FileMode) ([]byte, error) {
			return []byte("test"), nil
		})
		defer existStub.Reset()
		rs, err := data.ReadFromDisk(fileutils.Mode600, false)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(rs, convey.ShouldNotBeNil)
	})
	convey.Convey("normal situation ,read from back up file, no error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/backup/client.key", t),
			getAbsPath("./testdata/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(readFromFile, func(b *BackUpInstance, isMain, needPadding bool,
			mode os.FileMode) ([]byte, error) {
			return []byte("test"), nil
		})
		defer existStub.Reset()
		rs, err := data.ReadFromDisk(fileutils.Mode600, false)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(rs, convey.ShouldNotBeNil)
	})
	convey.Convey("both main and backup file not exist ,error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/backup/xxx.key", t),
			getAbsPath("./testdata/xxx.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(readFromFile, func(b *BackUpInstance, isMain, needPadding bool,
			mode os.FileMode) ([]byte, error) {
			return []byte("test"), nil
		})
		defer existStub.Reset()
		rs, err := data.ReadFromDisk(fileutils.Mode600, false)
		convey.So(rs, convey.ShouldEqual, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestInstanceNotInit(t *testing.T) {
	convey.Convey("instance not init ,error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/backup/xxx.key", t),
			getAbsPath("./testdata/xxx.key", t))
		data.mainPath = ""
		convey.So(err, convey.ShouldEqual, nil)
		existStub := gomonkey.ApplyFunc(readFromFile, func(b *BackUpInstance, isMain, needPadding bool,
			mode os.FileMode) ([]byte, error) {
			return []byte("test"), nil
		})
		defer existStub.Reset()
		rs, err := data.ReadFromDisk(fileutils.Mode600, false)
		convey.So(rs, convey.ShouldEqual, nil)
		convey.So(err, convey.ShouldEqual, ErrInstanceEmpty)
	})
}

func TestReadFromFile(t *testing.T) {
	convey.Convey("old version file ,no error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/client.key", t),
			getAbsPath("./testdata/backup/client.key", t))
		convey.So(err, convey.ShouldEqual, nil)
		writeMock := gomonkey.ApplyMethodFunc(data, "WriteToDisk",
			func(mode os.FileMode, needPadding bool) error {
				return nil
			})
		defer writeMock.Reset()
		rs, err := readFromFile(data, true, false, fileutils.Mode600)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(rs, convey.ShouldNotBeNil)
	})

	convey.Convey("new version file ,no error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/backup/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		writeMock := gomonkey.ApplyMethodFunc(data, "WriteToDisk",
			func(mode os.FileMode, needPadding bool) error {
				return nil
			})
		defer writeMock.Reset()
		rs, err := readFromFile(data, true, false, fileutils.Mode600)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(rs, convey.ShouldNotBeNil)
	})
	convey.Convey("new version file,but verify failed ,error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		writeMock := gomonkey.ApplyMethodFunc(data, "WriteToDisk",
			func(mode os.FileMode, needPadding bool) error {
				return nil
			})
		defer writeMock.Reset()
		verifyMock := gomonkey.ApplyMethodFunc(data, "Verify", func() error {
			return errors.New("writeMock err")
		})
		defer verifyMock.Reset()
		rs, err := readFromFile(data, true, false, fileutils.Mode600)
		convey.So(rs, convey.ShouldEqual, nil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestVerify(t *testing.T) {
	convey.Convey("data is empty ,error returned", t, func() {
		data, err := NewBKPInstance(nil, getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/backup/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		err = data.Verify()
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("checksum is empty ,error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/backup/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		data.checkSum = nil
		err = data.Verify()
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("verify pass ,no error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/backup/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		err = data.Verify()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("verify failed ,error returned", t, func() {
		data, err := NewBKPInstance([]byte("test"), getAbsPath("./testdata/config1", t),
			getAbsPath("./testdata/backup/config1", t))
		convey.So(err, convey.ShouldEqual, nil)
		data.checkSum = []byte("ddd")
		err = data.Verify()
		convey.So(err, convey.ShouldNotBeNil)
	})
}
