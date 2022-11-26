module software-manager

require (
	huawei.com/mindx/common/hwlog v0.0.0
	huawei.com/mindxedge/base v0.0.1
)

replace (
	huawei.com/mindxedge/base v0.0.1 => ./../

	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.0.10
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.3
)