package websocket

type CertPathInfo struct {
	RootCaPath  string
	SvrCertPath string
	SvrKeyPath  string
	ServerFlag  bool
}

type RegisterModuleInfo struct {
	MsgOpt     string
	MsgRes     string
	ModuleName string
}
