module alarm-manager

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gin-gonic/gin v1.9.1
	github.com/smartystreets/goconvey v1.7.2
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/backuputils v0.0.1
	huawei.com/mindx/common/checker v0.0.2
	huawei.com/mindx/common/database v0.0.2
	huawei.com/mindx/common/httpsmgr v0.0.2
	huawei.com/mindx/common/hwlog v0.10.12
	huawei.com/mindx/common/kmc v0.1.0

	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/test v0.0.1
	huawei.com/mindx/common/websocketmgr v0.0.1
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindx/common/limiter v0.0.1
	huawei.com/mindxedge/base v0.0.1
)

replace (
	huawei.com/mindx/common/backuputils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/backuputils v0.0.9
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/checker => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/checker v0.0.16
	huawei.com/mindx/common/database => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/database v0.0.9
	huawei.com/mindx/common/envutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/envutils v0.1.8
	huawei.com/mindx/common/fileutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/fileutils v0.0.18
	huawei.com/mindx/common/httpsmgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/httpsmgr v0.0.17
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.10.13
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.1.11
	huawei.com/mindx/common/limiter =>  codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/limiter v0.0.13
	huawei.com/mindx/common/modulemgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/modulemgr v0.0.14
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.2
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.5
	huawei.com/mindx/common/test => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/test v0.0.2
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.37
	huawei.com/mindx/common/websocketmgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/websocketmgr v0.0.36
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.39
	huawei.com/mindx/mef/common/cmsverify => ./../MEF_Utils/cmsverify
	huawei.com/mindxedge/base v0.0.1 => ./../
)
