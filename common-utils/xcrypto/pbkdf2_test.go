// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package xcrypto provides basic algorithm for crypto.
package xcrypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
	"time"
)

type testData struct {
	password []byte
	salt     []byte
	iter     int
	dk       []byte // derived key
}

var sha256TestVectors = []testData{
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		1,
		// pbkdf2 result of this test
		[]byte{
			0x1a, 0x8b, 0xb2, 0xf8, 0x6a, 0xf8, 0x52, 0x17,
			0x3d, 0x88, 0x1c, 0x5c, 0xab, 0x5f, 0x82, 0x4e,
			0x79, 0x35, 0xac, 0xd9,
		},
	},
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		2,
		[]byte{
			0x77, 0x27, 0x67, 0x98, 0x30, 0x86, 0xb4, 0xa8,
			0x9f, 0xd8, 0x4f, 0xbb, 0xeb, 0x66, 0x50, 0x81,
			0x87, 0xe9, 0x76, 0x95,
		},
	},
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		2,
		[]byte{
			0x77,
		},
	},
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		4096,
		[]byte{
			0xa8, 0x6b, 0x1a, 0x2e, 0x46, 0x1d, 0x28, 0xee,
			0x06, 0xd6, 0xa9, 0xcc, 0xea, 0xef, 0xd4, 0x0f,
			0x94, 0xa3, 0x0b, 0x39,
		},
	},
	{
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		1,
		[]byte{
			0xe7, 0x32, 0x51, 0x9f, 0xd4, 0x82, 0x60, 0x39,
		},
	},
	{
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		[]byte{
			0xe7, 0x32, 0x51, 0x9f, 0xd4, 0x82, 0x60, 0x39,
		},
	},
}

var sha512TestVectors = []testData{
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		0,
		[]byte{
			0x84, 0xc3, 0x36, 0x6d, 0xf2, 0x4d, 0x17, 0x97,
			0xa6, 0xc5, 0xe8, 0x47, 0x1d, 0x53, 0x9d, 0x5d,
			0x98, 0x17, 0xc0, 0xf0,
		},
	},
	{
		[]byte("testpassword"),
		[]byte("testsalt"),
		1,
		[]byte{
			0x84, 0xc3, 0x36, 0x6d, 0xf2, 0x4d, 0x17, 0x97,
			0xa6, 0xc5, 0xe8, 0x47, 0x1d, 0x53, 0x9d, 0x5d,
			0x98, 0x17, 0xc0, 0xf0,
		},
	},
}

func TestPbkdf2WithSHA256(t *testing.T) {
	for i, v := range sha256TestVectors {
		dk, _ := Pbkdf2WithSha256([]byte(v.password), []byte(v.salt), v.iter, len(v.dk))
		if !bytes.Equal(dk, v.dk) {
			t.Errorf("pbkdf2 with sha256 %d: expected %x, got %x", i, v.dk, dk)
		}
	}
}

func TestPbkdf2WithSHA256Noinput(t *testing.T) {
	const testLength = 10 // derived key length
	_, err := Pbkdf2WithSha256([]byte{}, []byte{}, testLength, testLength)
	if err == nil {
		t.Errorf("zero length must error")
	}
}

func TestPbkdf2ZeroDKLength(t *testing.T) {
	dk, err := Pbkdf2WithSha256([]byte(""), []byte(""), 0, 0)
	if !bytes.Equal(dk, []byte("")) || err == nil {
		t.Errorf("pbkdf2 with sha256 : expected got %x %x", []byte(""), dk)
	}
}

func TestPbkdf2MaxKeyLen(t *testing.T) {
	_, err := Pbkdf2WithSha256([]byte(""), []byte(""), 0, 65536)
	if err == nil {
		t.Errorf("pbkdf2 with key length 65536 expected error")
	}
}

func benchMark(pwdLen int, saltLen int, iter int, keyLen int) (time.Duration, error) {
	const maxPwdLen = 4096
	const maxSaltLen = 1024

	if pwdLen > maxPwdLen || saltLen > maxSaltLen {
		return 0, errors.New("invalid length")
	}
	p := make([]byte, pwdLen)
	salt := make([]byte, saltLen)
	if _, err := rand.Read(p); err != nil {
		return 0, err
	}
	if _, err := rand.Read(salt); err != nil {
		return 0, err
	}

	start := time.Now()
	dk, err := Pbkdf2WithSha256(p, salt, iter, keyLen)
	escaped := time.Since(start)
	if err != nil {
		return 0, err
	}
	if len(dk) != keyLen {
		return 0, errors.New("key length error")
	}
	return escaped, nil
}

func TestPbkdf2BenchMark1(t *testing.T) {
	const testLength = 4096
	const saltLen = 32
	const keyLen = 2048
	const iter = 20000
	escaped, err := benchMark(testLength, saltLen, iter, keyLen)
	if err != nil {
		t.Errorf("bench mark failed")
	}
	t.Logf("bench mark with 4096 salt 32 iterator 20000 time cost:%v", escaped)
	const maxTime int64 = 2000
	if escaped.Milliseconds() > maxTime {
		t.Errorf("benchmark sha256 cost too long time")
	}
}

func TestPbkdf2BenchMark2(t *testing.T) {
	const testLength = 4096
	const saltLen = 64
	const keyLen = 1024
	const iter = 20000
	escaped, err := benchMark(testLength, saltLen, iter, keyLen)
	if err != nil {
		t.Errorf("bench mark failed")
	}
	t.Logf("bench mark with 4096 key iterator 10000 time cost:%v", escaped)
	const maxTime int64 = 2000
	if escaped.Milliseconds() > maxTime {
		t.Errorf("bench mark cost too long time")
	}
}
