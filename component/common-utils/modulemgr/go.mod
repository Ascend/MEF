module huawei.com/mindx/common/modulemgr

go 1.17

require (
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/limiter v0.0.0
)

replace (
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/limiter => ../limiter
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/utils => ../utils
)
