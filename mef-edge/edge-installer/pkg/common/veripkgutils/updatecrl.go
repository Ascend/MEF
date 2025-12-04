// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package veripkgutils this file for updating crl
package veripkgutils

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindx/mef/common/cmsverify"

	"edge-installer/pkg/common/constants"
)

const (
	crlExpireRet   = 0x88200103
	maxCrlSizeInMb = 10
)

// PrepareVerifyCrl prepare crl to verify package
func PrepareVerifyCrl(newCrl string) (bool, string, error) {
	// when two input parameters are the same, the function can be used to check whether the CRL file is valid
	newCrlStatus, err := cmsverify.CheckCrl(newCrl)
	if err != nil {
		fmt.Println("new crl file is invalid")
		hwlog.RunLog.Errorf("new crl file is invalid, %v", err)
		return true, "", errors.New("new crl file is invalid")
	}
	if newCrlStatus != cmsverify.CompareSame && newCrlStatus != cmsverify.CrlExpiredOnly {
		fmt.Println("new crl file is invalid")
		hwlog.RunLog.Error("new crl file is invalid, inconsistency of the same certificate")
		return true, "", errors.New("new crl file is invalid")
	}
	if newCrlStatus == cmsverify.CrlExpiredOnly {
		fmt.Println("new crl has expired. check it manually")
		hwlog.RunLog.Warn("new crl has expired. check it manually")
	}

	if _, err = fileutils.RealFileCheck(
		constants.CrlOnDevicePath, true, false, maxCrlSizeInMb); err != nil {
		hwlog.RunLog.Warnf("check file [%s] failed, error: %v", constants.CrlOnDevicePath, err)
		return true, newCrl, nil
	}
	if err = fileutils.SetPathPermission(constants.CrlOnDevicePath, fileutils.Mode600, false,
		false); err != nil {
		hwlog.RunLog.Warnf("set crl permission failed, error: %v", err)
		return true, newCrl, nil
	}
	compareStatus, err := cmsverify.CheckCrl(constants.CrlOnDevicePath)
	if err != nil || (compareStatus != cmsverify.CompareSame && compareStatus != cmsverify.CrlExpiredOnly) {
		hwlog.RunLog.Warnf("the local crl is invalid, use crl in software package to verify")
		return true, newCrl, nil
	}

	return compareCrls(newCrl, constants.CrlOnDevicePath)
}

func compareCrls(crlToUpdate, crlOnDevice string) (bool, string, error) {
	if crlToUpdate == "" || crlOnDevice == "" {
		hwlog.RunLog.Error("crl is invalid")
		return false, "", errors.New("crl is invalid")
	}

	var compareRes cmsverify.CrlCompareStatus
	var err error
	needUpdateCrl := true
	verifyCrl := crlToUpdate
	compareRes, err = cmsverify.CompareCrls(crlToUpdate, crlOnDevice)
	if err != nil {
		hwlog.RunLog.Errorf("compare crls failed, error: %v", err)
		return false, "", errors.New("compare crls failed")
	}

	switch int(compareRes) {
	case constants.CompareSame:
		needUpdateCrl = false
		verifyCrl = crlOnDevice
		hwlog.RunLog.Info("the software package crl file is the same as the local crl file, " +
			"use the local crl file to verify and no update local crl file required")
	case constants.CompareNew:
		hwlog.RunLog.Info("the software package crl file is newer than the local crl file, " +
			"use software package crl file to verify and update local crl file")
	case constants.CompareOld:
		needUpdateCrl = false
		verifyCrl = crlOnDevice
		hwlog.RunLog.Info("the software package crl file is older than the local crl file, " +
			"use the local crl file to verify and no update local crl file required")
	default:
		hwlog.RunLog.Error("compare local crl file and the software package crl file failed, " +
			"use software package crl file to verify and update local crl file")
	}

	return needUpdateCrl, verifyCrl, nil
}

// UpdateLocalCrl update local crl file to verify crl
func UpdateLocalCrl(verifyCrl string) error {
	crlOnDeviceDir := filepath.Dir(constants.CrlOnDevicePath)
	if err := fileutils.CreateDir(crlOnDeviceDir, constants.Mode755); err != nil {
		hwlog.RunLog.Errorf("create dir [%s] failed, error: %v", crlOnDeviceDir, err)
		return fmt.Errorf("create dir [%s] failed", crlOnDeviceDir)
	}
	if _, err := fileutils.RealDirCheck(crlOnDeviceDir, true, false); err != nil {
		hwlog.RunLog.Errorf("check dir [%s] failed, error: %v", crlOnDeviceDir, err)
		return fmt.Errorf("check dir [%s] failed", crlOnDeviceDir)
	}

	if err := fileutils.CopyFile(verifyCrl, constants.CrlOnDevicePath); err != nil {
		hwlog.RunLog.Errorf("copy crl file to dir [%s] failed, error: %v", crlOnDeviceDir, err)
		return fmt.Errorf("copy crl file to dir [%s] failed", crlOnDeviceDir)
	}
	if err := fileutils.SetPathPermission(constants.CrlOnDevicePath, constants.Mode600, false,
		false); err != nil {
		hwlog.RunLog.Errorf("set new crl permission failed, error: %v", err)
		return errors.New("set new crl permission failed")
	}

	return nil
}
