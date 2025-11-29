// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

func TestWriteWithLock(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tempFile")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()
	date := []byte("test date")

	convey.Convey("TestWriteWithLock", t, func() {
		convey.So(WriteWithLock(tempFile.Name(), date), convey.ShouldResemble, nil)
	})

	convey.Convey("test func WriteWithLock failed, open file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.OpenFile, nil, test.ErrTest)
		defer p1.Reset()
		err = WriteWithLock(tempFile.Name(), date)
		expErr := fmt.Errorf("open file[%s] failed: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func WriteWithLock failed, lock file failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(syscall.Flock, test.ErrTest)
		defer p1.Reset()
		err = WriteWithLock(tempFile.Name(), date)
		expErr := fmt.Errorf("lock file[%s] failed: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func WriteWithLock failed, seek file failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Seek", int64(0), test.ErrTest)
		defer p1.Reset()
		err = WriteWithLock(tempFile.Name(), date)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func WriteWithLock failed, truncate file failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Truncate", test.ErrTest)
		defer p1.Reset()
		err = WriteWithLock(tempFile.Name(), date)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func WriteWithLock failed, write file failed", t, func() {
		var p1 = gomonkey.ApplyMethodReturn(&os.File{}, "Write", 0, test.ErrTest)
		defer p1.Reset()
		err = WriteWithLock(tempFile.Name(), date)
		expErr := fmt.Errorf("write file[%s] failed: %v", tempFile.Name(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestInSamePartition(t *testing.T) {
	convey.Convey("test func InSamePartition success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(syscall.Stat, nil)
		defer p1.Reset()
		_, err := InSamePartition("", "")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func InSamePartition failed, syscall.Stat failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},

			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(syscall.Stat, outputs)
		defer p1.Reset()

		_, err := InSamePartition("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
		_, err = InSamePartition("", "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestSetImmutable(t *testing.T) {
	convey.Convey("test func SetImmutable success", t, func() {
		var p1 = gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
		defer p1.Reset()
		convey.So(SetImmutable(""), convey.ShouldResemble, nil)
	})

	convey.Convey("test func SetImmutable failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		convey.So(SetImmutable(""), convey.ShouldResemble, test.ErrTest)
	})
}

func TestUnSetImmutable(t *testing.T) {
	convey.Convey("test func UnSetImmutable success", t, func() {
		var p1 = gomonkey.ApplyFunc(envutils.RunCommand, mockRunCommandForReturnNil)
		defer p1.Reset()
		convey.So(UnSetImmutable(""), convey.ShouldResemble, nil)
	})

	convey.Convey("test func UnSetImmutable failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		convey.So(UnSetImmutable(""), convey.ShouldResemble, test.ErrTest)
	})
}

func mockRunCommandForReturnNil(_ string, _ int, _ ...string) (string, error) {
	return "Execution succeeded.", nil
}

func TestJson(t *testing.T) {
	temFile := CreateJsonFile(t)
	defer func() {
		if err := os.Remove(temFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()
	data, err := LoadJsonFile(temFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	convey.Convey("TestLoadJsonFile", t, func() {
		convey.Convey("test func LoadJsonFile success", func() {
			_, err = LoadJsonFile(temFile.Name())
			convey.So(err, convey.ShouldResemble, nil)
		})

		convey.Convey("test func LoadJsonFile failed, load file failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
			defer p1.Reset()
			_, err = LoadJsonFile(temFile.Name())
			expErr := fmt.Errorf("read json file failed: %v", test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})

		convey.Convey("test func LoadJsonFile failed, unmarshal failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
			defer p1.Reset()
			_, err = LoadJsonFile(temFile.Name())
			expErr := fmt.Errorf("unmarshal json value failed: %v", test.ErrTest)
			convey.So(err, convey.ShouldResemble, expErr)
		})
	})

	convey.Convey("TestSaveJsonValue", t, func() {
		convey.Convey("test func SaveJsonValue success", func() {
			convey.So(SaveJsonValue(temFile.Name(), data), convey.ShouldResemble, nil)
		})

		convey.Convey("test func SaveJsonValue failed, marshal indent failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(json.MarshalIndent, nil, test.ErrTest)
			defer p1.Reset()
			convey.So(SaveJsonValue(temFile.Name(), data), convey.ShouldResemble, errors.New("marshal json value failed"))
		})

		convey.Convey("test func SaveJsonValue failed, write data failed", func() {
			var p1 = gomonkey.ApplyFuncReturn(fileutils.WriteData, test.ErrTest)
			defer p1.Reset()
			convey.So(SaveJsonValue(temFile.Name(), data), convey.ShouldResemble, errors.New("write json file failed"))
		})
	})
}

func TestTestSetJsonValue(t *testing.T) {
	content := map[string]string{"key1": "value1", "key2": "value2"}
	bytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}
	jsonValue := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &jsonValue); err != nil {
		panic(err)
	}
	convey.Convey("test func SetJsonValue success", t, func() {
		err = SetJsonValue(jsonValue, "new value1", "key1")
		convey.So(err, convey.ShouldResemble, nil)

		err = SetJsonValue(jsonValue, "new valueX", "keyX")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func SetJsonValue failed, object is nil", t, func() {
		err = SetJsonValue(nil, "new value1", "key1")
		convey.So(err, convey.ShouldResemble, errors.New("map is nil"))
	})

	convey.Convey("test func SetJsonValue failed, provide no value", t, func() {
		err = SetJsonValue(jsonValue, "new value1")
		convey.So(err, convey.ShouldResemble, errors.New("provide at least one name"))
	})
}

func CreateJsonFile(t *testing.T) *os.File {
	data := map[string]string{
		"name":        "JsonTest",
		"date":        "JsonTest",
		"ExecStart =": "JsonTest",
	}

	tempJson, err := os.CreateTemp("", "tempjson")
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(tempJson)
	if err := encoder.Encode(data); err != nil {
		t.Fatal(err)
	}
	if err := tempJson.Close(); err != nil {
		t.Fatal(err)
	}
	return tempJson
}

func TestSetPathOwnerGroupToMEFEdge(t *testing.T) {
	patch1 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil)
	defer patch1.Reset()
	patch2 := gomonkey.ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer patch2.Reset()
	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	}()

	convey.Convey("TestSetPathOwnerGroupToMEFEdge", t, func() {
		convey.So(SetPathOwnerGroupToMEFEdge(tempDir, false, true), convey.ShouldResemble, nil)
	})

	convey.Convey("test func SetPathOwnerGroupToMEFEdge failed, get mef id failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err = SetPathOwnerGroupToMEFEdge(tempDir, false, true)
		expErr := fmt.Errorf("get uid/gid of mef-edge failed: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func SetPathOwnerGroupToMEFEdge failed, set path owner group failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
		defer p1.Reset()
		err = SetPathOwnerGroupToMEFEdge(tempDir, false, true)
		expErr := fmt.Errorf("set dir [%s] owner and group failed, error: %v", tempDir, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestCreateBackupWithMefOwner(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(GetMefId, uint32(os.Geteuid()), uint32(os.Geteuid()), nil).
		ApplyFuncReturn(backuputils.BackUpFiles, nil).
		ApplyFuncReturn(SetEuidAndEgid, nil)
	defer patch.Reset()
	tempFilePath := "tempFile"
	tempFile, err := os.OpenFile(tempFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileutils.Mode600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = tempFile.Close(); err != nil {
			t.Fatal(err)
		}
		if err = os.RemoveAll(tempFilePath); err != nil {
			t.Fatal(err)
		}
	}()

	convey.Convey("test func CreateBackupWithMefOwner success", t, func() {
		convey.So(CreateBackupWithMefOwner(tempFilePath), convey.ShouldResemble, nil)

		// reset failed
		var p1 = gomonkey.ApplyMethodReturn(&EdgeGUidMgr{}, "ResetEUGid", test.ErrTest)
		defer p1.Reset()
		convey.So(CreateBackupWithMefOwner(tempFilePath), convey.ShouldResemble, nil)
	})

	convey.Convey("test func CreateBackupWithMefOwner failed, set euid and egid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(GetMefId, uint32(0), uint32(0), test.ErrTest)
		defer p1.Reset()
		innerErr := fmt.Errorf("get mef-edge uid/gid failed, %v", test.ErrTest)
		expErr := fmt.Errorf("set euid/egid to mef-edge failed: %v", innerErr)
		convey.So(CreateBackupWithMefOwner(tempFilePath), convey.ShouldResemble, expErr)
	})

	convey.Convey("test func CreateBackupWithMefOwner failed, back up failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(backuputils.BackUpFiles, test.ErrTest)
		defer p1.Reset()
		expErr := fmt.Errorf("back up file with mef-edge owner failed, %v", test.ErrTest)
		convey.So(CreateBackupWithMefOwner(tempFilePath), convey.ShouldResemble, expErr)
	})
}
