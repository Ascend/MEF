module huawei.com/mindx/common/websocketmgr

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gorilla/websocket v1.5.1
	github.com/smartystreets/goconvey v1.7.2
	huawei.com/mindx/common/checker v0.0.0
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/limiter v0.0.0
	huawei.com/mindx/common/modulemgr v0.0.0
	huawei.com/mindx/common/test v0.0.0
	huawei.com/mindx/common/utils v0.1.5
	huawei.com/mindx/common/x509 v0.0.0
)

replace (
	huawei.com/mindx/common/backuputils => ../backuputils
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/checker => ../checker
	huawei.com/mindx/common/envutils => ../envutils
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/kmc => ../kmc
	huawei.com/mindx/common/limiter => ../limiter
	huawei.com/mindx/common/modulemgr => ../modulemgr
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/terminal => ../terminal
	huawei.com/mindx/common/test => ../test
	huawei.com/mindx/common/utils => ../utils
	huawei.com/mindx/common/x509 => ../x509
)
