module huawei.com/mindx/common/envutils

go 1.17

require (
	huawei.com/mindx/common/hwlog v0.10.3
	huawei.com/mindx/common/terminal v0.0.5
	huawei.com/mindx/common/fileutils v0.0.4
)

replace (
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/terminal => ../terminal
	huawei.com/mindx/common/fileutils => ../fileutils
)
