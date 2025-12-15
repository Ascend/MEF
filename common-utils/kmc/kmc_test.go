// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package kmc interface test
package kmc

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"huawei.com/mindx/common/hwlog"
)

var (
	nilCtx = Context{ctx: unsafe.Pointer(nil)}
)

func TestMain(m *testing.M) {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitHwLogger(logConfig, logConfig); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
	m.Run()
}

func TestInitAndFinalize(t *testing.T) {
	if err := Initialize(Aes256gcm, "primary.dat", "standby.dat"); err != nil {
		t.Errorf("Initialkize failed %v\n", err)
	}
	if err := Initialize(Aes256gcm, "primary.dat", "standby.dat"); err == nil {
		t.Errorf("Initialkize failed %v\n", err)
	}
	Finalize()
	if kmcInstance.ctx != nil {
		t.Fail()
	}
	if err := os.Remove("primary.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
	if err := os.Remove("standby.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
}

func TestInitializeWithInvalidAlgId(t *testing.T) {
	const invalidAlgorithmId = 100
	if Initialize(invalidAlgorithmId, "primary.dat", "standby.dat") == nil {
		t.Errorf("init must failed")
	}
}

func TestFinalizeFailed(t *testing.T) {
	const KeErrParamCheck = 1004
	err := nilCtx.KeFinalizeEx()
	if err == nil || err.(*KeKmcError).Code() != KeErrParamCheck {
		t.Fail()
	}
	t.Logf("Fainalize Failed %v", err.Error())
}

func TestEncryptAndDecrypt(t *testing.T) {
	Initialize(Aes256gcm, "primary.dat", "standby.dat")
	cipherBuf, err := Encrypt(0, []byte("hello"))
	if err != nil {
		fmt.Printf("Encrypt error: %v\n", err)
		t.Fail()
	}
	plainBuf, err := Decrypt(0, cipherBuf)
	if err != nil {
		fmt.Printf("Encrypt error: %v\n", err)
		t.Fail()
	}
	if bytes.Compare(plainBuf, []byte("hello")) != 0 {
		t.Fail()
	}
	Finalize()
	if err := os.Remove("primary.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
	if err := os.Remove("standby.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
}

func TestEncryptAndDecryptZeroData(t *testing.T) {
	Initialize(Aes256gcm, "primary.dat", "standby.dat")
	_, err := Encrypt(0, []byte(""))
	if err == nil {
		fmt.Printf("Encrypt Zero data error: %v\n", err)
		t.Fail()
	}
	_, err = Decrypt(0, []byte(""))
	if err == nil {
		fmt.Printf("Decrypt Zero data error: %v\n", err)
		t.Fail()
	}
	Finalize()
	if err := os.Remove("primary.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
	if err := os.Remove("standby.dat"); err != nil {
		t.Errorf("remove test file failed %v\n", err)
	}
}
