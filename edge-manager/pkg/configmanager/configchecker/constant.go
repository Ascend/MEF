// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package configchecker constant
package configchecker

const (
	minHostPort = 1
	maxHostPort = 65535
	minPwdCount = 1
	maxPwdCount = 20
	nameReg     = "^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,254}[a-zA-Z0-9]){0,1}$"
	dnsReg      = "[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\\.?"
)
