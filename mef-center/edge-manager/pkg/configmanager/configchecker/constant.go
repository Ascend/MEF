// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configchecker constant
package configchecker

const (
	minHostPort = 1
	maxHostPort = 65535
	minPwdCount = 8
	maxPwdCount = 20
	nameReg     = "^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,254}[a-zA-Z0-9]){0,1}$"
	domainReg   = "^[a-zA-Z0-9][a-zA-Z0-9.-]{1,61}[a-zA-Z0-9]$"
)
