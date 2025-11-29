// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
