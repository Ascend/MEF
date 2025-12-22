module huawei.com/mindx/common/httpsmgr

go 1.17

require (
	github.com/gin-gonic/gin v1.9.1
	huawei.com/mindx/common/checker v0.0.0
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/limiter v0.0.0
	huawei.com/mindx/common/utils v0.0.0
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
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/terminal => ../terminal
	huawei.com/mindx/common/utils => ../utils
	huawei.com/mindx/common/x509 => ../x509
)
