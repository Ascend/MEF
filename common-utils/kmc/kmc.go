// Copyright(C) Huawei Technologies Co.,Ltd. 2022-2022. All rights reserved.

// Package kmc interface
package kmc

// #cgo CFLAGS: -I./include -Wall -Wno-unused-function  -fstack-protector-all -fPIE -fPIC
// #cgo LDFLAGS: -ldl -Wl,-z,relro -Wl,-z,noexecstack -fPIE
// #include <stdlib.h>
// #include <string.h>
// #include "kmc.h"
// extern void goLoggerCallback(void *ctx,LogLevel level, char *msg);
import "C"

import (
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
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
	// KeyStatusInactive  set key inactive
	KeyStatusInactive = 0
	// KeyStatusActive   set key active, can not be delete when active
	KeyStatusActive = 1
	// KeyStatusTobeActive intermediate status
	KeyStatusTobeActive = 2
	domaincount         = 1
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
	// DefaultkeyLifetimeDays default key lifeime
	DefaultkeyLifetimeDays = 180
	minKeyLifetimeDays     = 0
	maxkeyLifetimeDays     = 180
	maxKeyStorePathLength  = 1024

	defaultPrimaryKey   = "/etc/mindx-dl/kmc_primary_store/master.ks"
	defaultStandbyKey   = "/etc/mindx-dl/.config/backup.ks"
	minKeyIterationTime = 10000 // key派生算法pbkef2最低迭代次数
	maxSaltLen          = 128
	minSaltLen          = 16
	defaultSaltLen      = 16
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
	kmcNeedInit     bool = true
	kmcSaltLen      uint = defaultSaltLen
	kmcInstance     Context
	keyLifetimeDays int
	logLevelDict    = map[int]C.LogLevel{
		Disable: C.LOG_DISABLE,
		Error:   C.LOG_ERROR,
		Warn:    C.LOG_WARN,
		Info:    C.LOG_INFO,
		Debug:   C.LOG_DEBUG,
		Trace:   C.LOG_TRACE,
	}
)

// convertToKmcInitConfig Configuration file in go format, converted to C format
func convertToKmcInitConfig(config *InitConfig) (*C.KmcConfig, error) {
	var err error
	var kmcConfig C.KmcConfig
	kmcConfig.primaryKeyStoreFile, err = convertToCCharArray(filepath.FromSlash(config.PrimaryKeyStoreFile))
	if err != nil {
		return nil, err
	}
	kmcConfig.standbyKeyStoreFile, err = convertToCCharArray(filepath.FromSlash(config.StandbyKeyStoreFile))
	if err != nil {
		return nil, err
	}
	kmcConfig.domainCount = C.int(config.DomainCount)
	kmcConfig.role = C.int(config.Role)
	kmcConfig.procLockPerm = C.int(config.ProcLockPerm)
	kmcConfig.sdpAlgId = C.int(config.SdpAlgId)
	kmcConfig.hmacAlgId = C.int(config.HmacAlgId)
	kmcConfig.semKey = C.int(config.SemKey)

	if config.WorkKeyIter < minKeyIterationTime {
		kmcConfig.workKeyIter = minKeyIterationTime
	} else {
		kmcConfig.workKeyIter = C.int(config.WorkKeyIter)
	}
	if config.RootKeyIter < minKeyIterationTime {
		kmcConfig.rootKeyIter = minKeyIterationTime
	} else {
		kmcConfig.rootKeyIter = C.int(config.RootKeyIter)
	}

	return &kmcConfig, nil
}

// convertToCCharArray Adapt to the path format of the KMC configuration file, char[] array
func convertToCCharArray(goStr string) ([C.SEC_PATH_MAX]C.char, error) {
	var cChar [C.SEC_PATH_MAX]C.char

	// reserve 1 byte for the null terminator
	maxCopyLen := len(cChar) - 1
	if len(goStr) > maxCopyLen {
		return cChar, fmt.Errorf("convertToCCharArray failed, the length of the path exceeds SEC_PATH_MAX[%v]",
			C.SEC_PATH_MAX)
	}

	for i := 0; i < len(goStr); i++ {
		cChar[i] = C.char(goStr[i])
	}

	// explicitly add a null terminator
	cChar[len(goStr)] = 0

	return cChar, nil
}

// SetSaltLen Set salt length in MK derive algorithm pbkdf2 in KMC
func SetSaltLen(saltLen uint) {
	if saltLen < minSaltLen || saltLen > maxSaltLen {
		return
	}
	kmcSaltLen = saltLen
}

