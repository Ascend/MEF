// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package kmc

// define the constant that used in kmc update and key re-encrypt
const (
	KeySuffix = ".key"
)

// ReEncryptParam is the struct that defines how a path should be re-encrypted after kmc-updating
type ReEncryptParam struct {
	Path       string
	SuffixList []string
}

// UpdateKmcTask is the struct for update kmc task which is to update single component's kmc key
type UpdateKmcTask struct {
	ReEncryptParamList []ReEncryptParam
	Ctx                *Context
}

// ManualUpdateKmcTask is the struct to manually update kmc keys
type ManualUpdateKmcTask struct {
	UpdateKmcTask
}

// RunTask is the main func to start manually update kmc task
func (muk *ManualUpdateKmcTask) RunTask() error {
	return nil
}
