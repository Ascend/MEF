// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util to init util service
package util

// UpgradeSfwReq request of upgrading software
type UpgradeSfwReq struct {
	NodeIDs         []int64 `json:"nodeIDs"`
	SoftwareName    string  `json:"softwareName"`
	SoftwareVersion string  `json:"softwareVersion"`
}

// DownloadSfwReq request of downloading software
type DownloadSfwReq struct {
	NodeID          string `json:"nodeID"`
	SoftwareName    string `json:"softwareName"`
	SoftwareVersion string `json:"softwareVersion,omitempty"`
}
