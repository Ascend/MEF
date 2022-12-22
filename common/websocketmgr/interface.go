package websocketmgr

type NetProxyIntf interface {
	Start() error
	Send(msg interface{}) error
	Stop() error
	GetName() string
}

type HandleMsgIntf interface {
	handleMsg(msg []byte)
}
