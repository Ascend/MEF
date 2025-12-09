// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handler for
package handler

import "huawei.com/mindx/common/modulemgr/model"

// HandleBase handle base
type HandleBase interface {
	Handle(msg *model.Message) error
}

// PostHandleBase post handle base
type PostHandleBase interface {
	Parse(msg *model.Message) error
	Check(msg *model.Message) error
	Handle(msg *model.Message) error
	PrintOpLogOk()
	PrintOpLogFail()
}
