module edge-manager

go 1.20

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gin-gonic/gin v1.10.0
	github.com/smartystreets/goconvey v1.7.2
	gorm.io/driver/sqlite v1.5.3
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/backuputils v0.0.1
	huawei.com/mindx/common/cache v0.0.2
	huawei.com/mindx/common/checker v0.0.3
	huawei.com/mindx/common/database v0.0.2
	huawei.com/mindx/common/fileutils v0.0.14
	huawei.com/mindx/common/httpsmgr v0.0.2
	huawei.com/mindx/common/hwlog v0.10.12
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/test v0.0.1
	huawei.com/mindx/common/utils v0.1.13
	huawei.com/mindx/common/websocketmgr v0.0.2
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindx/common/xcrypto v0.0.2
	huawei.com/mindxedge/base v0.0.1
	k8s.io/api v0.28.1
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
)


replace (
	huawei.com/mindx/common/backuputils => ./../../common-utils/backuputils
	huawei.com/mindx/common/cache => ./../../common-utils/cache
	huawei.com/mindx/common/checker => ./../../common-utils/checker
	huawei.com/mindx/common/database => ./../../common-utils/database
	huawei.com/mindx/common/envutils => ./../../common-utils/envutils
	huawei.com/mindx/common/fileutils => ./../../common-utils/fileutils
	huawei.com/mindx/common/httpsmgr => ./../../common-utils/httpsmgr
	huawei.com/mindx/common/hwlog => ./../../common-utils/hwlog
	huawei.com/mindx/common/kmc => ./../../common-utils/kmc
	huawei.com/mindx/common/limiter => ./../../common-utils/limiter
	huawei.com/mindx/common/modulemgr => ./../../common-utils/modulemgr
	huawei.com/mindx/common/rand => ./../../common-utils/rand
	huawei.com/mindx/common/terminal => ./../../common-utils/terminal
	huawei.com/mindx/common/test => ./../../common-utils/test
	huawei.com/mindx/common/utils => ./../../common-utils/utils
	huawei.com/mindx/common/websocketmgr => ./../../common-utils/websocketmgr
	huawei.com/mindx/common/x509 => ./../../common-utils/x509
	huawei.com/mindx/common/xcrypto => ./../../common-utils/xcrypto
	huawei.com/mindx/mef/common/cmsverify => ./../../common-utils/cmsverify
	huawei.com/mindxedge/base v0.0.1 => ./../
)
