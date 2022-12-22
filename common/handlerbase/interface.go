// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package handlerbase

import (
	"huawei.com/mindxedge/base/modulemanager/model"
)

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
