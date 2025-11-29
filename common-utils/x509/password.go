//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package x509 provides the capability of x509.
package x509

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"regexp"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/terminal"
)

const (
	capacity       = 64
	minCount       = 2
	maxPasswordLen = 256
)

// GetRandomPass produce the new password.
// Password is randomly generated and encoded by base64.
// If the encoded password only contains one single type character (lower case, upper case or digit),
// it will be abandoned.
func GetRandomPass() ([]byte, error) {
	k := make([]byte, byteSize, byteSize)
	if _, err := rand.Read(k); err != nil {
		hwlog.RunLog.Error("get random words failed")
		return nil, err
	}
	length := base64.RawStdEncoding.EncodedLen(byteSize)
	if length > capacity || length < byteSize {
		hwlog.RunLog.Warn("the length of slice is abnormal")
	}
	dst := make([]byte, length, length)
	base64.RawStdEncoding.Encode(dst, k)
	var checkRes int
	regx := []string{"[A-Z]+", "[a-z]+", "[0-9]+", "[+/=]+"}
	for _, r := range regx {
		if res, err := regexp.Match(r, dst); err != nil || !res {
			continue
		}
		checkRes++
	}
	if checkRes < minCount {
		return nil, errors.New("the password is to simple,please retry")
	}
	return dst, nil
}

// ReadPassWd scan the screen and input the password info
func ReadPassWd() ([]byte, error) {
	fmt.Print("Enter Private Key Password: ")
	bytePassword, err := terminal.ReadPassword(0, maxPasswordLen)
	if err != nil {
		return nil, errors.New("program error")
	}
	if len(bytePassword) > maxLen {
		return nil, errors.New("input too long")
	}
	return bytePassword, nil
}

// ParsePrivateKeyWithPassword  decode the private key
func ParsePrivateKeyWithPassword(keyBytes []byte, pd []byte) (*pem.Block, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("decode key file failed")
	}
	buf := block.Bytes
	pemBlock := pem.Block{
		Type:    block.Type,
		Headers: nil,
		Bytes:   buf,
	}

	if !x509.IsEncryptedPEMBlock(block) {
		hwlog.RunLog.Warn("detect that you provided private key is not encrypted")
		return &pemBlock, nil
	}
	var err error
	newPd := pd
	if len(newPd) == 0 {
		newPd, err = ReadPassWd()
	}
	if err != nil {
		return nil, err
	}
	defer PaddingAndCleanSlice(newPd)
	buf, err = x509.DecryptPEMBlock(block, newPd)
	if err == nil {
		pemBlock.Bytes = buf
		return &pemBlock, nil
	}
	if err == x509.IncorrectPasswordError {
		return nil, err
	}
	return nil, errors.New("cannot decode encrypted private keys")
}

// OverridePassWdFile override password file with 0,1,random and then write new data
func OverridePassWdFile(path string, data []byte, mode os.FileMode) error {
	// Override with zero
	overrideByte := make([]byte, byteSize*maxLen, byteSize*maxLen)
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	for i := range overrideByte {
		overrideByte[i] = 0xff
	}
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	if _, err := rand.Read(overrideByte); err != nil {
		return errors.New("get random words failed")
	}
	if err := write(path, overrideByte, mode); err != nil {
		return err
	}
	if err := write(path, data, mode); err != nil {
		return err
	}
	return nil
}

// DecryptPrivateKeyWithPd  decrypt Private key By password
func DecryptPrivateKeyWithPd(keyFile, keyBkp string, passwd []byte) (*pem.Block, error) {
	keyInstance, err := NewBKPInstance(nil, keyFile, keyBkp)
	if err != nil {
		return nil, err
	}
	keyBytes, err := keyInstance.ReadFromDisk(fileutils.Mode600, true)
	if err != nil {
		return nil, err
	}
	defer PaddingAndCleanSlice(keyBytes)
	block, err := ParsePrivateKeyWithPassword(keyBytes, passwd)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// PaddingAndCleanSlice fill slice with zero
func PaddingAndCleanSlice(pd []byte) {
	for i := range pd {
		pd[i] = 0
	}
	pd = nil
}
