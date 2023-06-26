module huawei.com/mindxedge/base

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gin-gonic/gin v1.9.1
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.1
	gorm.io/driver/sqlite v1.4.2
	gorm.io/gorm v1.22.3
	huawei.com/mindx/common/checker v0.0.2
	huawei.com/mindx/common/envutils v0.0.4
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/rand v0.0.1
	huawei.com/mindx/common/utils v0.1.5
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindx/mef/common/cmsverify v0.0.0-00010101000000-000000000000
)

replace (
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/checker => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/checker v0.0.2
	huawei.com/mindx/common/envutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/envutils v0.0.8
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.10.6
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.1.6
	huawei.com/mindx/common/modulemgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/modulemgr v0.0.2
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.1
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.5
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.19
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.18
	huawei.com/mindx/mef/common/cmsverify => ./MEF_Utils/cmsverify
)
