module huawei.com/mindx/common/database

go 1.17

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.7.2
	gorm.io/driver/sqlite v1.5.3
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/backuputils v0.0.1
	huawei.com/mindx/common/fileutils v0.0.1
	huawei.com/mindx/common/hwlog v0.10.5
)

replace (
	huawei.com/mindx/common/backuputils => ./../backuputils
	huawei.com/mindx/common/cache => ./../cache
	huawei.com/mindx/common/fileutils => ./../fileutils
	huawei.com/mindx/common/hwlog => ./../hwlog
	huawei.com/mindx/common/rand => ./../rand
	huawei.com/mindx/common/utils => ./../utils
)