// KeSetLoggerLevel Set log level
func KeSetLoggerLevel(logLevel int) *KeKmcError {
	if logLevel < 0 || logLevel > Trace {
		return paramCheckErr
	}
	if ok := C.KeSetLoggerLevel(logLevelDict[logLevel]); int(ok) != 0 {
		return NewKmcError(int(ok), "set logger level error")
	}
	return nil
}

// KeSetLoggerCallBack Set log level
func KeSetLoggerCallBack() *KeKmcError {
	logger = &hwlog.LoggerAdaptor{}
	if ok := C.KeSetLoggerCallbackEx(C.LoggerCallbackEx(C.goLoggerCallback)); int(ok) != 0 {
		return NewKmcError(int(ok), "set logger callback error")
	}
	return nil
}

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

func initDynamicLib() error {
	if kmcNeedInit {
		if ret := C.InitKMC(); ret != C.KE_SUCCESS {
			return kmcLibErr
		}
		keyLifetimeDays = DefaultkeyLifetimeDays
		kmcNeedInit = false
	}
	return nil
}

// Initialize  initialize KMC instance
func Initialize(sdpAlgID int, primaryKey, standbyKey string) error {
	if kmcInstance.ctx != nil {
		return nil
	}
	cfg := NewKmcInitConfig()
	if primaryKey != "" && standbyKey != "" {
		cfg.PrimaryKeyStoreFile = primaryKey
		cfg.StandbyKeyStoreFile = standbyKey
	}
	if _, err := fileutils.CheckOriginPath(cfg.PrimaryKeyStoreFile); err != nil {
		return NewKmcError(-1, "master:"+err.Error())
	}
	if _, err := fileutils.CheckOriginPath(cfg.StandbyKeyStoreFile); err != nil {
		return NewKmcError(-1, "bask:"+err.Error())
	}
	cfg.SdpAlgId = sdpAlgID
	var err error
	kmcInstance, err = KeInitializeEx(cfg)
	return err
}

// Finalize  finalize kmc instance
func Finalize() error {
	return kmcInstance.KeFinalizeEx()
}

// Encrypt Encrypt by domain id
func Encrypt(domainID uint, data []byte) ([]byte, error) {
	if err := kmcInstance.KeCheckAndUpdateMkEx(domainID, DefaultkeyLifetimeDays-keyLifetimeDays); err != nil {
		hwlog.RunLog.Warnf("update MK failed %v", err)
	}
	return kmcInstance.KeEncryptByDomainEx(domainID, data)
}

// Decrypt Decrypt by domain id
func Decrypt(domainID uint, cipherText []byte) ([]byte, error) {
	return kmcInstance.KeDecryptByDomainEx(domainID, cipherText)
}

// UpdateLifetimeDays  update the key update time
func UpdateLifetimeDays(lifetime int) error {
	if lifetime <= minKeyLifetimeDays || lifetime > maxkeyLifetimeDays {
		return paramCheckErr
	}
	keyLifetimeDays = lifetime
	return nil
}

func goBytes(cText *C.char, cLen C.int) []byte {
	return C.GoBytes(unsafe.Pointer(cText), cLen)
}

func freeAndCleanCharp(p *C.char, length C.int) {
	if p != nil {
		bs := unsafe.Slice((*byte)(unsafe.Pointer(p)), uint(length))
		for i := 0; i < int(length); i++ {
			bs[i] = 0
		}
		C.free(unsafe.Pointer(p))
	}
}

// KeInitializeEx Initialization
func KeInitializeEx(config *InitConfig) (Context, error) {
	ctx := Context{ctx: unsafe.Pointer(nil)}
	if err := initDynamicLib(); err != nil {
		return ctx, err
	}
	if (len(config.PrimaryKeyStoreFile) > maxKeyStorePathLength) ||
		(len(config.StandbyKeyStoreFile) > maxKeyStorePathLength) {
		return ctx, paramCheckErr
	}
	if err := KeSetLoggerLevel(config.LogLevel); err != nil {
		return ctx, err
	}
	if err := KeSetLoggerCallBack(); err != nil {
		return ctx, err
	}
	kmcConfig, err := convertToKmcInitConfig(config)
	if err != nil {
		return ctx, err
	}
	var kmcConfigEx C.KmcConfigEx
	kmcConfigEx.enableHw = 0
	kmcConfigEx.kmcConfig = *kmcConfig
	if ok := C.KeInitializeEx(&kmcConfigEx, &ctx.ctx,
		C.uint(kmcSaltLen)); int(ok) != 0 {
		return ctx, NewKmcError(int(ok), "KeInitializeEx failed")
	}
	return ctx, nil
}

