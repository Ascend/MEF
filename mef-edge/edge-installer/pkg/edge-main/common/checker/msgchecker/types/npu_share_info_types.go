// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package types for npu share info
package types

// NpuSharingInfo [struct] to describe  npu sharing info
type NpuSharingInfo struct {
	NpuSharingEnabled *bool `json:"npu_sharing_enabled,omitempty" binding:"omitempty"`
}
