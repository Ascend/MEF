module huawei.com/mindx/common/logmgmt

go 1.17

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.7.2
	huawei.com/mindx/common/envutils v0.0.0
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/modulemgr v0.0.0
	huawei.com/mindx/common/utils v0.0.0
)

replace (
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/envutils => ../envutils
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/modulemgr => ../modulemgr
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/terminal => ../terminal
	huawei.com/mindx/common/utils => ../utils
)