// KeActiveNewKeyEx active key by domain ID
func (ctx Context) KeActiveNewKeyEx(domainID uint) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if ret := C.KeActiveNewKeyEx(ctx.ctx, C.uint(domainID)); int(ret) != 0 {
		return NewKmcError(int(ret), "active new key failed")
	}
	return nil
}

// KeRefreshMkMaskEx refresh master key mask
func (ctx Context) KeRefreshMkMaskEx() error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if ok := C.KeRefreshMkMaskEx(ctx.ctx); int(ok) != 0 {
		return NewKmcError(int(ok), "refresh master key mask failed")
	}
	return nil
}

// KeGeneratedKeyAndGetIDEx Import the master key externally and get the appropriate id
func (ctx Context) KeGeneratedKeyAndGetIDEx(domainId uint) (uint, error) {
	if ctx.ctx == nil {
		return 0, kmcNotInitErr
	}
	if ret := int(C.KeActiveNewKeyEx(ctx.ctx, C.uint(domainId))); ret != 0 {
		return 0, NewKmcError(ret, "active new key failed")
	}
	return ctx.KeGetMaxMkIDEx(domainId)
}

// KeGetMaxMkIDEx return the max key id
func (ctx Context) KeGetMaxMkIDEx(domainId uint) (uint, error) {
	if ctx.ctx == nil {
		return 0, kmcNotInitErr
	}
	var maxKeyId C.uint
	if ok := int(C.KeGetMaxMkIDEx(ctx.ctx, C.uint(domainId), &maxKeyId)); ok != 0 {
		return 0, NewKmcError(ok, "get max key id failed")
	}
	return uint(maxKeyId), nil
}

// KeRegisterByteKeyEx External import master key
func (ctx Context) KeRegisterByteKeyEx(domainId uint, keyId uint, key []byte) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if len(key) == 0 {
		return paramCheckErr
	}
	ret := C.KeRegisterByteKeyEx(ctx.ctx, C.uint(domainId), C.uint(keyId),
		(*C.uchar)(unsafe.Pointer(&key[0])), C.int(len(key)))
	if int(ret) != 0 {
		return NewKmcError(int(ret), "outside init kmc failed")
	}
	return nil
}

// KeEncryptByDomainEx Encrypt
func (ctx Context) KeEncryptByDomainEx(domainId uint, plainText []byte) ([]byte, error) {
	if ctx.ctx == nil {
		return nil, kmcNotInitErr
	}
	if len(plainText) == 0 {
		return nil, paramCheckErr
	}
	cCipherText := (*C.char)(nil)
	cCipherTextLen := C.int(0)
	ok := C.KeEncryptByDomainEx(ctx.ctx, C.uint(domainId), (*C.char)(unsafe.Pointer(&plainText[0])),
		C.int(len(plainText)), &cCipherText, &cCipherTextLen)
	if int(ok) != 0 || cCipherText == nil {
		return nil, NewKmcError(int(ok), "encrypt failed")
	}
	defer C.free(unsafe.Pointer(cCipherText))
	return goBytes(cCipherText, cCipherTextLen), nil
}

// KeHmacByDomainV2Ex get hmac by domain ID
func (ctx Context) KeHmacByDomainV2Ex(domainId uint, plainText []byte) ([]byte, error) {
	if ctx.ctx == nil {
		return nil, kmcNotInitErr
	}
	if len(plainText) == 0 {
		return nil, paramCheckErr
	}
	hmacData := (*C.char)(nil)
	hmacDateLen := C.int(0)
	ok := C.KeHmacByDomainV2Ex(ctx.ctx, C.uint(domainId), (*C.char)(unsafe.Pointer(&plainText[0])),
		C.int(len(plainText)), &hmacData, &hmacDateLen)
	if int(ok) != 0 || hmacData == nil {
		return nil, NewKmcError(int(ok), "calculate hmac failed")
	}
	defer C.free(unsafe.Pointer(hmacData))
	return goBytes(hmacData, hmacDateLen), nil
}

// KeHmacVerifyByDomainEx verify hamc
func (ctx Context) KeHmacVerifyByDomainEx(domainId uint, plainText, hmacData []byte) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if len(plainText) == 0 || len(hmacData) == 0 {
		return paramCheckErr
	}
	ok := C.KeHmacVerifyByDomainEx(ctx.ctx, C.uint(domainId), (*C.char)(unsafe.Pointer(&plainText[0])),
		C.int(len(plainText)), (*C.char)(unsafe.Pointer(&hmacData[0])), C.int(len(hmacData)))
	if int(ok) != 0 {
		return NewKmcError(int(ok), "verify hamc failed")
	}
	return nil
}

