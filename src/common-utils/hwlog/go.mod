module huawei.com/mindx/common/hwlog

go 1.17

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/smartystreets/goconvey v1.7.2
	huawei.com/mindx/common/cache v0.0.0
	huawei.com/mindx/common/fileutils v0.0.0
)

replace (
	huawei.com/mindx/common/cache => ./../cache
	huawei.com/mindx/common/fileutils => ./../fileutils
	huawei.com/mindx/common/rand => ./../rand
	huawei.com/mindx/common/utils => ./../utils
)
