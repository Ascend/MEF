// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal node msg
package edgemsgmanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"huawei.com/mindxedge/base/common"
)

// SoftwareDownloadInfo content for download software
type SoftwareDownloadInfo struct {
	SerialNumbers []string     `json:"serialNumbers"`
	SoftwareName  string       `json:"softwareName"`
	DownloadInfo  DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo [struct] to software download info
type DownloadInfo struct {
	Package  string    `json:"package"`
	SignFile string    `json:"signFile,omitempty"`
	CrlFile  string    `json:"crlFile,omitempty"`
	UserName string    `json:"username"`
	Password *Password `json:"password"`
}

// UpgradeSoftwareReq update software
type UpgradeSoftwareReq struct {
	SerialNumbers []string `json:"serialNumbers"`
	SoftwareName  string   `json:"softwareName"`
}

// Password the password struct
type Password []byte

// MarshalJSON marshal password, err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (p Password) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	var isNotFirst bool
	for i := range p {
		if isNotFirst {
			buffer.WriteByte(',')
		}
		isNotFirst = true
		buffer.WriteString(strconv.Itoa(int(p[i])))
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshal password
func (p *Password) UnmarshalJSON(data []byte) error {
	var pArr []json.RawMessage
	if err := json.Unmarshal(data, &pArr); err != nil {
		return errors.New("unmarshal pwd data failed")
	}
	pBytes := make([]byte, len(pArr))
	for i := range pArr {
		num, err := strconv.ParseUint(string(pArr[i]), common.BaseHex, common.BitSize8)
		if err != nil {
			return errors.New("parse pwd to uint failed")
		}
		pBytes[i] = byte(num)
	}
	*p = pBytes
	return nil
}