// KeDecryptByDomainEx Decrypt
func (ctx Context) KeDecryptByDomainEx(domainId uint, cipherText []byte) ([]byte, error) {
	if ctx.ctx == nil {
		return nil, kmcNotInitErr
	}
	if len(cipherText) == 0 {
		return nil, paramCheckErr
	}
	cPlainText := (*C.char)(nil)
	cPlainTextLen := C.int(0)
	ok := C.KeDecryptByDomainEx(ctx.ctx, C.uint(domainId), (*C.char)(unsafe.Pointer(&cipherText[0])),
		C.int(len(cipherText)), &cPlainText, &cPlainTextLen)
	if int(ok) != 0 || cPlainText == nil {
		return nil, NewKmcError(int(ok), "decrypt failed")
	}
	defer freeAndCleanCharp(cPlainText, cPlainTextLen)
	return goBytes(cPlainText, cPlainTextLen), nil
}

// KeGetCipherDataLenEx Get ciphertext length
func (ctx Context) KeGetCipherDataLenEx(plainTextLen int) (int, error) {
	if ctx.ctx == nil {
		return 0, kmcNotInitErr
	}
	if plainTextLen < 0 {
		return 0, paramCheckErr
	}
	var cipherTextLen C.int
	if ok := C.KeGetCipherDataLenEx(ctx.ctx, C.int(plainTextLen), &cipherTextLen); int(ok) != 0 {
		return 0, NewKmcError(int(ok), "get the length of cipher data failed")
	}
	return int(cipherTextLen), nil
}

// KeGetKeyByIDEx Export master key
func (ctx Context) KeGetKeyByIDEx(domainId uint, keyId uint, getAsBase64 bool) ([]byte, error) {
	if ctx.ctx == nil {
		return nil, kmcNotInitErr
	}
	plainKey := (*C.char)(nil)
	plainKeyLen := C.int(0)
	base64Ind := 0
	if getAsBase64 {
		base64Ind = 1
	}
	if ok := C.KeGetKeyByIDEx(ctx.ctx, C.uint(domainId), C.uint(keyId),
		&plainKey, &plainKeyLen, C.int(base64Ind)); int(ok) != 0 {
		return nil, NewKmcError(int(ok), "get the key by id failed")
	}
	defer freeAndCleanCharp(plainKey, plainKeyLen)
	return goBytes(plainKey, plainKeyLen), nil
}

// KeSetMkStatusEx set master key status
func (ctx Context) KeSetMkStatusEx(domainId uint, keyId uint, status uint8) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if (status != KeyStatusActive) && (status != KeyStatusInactive) && (status != KeyStatusTobeActive) {
		return paramCheckErr
	}
	if ok := C.KeSetMkStatusEx(ctx.ctx, C.uint(domainId), C.uint(keyId), C.uchar(status)); int(ok) != 0 {
		return NewKmcError(int(ok), "set mk status failed")
	}
	return nil
}

// KeRemoveKeyByIDEx Delete master key
func (ctx Context) KeRemoveKeyByIDEx(domainId uint, keyId uint) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if ok := C.KeRemoveKeyByIDEx(ctx.ctx, C.uint(domainId), C.uint(keyId)); int(ok) != 0 {
		return NewKmcError(int(ok), "remove the key by id failed")
	}
	return nil
}

// KeCheckAndUpdateMkEx check and update mk
func (ctx Context) KeCheckAndUpdateMkEx(domainId uint, advanceDay int) error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if ok := C.KeCheckAndUpdateMkEx(ctx.ctx, C.uint(domainId), C.int(advanceDay)); int(ok) != 0 {
		return NewKmcError(int(ok), "check and update mk failed")
	}
	return nil
}

// UpdateRootKey update the root key
func (ctx Context) UpdateRootKey() error {
	if ctx.ctx == nil {
		return kmcNotInitErr
	}
	if ok := C.KeUpdateRootKeyEx(ctx.ctx); int(ok) != 0 {
		return NewKmcError(int(ok), " update root key failed")
	}
	return nil
}

// KeSecureEraseKeystoreEx secure erase keystore file
func (ctx *Context) KeSecureEraseKeystoreEx() error {
	if ctx == nil {
		return kmcNotInitErr
	}
	if ok := C.KeSecureEraseKeystoreEx(ctx.ctx); int(ok) != 0 {
		return NewKmcError(int(ok), "erase keystore failed")
	}
	return nil
}

// KeFinalizeEx Free
func (ctx *Context) KeFinalizeEx() error {
	if ctx == nil {
		return paramCheckErr
	}
	if ok := int(C.KeFinalizeEx(&ctx.ctx)); ok != 0 {
		return NewKmcError(ok, "finalize failed")
	}
	return nil
}
