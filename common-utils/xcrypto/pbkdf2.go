// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"errors"
	"unsafe"
)

/*
#cgo LDFLAGS: -ldl
#include <stdio.h>
#include <dlfcn.h>

#define LIB_CRYPTO      "libcrypto.so"      // for common
#define LIB_CRYPTO_1_1  "libcrypto.so.1.1"  // for openssl 1.1
#define LIB_CRYPTO_1_0  "libcrypto.so.10"   // for openssl 1.0
#define FUNC_HASH_ALGO  "EVP_sha256"
#define FUNC_PBKDF2     "PKCS5_PBKDF2_HMAC"
#define PBKDF2_SUCCESS  (1)

typedef void* (*EVP_method)();

int (*pbkdf2)(const char *pass, int passlen,
              const unsigned char *salt, int saltlen, int iter,
              const void *digest, int keylen, unsigned char *out);

int pbkdf2_sha256(const char* pass, int passlen,
                  const unsigned char *salt, int saltlen,
                  int iter, int keylen, unsigned char *out)
{
    if (pass == NULL || salt == NULL || out == NULL) {
        return -1;
    }
    static void *handle = NULL;
    if (handle == NULL) {
        handle = dlopen(LIB_CRYPTO, RTLD_LAZY);
    }
    if (handle == NULL) {
        handle = dlopen(LIB_CRYPTO_1_1, RTLD_LAZY);
    }
    if (handle == NULL) {
        handle = dlopen(LIB_CRYPTO_1_0, RTLD_LAZY);
    }

    if (handle == NULL) {
        return -1;
    }

    EVP_method sha256 = dlsym(handle, "EVP_sha256");
    if (sha256 == NULL) {
        return -1;
    }
    pbkdf2 = dlsym(handle, "PKCS5_PBKDF2_HMAC");
    if (pbkdf2 == NULL) {
        return -1;
    }
    int ret = pbkdf2(pass, passlen, salt, saltlen, iter, sha256(), keylen, out);
    return ret == PBKDF2_SUCCESS ? 0 : -1;
}
*/
import "C"

const maxKeyLen = 10240

// Pbkdf2WithSha256 impletment of pbkdf2 algorithm with HMAC SHA256
func Pbkdf2WithSha256(pwd []byte, salt []byte, iter int, keyLen int) ([]byte, error) {
	if len(pwd) == 0 || len(salt) == 0 || keyLen > maxKeyLen {
		return nil, errors.New("invalid length")
	}
	out := make([]byte, keyLen)

	ret := C.pbkdf2_sha256((*C.char)(unsafe.Pointer(&pwd[0])), C.int(len(pwd)),
		(*C.uchar)(unsafe.Pointer(&salt[0])), C.int(len(salt)),
		C.int(iter), C.int(keyLen),
		(*C.uchar)(unsafe.Pointer(&out[0])))
	if int(ret) != 0 {
		return nil, errors.New("pbkdf2 failed")
	}
	return out, nil
}
