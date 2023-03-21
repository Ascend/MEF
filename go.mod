module huawei.com/mindxedge/base

go 1.16

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/gorilla/websocket v1.5.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	gorm.io/driver/sqlite v1.4.2
	gorm.io/gorm v1.22.3
	huawei.com/mindx/common/hwlog v0.10.2
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/rand v0.0.1
	huawei.com/mindx/common/utils v0.1.5
	huawei.com/mindx/common/x509 v0.0.8
)

replace (
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.10.3-0.20230213134501-668f7ebdd348
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.1.0
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.1
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.5
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.5
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.8
	huawei.com/mindx/mef/common/cmsverify => ./MEF_Utils/cmsverify
)
