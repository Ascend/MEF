// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cmsverify to verify package valid, contains verify cms file and compare crls
package cmsverify

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
	// CrlExpiredOnly the CMSCBB_ERR_PKI_CRL_HAS_EXPIRED code from cms, means crl is corrected but expired
	CrlExpiredOnly CrlCompareStatus = 0x88200103
)

// CompareCrls compare the update signed time of them
func CompareCrls(crlToUpdate, crlOnDevice string) (CrlCompareStatus, error) {
	return CompareSame, nil
}

// CheckCrl check if the imported crl is valid
func CheckCrl(crlPath string) (CrlCompareStatus, error) {
	return CompareSame, nil
}

// VerifyPackage verify package valid
func VerifyPackage(crlName, cmsName, fileName string) error {
	return nil
}
