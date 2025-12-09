// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cmsverify to verify package valid, contains verify cms file and compare crls
package cmsverify

// #cgo CFLAGS: -I./include -Wall -Wno-unused-function  -fstack-protector-all -fPIE -fPIC
// #cgo LDFLAGS: -L./lib -ldl -Wl,-z,relro -Wl,-z,noexecstack -lcms_verify -lsecurec -lcrypto
// #include <stdlib.h>
// #include <string.h>
// #include "securec.h"
// #include "cms_api.h"
// #include "cmscbb_cms_vrf.h"
import "C"
import (
	"fmt"
	"unsafe"

	"huawei.com/mindx/common/fileutils"
)

// CrlCompareStatus crl compare result status
type CrlCompareStatus int

const (
	// CompareSame two crls are same
	CompareSame CrlCompareStatus = 0
	// CompareNew crl to update signed time is newer
	CompareNew CrlCompareStatus = 1
	// CompareOld crl to update signed time is older
	CompareOld CrlCompareStatus = 2
	// CompareFailed can't compare or some other errors
	CompareFailed CrlCompareStatus = 3
)

func checkPara(crlName, cmsName, fileName string) error {
	for _, file := range []string{crlName, cmsName, fileName} {
		if _, err := fileutils.CheckOriginPath(file); err != nil {
			return fmt.Errorf("check file: [%s] path failed", file)
		}

		if !fileutils.IsExist(fileName) {
			return fmt.Errorf("file: [%s] is not exist", file)
		}
	}

	return nil
}

// CompareCrls compare the update signed time of them
func CompareCrls(crlToUpdate, crlOnDevice string) (CrlCompareStatus, error) {
	if _, err := fileutils.CheckOriginPath(crlToUpdate); err != nil {
		return CompareFailed, fmt.Errorf("check crl to update path: [%s] failed", crlToUpdate)
	}

	if _, err := fileutils.CheckOriginPath(crlOnDevice); err != nil {
		return CompareFailed, fmt.Errorf("check crl on device path: [%s] failed", crlOnDevice)
	}

	cTypeCrlToUpdateName := C.CString(crlToUpdate)
	cTypeCrlOnDeviceName := C.CString(crlOnDevice)
	defer func() {
		C.free(unsafe.Pointer(cTypeCrlToUpdateName))
		C.free(unsafe.Pointer(cTypeCrlOnDeviceName))
	}()

	var cTypeStats C.CmscbbCrlPeriodStat // compare result status
	retCompareCrls := C.CompareCrls(cTypeCrlToUpdateName, cTypeCrlOnDeviceName,
		(*C.CmscbbCrlPeriodStat)(unsafe.Pointer(&cTypeStats)))
	if retCompareCrls != 0 {
		return CompareFailed, fmt.Errorf("crl compare failed: %v", retCompareCrls)
	}

	crlCompareStatus := CrlCompareStatus(cTypeStats)
	return crlCompareStatus, nil
}

// VerifyPackage verify package valid
func VerifyPackage(crlName, cmsName, fileName string) error {
	if err := checkPara(crlName, cmsName, fileName); err != nil {
		return err
	}

	cTypeCrlName := C.CString(crlName)
	cTypeCmsName := C.CString(cmsName)
	cTypeFileName := C.CString(fileName)
	defer func() {
		C.free(unsafe.Pointer(cTypeCrlName))
		C.free(unsafe.Pointer(cTypeCmsName))
		C.free(unsafe.Pointer(cTypeFileName))
	}()

	retVerifyCms := C.VerifyCmsFile(cTypeCrlName, cTypeCmsName, cTypeFileName)
	if retVerifyCms != 0 {
		return fmt.Errorf("cms verify failed: %v", retVerifyCms)
	}

	return nil
}
