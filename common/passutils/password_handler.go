// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package passutils this package is for handle password check and has
package passutils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"

	"golang.org/x/crypto/pbkdf2"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

const (
	// SaltLength salt length
	SaltLength = 16
	// BytesOfEncryptedString the bytes of encrypt string
	BytesOfEncryptedString = 32
	// IterCount iteration count
	IterCount = 10000
)

var hashFunc = sha256.New

// CheckPassWord 校验密码复杂度等
func CheckPassWord(userName string, passWord *string) error {
	if err := checkPassWordForPattern(userName, passWord); err != nil {
		return err
	}

	return CheckPassWordComplexity(passWord)
}

// CheckPassWordComplexity check password complexity
func CheckPassWordComplexity(b *string) error {
	complexCheckRegexArr := []string{
		common.LowercaseCharactersRegex,
		common.UppercaseCharactersRegex,
		common.BaseNumberRegex,
		common.SpecialCharactersRegex,
	}
	complexCount := 0
	for _, pattern := range complexCheckRegexArr {
		if matched, err := regexp.MatchString(pattern, *b); matched && err == nil {
			complexCount++
		}
	}

	if complexCount < common.MinComplexCount {
		return errors.New("password complex not meet the requirement")
	}
	return nil
}

func checkPassWordForPattern(userName string, passWord *string) error {
	if matched, err := regexp.MatchString(common.PassWordRegex, *passWord); err != nil || !matched {
		return errors.New("password not meet requirement")
	}
	if userName == *passWord {
		return errors.New("password cannot equals username")
	}
	if utils.ReverseString(userName) == *passWord {
		return errors.New("password cannot equal reversed username")
	}
	return nil
}

// ComparePassword 比较混淆后的新旧密码
func ComparePassword(newPassword *string, hashVal string, salt string) bool {
	saltByte, err := convertStringToByteArr(salt)
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}
	encryptPassWord := getEncryptedStrBySalt(newPassword, saltByte)
	return hashVal == encryptPassWord
}

// GetEncryptPassword 获取混淆后的密码
func GetEncryptPassword(plainText *string) (string, string, error) {
	salt, err := getSafeRandomBytes(SaltLength)
	if err != nil {
		return "", "", err
	}
	cipherText := pbkdf2.Key([]byte(*plainText), salt, IterCount, BytesOfEncryptedString, hashFunc)
	return convertByteArrToString(cipherText), convertByteArrToString(salt), nil
}

func getEncryptedStrBySalt(plainText *string, salt []byte) string {
	cipherText := pbkdf2.Key([]byte(*plainText), salt, IterCount, BytesOfEncryptedString, hashFunc)
	return convertByteArrToString(cipherText)
}

func getSafeRandomBytes(saltLen int) ([]byte, error) {
	if saltLen <= 0 {
		saltLen = SaltLength
	}
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func convertByteArrToString(b []byte) string {
	return hex.EncodeToString(b)
}

func convertStringToByteArr(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
