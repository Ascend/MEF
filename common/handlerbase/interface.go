package handlerbase

import (
	"huawei.com/mindxedge/base/modulemanager/model"
)

type HandleBase interface {
	Handle(msg *model.Message) error
}

type PostHandleBase interface {
	Parse(msg *model.Message) error
	Check(msg *model.Message) error
	Handle(msg *model.Message) error
	PrintOpLogOk()
	PrintOpLogFail()
}
