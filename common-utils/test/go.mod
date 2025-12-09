module huawei.com/mindx/common/test

go 1.17

require (
	gorm.io/driver/sqlite v1.4.2
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/hwlog v0.10.5
)

replace (
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/utils => ../utils
)
