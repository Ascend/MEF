// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package certutils

const (
	// priKeyLength private key length
	priKeyLength = 4096
	// validationYearCA root ca validate year
	validationYearCA = 10
	// validationYearCert service Cert validate year
	validationYearCert = 10
	// validationMonth Cert validate month
	validationMonth = 0
	// validationDay Cert validate day
	validationDay = 0
	// bigIntSize server_number
	bigIntSize = 2022
	// caCountry issue country
	caCountry = "CN"
	// caOrganization issue organization
	caOrganization = "Huawei"
	// caOrganizationalUnit issue unit
	caOrganizationalUnit = "Ascend"
	// caCommonName issue name
	caCommonName = "MEF"
	// pubCertType Cert type
	pubCertType = "CERTIFICATE"
	// privKeyType Cert key type
	privKeyType = "RSA PRIVATE KEY"
	// fileMode Cert file mode
	fileMode = 0600
)

// MEF-Center cert constant
const (
	DefaultNameSpace  = "default"
	DefaultSecretName = "image-pull-secret"
	CertSizeLimited   = 1024 * 1024
	SecretNotFound    = "not found"
)
