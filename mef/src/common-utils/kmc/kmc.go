// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package kmc interface
package kmc

import (
	"errors"
	"syscall"
	"unsafe"
)

// InitConfig KMC initialization parameter configuration
type InitConfig struct {
	PrimaryKeyStoreFile string // PrimaryKeyStoreFile primary key store file
	StandbyKeyStoreFile string // StandbyKeyStoreFile standby key store file
	DomainCount         int    // DomainCount domain count
	ProcLockPerm        int    // ProcLockPerm process lock permissions
	SdpAlgId            int    // SdpAlgId SDP interface encryption algorithm ID
	HmacAlgId           int    // HmacAlgId HMAC interface algorithm ID
	SemKey              int    // SemKey Semaphore key name
	WorkKeyIter         int    // WorkKeyIter pbkdf2 iteration time for work key
	RootKeyIter         int    // RootKeyIter pbkdf2 iteration time for work key
	Role                int    // Role role
	LogLevel            int    // Kmc Log level
}

// Context defines Context
type Context struct {
	ctx unsafe.Pointer
}

const (
	domaincount = 1
	// ProcLockPerm process lock permissions
	ProcLockPerm = 0600

	// default semkey defined in kmc-ext
	defaultProcSemKey = 0x20161227

	// minimum semkey defined in kmc-ext
	minProcSemKey = 0x20161111

	// maximum semkey defined in kmc-ext
	maxProcSemKey = 0x20169999

	// Aes256gcm AES256-GCM
	Aes256gcm = 9
	// HmacSha256 Hmac-Sha256
	HmacSha256 = 2052

	defaultPrimaryKey = "/etc/mindx-dl/kmc_primary_store/master.ks"
	defaultStandbyKey = "/etc/mindx-dl/.config/backup.ks"

	defaultSaltLen = 16
	// RoleMaster [KMC] Master role
	RoleMaster = 1
	// Disable log level disable
	Disable = 0
	// Error log level error
	Error = 1
	// Warn log level warn
	Warn = 2
	// Info log level info
	Info = 3
	// Debug log level debug
	Debug = 4
	// Trace log level trace
	Trace = 5
)

var (
	kmcInstance Context
)

func semget(semkey uintptr) error {
	const defaultPemission = 00600
	_, _, errno := syscall.Syscall(syscall.SYS_SEMGET, semkey, 1, defaultPemission)
	if errno != 0 && errno != syscall.ENOENT {
		return errors.New("semget failed")
	}
	return nil
}

func getSemKey() int {
	for k := minProcSemKey; k < maxProcSemKey; k++ {
		if err := semget(uintptr(k)); err != nil {
			continue
		}
		return k
	}
	return defaultProcSemKey
}

// NewKmcInitConfig Initial configuration
func NewKmcInitConfig() *InitConfig {
	return &InitConfig{
		PrimaryKeyStoreFile: defaultPrimaryKey,
		StandbyKeyStoreFile: defaultStandbyKey,
		DomainCount:         domaincount,
		Role:                RoleMaster,
		ProcLockPerm:        ProcLockPerm,
		SdpAlgId:            Aes256gcm,
		HmacAlgId:           HmacSha256,
		SemKey:              getSemKey(),
		LogLevel:            Info,
	}
}

// Initialize initialize KMC instance
func Initialize(sdpAlgID int, primaryKey, standbyKey string) error {
	return nil
}

// Finalize  finalize kmc instance
func Finalize() error {
	return kmcInstance.KeFinalizeEx()
}

// Encrypt by domain id
func Encrypt(domainID uint, data []byte) ([]byte, error) {
	return data, nil
}

// Decrypt by domain id
func Decrypt(domainID uint, cipherText []byte) ([]byte, error) {
	return cipherText, nil
}

// KeInitializeEx Initialization
func KeInitializeEx(config *InitConfig) (Context, error) {
	ctx := Context{ctx: unsafe.Pointer(nil)}
	return ctx, nil
}

// KeFinalizeEx Free
func (ctx *Context) KeFinalizeEx() error {
	return nil
}
