module huawei.com/mindx/common/checker

go 1.17

require (
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.1
	huawei.com/mindx/common/hwlog v0.10.3
)

replace (
	huawei.com/mindx/common/cache => ./../cache
	huawei.com/mindx/common/fileutils => ./../fileutils
	huawei.com/mindx/common/hwlog => ./../hwlog
	huawei.com/mindx/common/rand => ./../rand
	huawei.com/mindx/common/utils => ./../utils
)
