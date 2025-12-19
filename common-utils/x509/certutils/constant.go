// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package certutils

const (
	// priKeyLength private key length
	priKeyLength = 3072
	// validationYearCA root ca validate year
	validationYearCA = 10
	// validationYearCert service Cert validate year
	validationYearCert = 10
	// validationMonth Cert validate month
	validationMonth = 0
	// validationDay Cert validate day
	validationDay = 0
	// caCountry issue country
	caCountry = "CN"
	// caOrganization issue organization
	caOrganization = "Huawei"
	// caOrganizationalUnit issue unit
	caOrganizationalUnit = "CPL Ascend"
	// pubCertType Cert type
	pubCertType = "CERTIFICATE"
	pubCsrType  = "CERTIFICATE REQUEST"
	// privKeyType Cert key type
	privKeyType = "RSA PRIVATE KEY"
	// OneDayAgo for compatible with different time zone when issue cert
	OneDayAgo = "-24h"
)
