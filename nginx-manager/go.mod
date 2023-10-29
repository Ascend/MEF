module nginx-manager

go 1.16

require (
	github.com/gin-gonic/gin v1.9.1
	huawei.com/mindx/common/backuputils v0.0.1
	huawei.com/mindx/common/checker v0.0.3
	huawei.com/mindx/common/envutils v0.0.4
	huawei.com/mindx/common/fileutils v0.0.1
	huawei.com/mindx/common/httpsmgr v0.0.2
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/logmgmt v0.0.1
	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindxedge/base v0.0.1
)

replace (
	huawei.com/mindx/common/backuputils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/backuputils v0.0.8
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/checker => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/checker v0.0.10
	huawei.com/mindx/common/database => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/database v0.0.8
	huawei.com/mindx/common/envutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/envutils v0.1.5
	huawei.com/mindx/common/fileutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/fileutils v0.0.14
	huawei.com/mindx/common/httpsmgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/httpsmgr v0.0.15
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.10.11
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.1.11
	huawei.com/mindx/common/limiter => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/limiter v0.0.10
	huawei.com/mindx/common/logmgmt => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/logmgmt v0.0.7
	huawei.com/mindx/common/modulemgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/modulemgr v0.0.4
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.2
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.5
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.23
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.31
	huawei.com/mindx/mef/common/cmsverify => ./../MEF_Utils/cmsverify
	huawei.com/mindxedge/base v0.0.1 => ./../
)
