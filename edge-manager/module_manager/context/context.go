package context

import "time"
import "edge-manager/module_manager/model"


type ModuleMessageContext interface {
	Registry(moduleName string) error
	Unregistry(moduleName string) error

	Send(msg *model.Message) error
	Receive(moduleName string) (*model.Message, error)
	SendSync(msg *model.Message, duration time.Duration) (*model.Message, error)
	SendResp(msg *model.Message) error
}
