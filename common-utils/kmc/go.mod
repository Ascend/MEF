module huawei.com/mindx/common/kmc

go 1.17

require (
	huawei.com/mindx/common/backuputils v0.0.0
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/utils v0.0.0
)

replace (
	huawei.com/mindx/common/backuputils => ../backuputils
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/utils => ../utils
)
