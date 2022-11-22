module huawei.com/mindxedge/base

go 1.16

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/stretchr/testify v1.7.1
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.0
	huawei.com/mindx/common/hwlog v0.0.0
)

replace (
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.0.10
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.3
)
