// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build !MEFEdge_A500

// Package util
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/path/pathmgr"
)

const (
	regMatchCount = 2
	dmiDecode     = "dmidecode"
	dmiType       = "-t1"
	snRegPattern  = "Serial Number: (.*)"
)

// GetSerialNumber get serial number
func GetSerialNumber(installRootDir string) (string, error) {
	sn, err := getA500Sn()
	if err == nil {
		return sn, nil
	}
	sn, err = getA500ProSn()
	if err == nil {
		return sn, nil
	}
	sn = getSnFromFile(installRootDir)
	if sn != "" {
		return sn, nil
	}
	return GetUuid()
}

func getA500ProSn() (string, error) {
	output, err := envutils.RunCommand(dmiDecode, envutils.DefCmdTimeoutSec, dmiType)
	if err != nil {
		return "", fmt.Errorf("get a500pro serial number failed, error: %v", err)
	}
	compileReg := regexp.MustCompile(snRegPattern)
	matches := compileReg.FindStringSubmatch(output)
	if len(matches) < regMatchCount {
		return "", errors.New("get a500pro serial number failed, error: serial number not found")
	}
	return matches[1], nil
}

func getSnFromFile(installRootDir string) string {
	snFile := pathmgr.NewConfigPathMgr(installRootDir).GetSnPath()
	data, err := fileutils.LoadFile(snFile)
	if err != nil {
		return ""
	}

	snMap := make(map[string]string)
	if err = json.Unmarshal(data, &snMap); err != nil {
		return ""
	}
	sn, ok := snMap["serialNumber"]
	if !ok {
		return ""
	}

	snChecker := checker.GetRegChecker("", `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true)
	ret := snChecker.Check(sn)
	if !ret.Result {
		return ""
	}

	return sn
}
