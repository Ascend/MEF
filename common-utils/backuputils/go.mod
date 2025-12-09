module huawei.com/mindx/common/backuputils

go 1.17

require (
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/fileutils v0.0.1
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/utils v0.0.0
)

replace (
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/utils => ../utils
)
