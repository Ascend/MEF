// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.

// Package utils this file for password handler
package utils

import (
	"bytes"
	"errors"
	"regexp"
	"unsafe"
)

const (
	lowercaseCharactersRegex = `[a-z]{1,}`
	uppercaseCharactersRegex = `[A-Z]{1,}`
	baseNumberRegex          = `[0-9]{1,}`
	specialCharactersRegex   = `[!\"#$%&'()*+,\-. /:;<=>?@[\\\]^_\x60{|}~]{1,}`
	passWordRegex            = `^[a-zA-Z0-9!\"#$%&'()*+,\-. /:;<=>?@[\\\]^_\x60{|}~]{8,64}$`
	minComplexCount          = 2
)

// CheckPassWordComplexity check password complexity
func CheckPassWordComplexity(s []byte) error {
	complexCheckRegexArr := []string{
		lowercaseCharactersRegex,
		uppercaseCharactersRegex,
		baseNumberRegex,
		specialCharactersRegex,
	}
	complexCount := 0
	for _, pattern := range complexCheckRegexArr {
		if matched, err := regexp.Match(pattern, s); matched && err == nil {
			complexCount++
		}
	}
	if complexCount < minComplexCount {
		return errors.New("password complex not meet the requirement")
	}
	return nil
}

// ValidatePassWord validate password
func ValidatePassWord(userName string, passWord []byte) error {
	if err := commonCheckForPassWord(userName, passWord); err != nil {
		return err
	}
	return CheckPassWordComplexity(passWord)
}

func commonCheckForPassWord(userName string, passWord []byte) error {
	if matched, err := regexp.Match(passWordRegex, passWord); err != nil || !matched {
		return errors.New("password not meet requirement")
	}
	var userNameByte []byte = []byte(userName)
	if bytes.Equal(userNameByte, passWord) {
		return errors.New("password cannot equals username")
	}
	var reverseUserName = ReverseString(userName)
	var reverseUserNameByte []byte = []byte(reverseUserName)
	if bytes.Equal(reverseUserNameByte, passWord) {
		return errors.New("password cannot equal reversed username")
	}
	return nil
}

// ClearSliceByteMemory clear slice byte memory
func ClearSliceByteMemory(sliceByte []byte) {
	for i := 0; i < len(sliceByte); i++ {
		sliceByte[i] = 0
	}
}

// ClearStringMemory clear string memory for sensitive data in 3rd party component
func ClearStringMemory(s string) {
	// 对于单个字符表示的字符串，go的runtime实现的字符串共享字符数据，位于一个统一的staticbytes数组中
	// 所以 禁止 清除他，否则程序后面可能异常，密码等敏感信息长度需要大于1
	if len(s) <= 1 {
		return
	}
	bs := *(*[]byte)(unsafe.Pointer(&s))
	for i := 0; i < len(bs); i++ {
		bs[i] = 0
	}
}
