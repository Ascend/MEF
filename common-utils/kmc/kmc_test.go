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
	"reflect"
	"testing"
	"unsafe"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
)

const (
	exDomainId       = 13
	exPlain          = "12345678"
	testDomainID     = 10
	testKeyID        = 1
	testAdvanceDay   = 90
	testHmacDomainID = 0
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

func setup(t *testing.T, name string) (Context, error) {
	cfg := NewKmcInitConfig()
	cfg.PrimaryKeyStoreFile = fmt.Sprintf("test_%s_primary.dat", name)
	cfg.StandbyKeyStoreFile = fmt.Sprintf("test_%s_standby.dat", name)
	return KeInitializeEx(cfg)
}

func teardown(t *testing.T, name string) {
	if os.Remove(fmt.Sprintf("test_%s_primary.dat", name)) != nil {
		t.Fail()
	}
	if os.Remove(fmt.Sprintf("test_%s_standby.dat", name)) != nil {
		t.Fail()
	}
	t.Log("Success")
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

func TestSetLoggerLevel(t *testing.T) {
	const maxLogLevel int = 6
	err1 := KeSetLoggerLevel(maxLogLevel)
	err2 := KeSetLoggerLevel(0)
	if err1 == nil || err2 != nil {
		t.Fail()
	}
}

func TestUpdateLifetimeDays(t *testing.T) {
	const lifeimeDays = 90
	if UpdateLifetimeDays(0) == nil {
		t.Fail()
	}
	if UpdateLifetimeDays(lifeimeDays) != nil {
		t.Fail()
	}
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

func TestGeneratedMKEncryptDecryptEx(t *testing.T) {
	testCtx, err := setup(t, "encryptanddecrypt")
	if err != nil {
		t.Fail()
	}
	_, err = testCtx.KeGeneratedKeyAndGetIDEx(exDomainId)
	if err != nil {
		t.Fail()
	}
	plainText := []byte(exPlain)
	cipherText, err1 := testCtx.KeEncryptByDomainEx(exDomainId, plainText)
	plainTextNew, err2 := testCtx.KeDecryptByDomainEx(exDomainId, cipherText)
	if err1 != nil || err2 != nil {
		t.Fail()
	}

	if !reflect.DeepEqual([]byte(exPlain), plainTextNew) {
		t.Errorf("expected:%v. got:%v", []byte(exPlain), plainTextNew)
	}
	testCtx.KeFinalizeEx()
	teardown(t, "encryptanddecrypt")
}

func TestKeGetCipherDataLenEx(t *testing.T) {
	testCtx, err := setup(t, "getcipherdatalen")
	if err != nil {
		t.Fail()
	}
	const plainLength = 100
	const invalidPlainLength = -100
	cipherLen, err := testCtx.KeGetCipherDataLenEx(plainLength)
	if err != nil || cipherLen <= plainLength {
		t.Fail()
	}
	cipherLen, err = testCtx.KeGetCipherDataLenEx(invalidPlainLength)
	if err == nil || cipherLen != 0 {
		t.Fail()
	}
	testCtx.KeFinalizeEx()
	teardown(t, "getcipherdatalen")
}

func TestRegisterAndGetKey(t *testing.T) {
	tmpCtx, err := setup(t, "register")
	if err != nil {
		t.Errorf("setup TestRegisterByteKeyEx failed\n")
	}
	if err = tmpCtx.KeRegisterByteKeyEx(testDomainID, testKeyID, []byte("hello")); err != nil {
		t.Errorf("register failed %v\n", err)
	}
	key, err1 := tmpCtx.KeGetKeyByIDEx(testDomainID, testKeyID, false)
	keyBase64, err2 := tmpCtx.KeGetKeyByIDEx(testDomainID, testKeyID, true)
	tmpCtx.KeFinalizeEx()
	if err1 != nil || bytes.Compare(key, []byte("hello")) != 0 {
		t.Errorf("get key failed %v\n", err1)
	}
	// base64 of hello is aGVsbG8=
	if err2 != nil || string(keyBase64) != "aGVsbG8=" {
		t.Errorf("get base64 key failed %v %v\n", err1, string(keyBase64))
	}
	teardown(t, "register")
}

func TestRegisterByteKeyExZeroLength(t *testing.T) {
	tmpCtx, err := setup(t, "register")
	if err != nil {
		t.Errorf("setup TestRegisterByteKeyEx failed\n")
	}
	if err = tmpCtx.KeRegisterByteKeyEx(testDomainID, testKeyID, []byte("")); err == nil {
		t.Errorf("register success, except failed")
	}
	teardown(t, "register")
}

func TestSetKeyStatus(t *testing.T) {
	tmpCtx, err := setup(t, "keystatus")
	if err != nil {
		t.Errorf("setup failed %v\n", err)
	}
	if err = tmpCtx.KeRegisterByteKeyEx(testDomainID, testKeyID, []byte("this is key")); err != nil {
		t.Errorf("register key failed error code = %d  info = %v\n", err.(*KeKmcError).Code(), err)
	}
	if err = tmpCtx.KeSetMkStatusEx(testDomainID, testKeyID, 0); err != nil {
		t.Errorf("set MK status failed error code = %d  info = %v\n", err.(*KeKmcError).Code(), err)
	}
	const invalidKeyStatus = 5
	if err = tmpCtx.KeSetMkStatusEx(testDomainID, testKeyID, invalidKeyStatus); err == nil {
		t.Errorf("set MK status with invalid key status success")
	}
	key, err := tmpCtx.KeGetKeyByIDEx(testDomainID, testKeyID, false)
	tmpCtx.KeFinalizeEx()
	if bytes.Compare(key, []byte("this is key")) != 0 {
		t.Fail()
	}
	teardown(t, "keystatus")
}

func TestKeRemoveKeyByIDEx(t *testing.T) {
	tmpCtx, err := setup(t, "remove")
	if err != nil {
		t.Errorf("set up failed %v\n", err)
	}
	if err = tmpCtx.KeRegisterByteKeyEx(testDomainID, testKeyID, []byte("this is test key")); err != nil {
		t.Errorf("register key failed code = %d info = %v", err.(*KeKmcError).Code(), err)
	}
	// remove will fail because key is active
	if err = tmpCtx.KeRemoveKeyByIDEx(testDomainID, testKeyID); err == nil {
		t.Errorf("try remove active key success")
	}
	// set key inactive
	tmpCtx.KeSetMkStatusEx(testDomainID, testKeyID, KeyStatusInactive)
	// remove will success
	if err = tmpCtx.KeRemoveKeyByIDEx(testDomainID, testKeyID); err != nil {
		t.Errorf("remove key failed code = %d info = %v", err.(*KeKmcError).Code(), err)
	}
	key, err := tmpCtx.KeGetKeyByIDEx(testDomainID, testKeyID, false)
	if err == nil || len(key) > 0 {
		t.Fail()
	}
	teardown(t, "remove")
}

func TestKeCheckAndUpdateMkEx(t *testing.T) {
	tmpCtx, err := setup(t, "checkandupdate")
	_, err = tmpCtx.KeGeneratedKeyAndGetIDEx(0)
	if err != nil {
		t.Errorf("generate key and get id failed code = %d info = %v", err.(*KeKmcError).Code(), err)
	}
	err = tmpCtx.KeCheckAndUpdateMkEx(0, testAdvanceDay)
	if err != nil {
		t.Errorf("check and update failed code = %d info = %v", err.(*KeKmcError).Code(), err)
	}
	teardown(t, "checkandupdate")
}

func TestKeActiveNewKeyExSuccess(t *testing.T) {
	tmpCtx, err := setup(t, "activekey")
	err = tmpCtx.KeActiveNewKeyEx(0)
	if err != nil {
		t.Errorf("Active new key failed %v", err)
	}
	teardown(t, "activekey")
}

func TestEncryptUnitializeEncryptAndDecrypt(t *testing.T) {
	cipher, err := nilCtx.KeEncryptByDomainEx(testDomainID, []byte("test data"))
	if err == nil || len(cipher) > 0 || err.(*KeKmcError).Code() != kmcNotInit {
		t.Errorf("must encrypt failed")
	}
	plain, err := nilCtx.KeDecryptByDomainEx(testDomainID, []byte("test data"))
	if err == nil || len(plain) > 0 || err.(*KeKmcError).Code() != kmcNotInit {
		t.Errorf("must decrypt failed %v %v %v", err, plain, err.(*KeKmcError).Code())
	}
}

func TestCalculateHmacAndVerify(t *testing.T) {
	testCtx, err := setup(t, "calHmacAndVerify")
	if err != nil {
		t.Fail()
	}
	plainText := []byte(exPlain)
	cipher, err1 := testCtx.KeHmacByDomainV2Ex(testHmacDomainID, plainText)
	err2 := testCtx.KeHmacVerifyByDomainEx(testHmacDomainID, plainText, cipher)
	if err1 != nil || err2 != nil {
		t.Fail()
	}
	testCtx.KeFinalizeEx()
	teardown(t, "calHmacAndVerify")
}

func TestUnitializeKeRegisterByteKeyEx(t *testing.T) {
	err := nilCtx.KeRegisterByteKeyEx(testDomainID, testKeyID, []byte("this is test key"))
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit {
		t.Errorf("must register failed %v %v", err, err.(*KeKmcError).Code())
	}
}

func TestUnitializeKeActiveNewKeyEx(t *testing.T) {
	err := nilCtx.KeActiveNewKeyEx(testDomainID)
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit {
		t.Errorf("must active failed %v %v", err, err.(*KeKmcError).Code())
	}
}

func TestUnitializeKeGeneratedKeyAndGetIDEx(t *testing.T) {
	id, err := nilCtx.KeGeneratedKeyAndGetIDEx(testDomainID)
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit || id == testDomainID {
		t.Errorf("must generate key and get id failed %v %v", err, err.(*KeKmcError).Code())
	}
}
func TestUnitializeKeGetMaxMkIDEx(t *testing.T) {
	id, err := nilCtx.KeGetMaxMkIDEx(testDomainID)
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit || id == testDomainID {
		t.Errorf("must get max mk id failed %v %v", err, err.(*KeKmcError).Code())
	}
}
func TestUnitializeCipherLen(t *testing.T) {
	cipherLen, err := nilCtx.KeGetCipherDataLenEx(0)
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit || cipherLen != 0 {
		t.Errorf("must get cipher data len failed %v %v", err, err.(*KeKmcError).Code())
	}
}

func TestUnitializeKeGetKeyByIDEx(t *testing.T) {
	k, err := nilCtx.KeGetKeyByIDEx(testDomainID, 0, true)
	if err == nil || err.(*KeKmcError).Code() != kmcNotInit || k != nil {
		t.Errorf("must get key by id failed %v %v", err, err.(*KeKmcError).Code())
	}
}

func TestUnitializeKeyStatus(t *testing.T) {
	err1 := nilCtx.KeSetMkStatusEx(testDomainID, 0, KeyStatusActive)
	err2 := nilCtx.KeRemoveKeyByIDEx(testDomainID, 0)
	err3 := nilCtx.KeCheckAndUpdateMkEx(testDomainID, 0)
	if err1.(*KeKmcError).Code() != kmcNotInit || err2.(*KeKmcError).Code() != kmcNotInit ||
		err3.(*KeKmcError).Code() != kmcNotInit {
		t.Errorf("test failed")
	}
}

func TestSaltLen(t *testing.T) {
	const saltLenTooLong = 129
	const saltLenTooShort = 15
	const saltLenSuccess = 32
	SetSaltLen(saltLenTooLong)
	if kmcSaltLen == saltLenTooLong {
		t.Errorf("test failed: salt length error")
	}
	SetSaltLen(saltLenTooShort)
	if kmcSaltLen == saltLenTooShort {
		t.Errorf("test failed: salt length error")
	}
	SetSaltLen(saltLenSuccess)
	if kmcSaltLen != saltLenSuccess {
		t.Errorf("test failed: set salt len failed")
	}
}

func TestUpdateKmcTask(t *testing.T) {
	convey.Convey("test update kmc task", t, func() {
		config := NewKmcInitConfig()
		config.PrimaryKeyStoreFile = "./testKmcDir"
		config.StandbyKeyStoreFile = "./testKmcDir"
		config.SdpAlgId = Aes256gcmId

		ctx, err := KeInitializeEx(config)
		convey.So(err, convey.ShouldBeNil)
		task := ManualUpdateKmcTask{
			UpdateKmcTask: UpdateKmcTask{
				Ctx: &ctx,
			},
		}
		err = task.RunTask()
		convey.So(err, convey.ShouldBeNil)
	})
}

func makeString(strLen int) string {
	const num = 26
	const minLen = 0
	const maxLen = 10000

	if strLen < minLen || strLen > maxLen {
		return ""
	}
	str := make([]byte, strLen)
	for i := range str {
		str[i] = 'a' + byte(i%num)
	}
	return string(str)
}

func TestConvertToCCharArray(t *testing.T) {
	convey.Convey("test an empty string", t, func() {
		goStr := ""
		cChar, err := convertToCCharArray(goStr)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cChar[0], convey.ShouldBeZeroValue)
	})

	convey.Convey("test a string with a length less than SEC_PATH_MAX", t, func() {
		goStr := "test"
		cChar, err := convertToCCharArray(goStr)
		convey.So(err, convey.ShouldBeNil)
		for i := 0; i < len(goStr) && i < len(cChar); i++ {
			convey.So(goStr[i], convey.ShouldEqual, byte(cChar[i]))
		}
		convey.So(len(cChar), convey.ShouldBeGreaterThan, len(goStr))
		convey.So(cChar[len(goStr)], convey.ShouldBeZeroValue)
	})

	convey.Convey("test a string with a length equal to SEC_PATH_MAX", t, func() {
		const secPathMax = 4096
		goStr := makeString(secPathMax - 1)
		cChar, err := convertToCCharArray(goStr)
		convey.So(err, convey.ShouldBeNil)
		for i := 0; i < len(goStr) && i < len(cChar); i++ {
			convey.So(goStr[i], convey.ShouldEqual, byte(cChar[i]))
		}
		convey.So(len(cChar), convey.ShouldBeGreaterThan, len(goStr))
		convey.So(cChar[len(goStr)], convey.ShouldBeZeroValue)
	})

	convey.Convey("test a string with a length bigger than SEC_PATH_MAX", t, func() {
		const secPathMax = 4096
		goStr := makeString(secPathMax)
		_, err := convertToCCharArray(goStr)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("convertToCCharArray failed, the length of the "+
			"path exceeds SEC_PATH_MAX[%v]", secPathMax))
	})
}
