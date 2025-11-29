//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils provides the util func
package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

const (
	maxSliceSize = 65536
	maskLen      = 2
	// SplitFlag backup file splitflag
	SplitFlag = "\n</=--*^^||^^--*=/>"
)

// ReplacePrefix replace string with prefix
func ReplacePrefix(source, prefix string) string {
	if prefix == "" {
		prefix = "****"
	}
	if len(source) <= maskLen {
		return prefix
	}
	end := string([]rune(source)[maskLen:len(source)])
	return prefix + end
}

// MaskPrefix mask string prefix with ****
func MaskPrefix(source string) string {
	return ReplacePrefix(source, "")
}

// GetSha256Code return the sha256 hash bytes
func GetSha256Code(data []byte) []byte {
	hash256 := sha256.New()
	if _, err := hash256.Write(data); err != nil {
		fmt.Println(err)
		return nil
	}
	return hash256.Sum(nil)
}

// ReverseString reverse string
func ReverseString(s string) string {
	runes := []rune(s)
	for start, end := 0, len(runes)-1; start < end; start, end = start+1, end-1 {
		runes[start], runes[end] = runes[end], runes[start]
	}
	return string(runes)
}

// BinaryFormat cast binary data to string
func BinaryFormat(b []byte, minByteLen int) string {
	if minByteLen > maxSliceSize {
		minByteLen = maxSliceSize
	}
	bPadding := b
	if len(b) < minByteLen {
		bPadding = make([]byte, minByteLen)
		bPadding = append(bPadding[:minByteLen-len(b)], b...)
	}

	var hexStrings []string
	for _, ch := range bPadding {
		hexStrings = append(hexStrings, fmt.Sprintf("%02X", ch))
	}
	return strings.Join(hexStrings, ":")
}

// IsDigitString return string is all digit
func IsDigitString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}
