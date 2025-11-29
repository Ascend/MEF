// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
