module edge-manager

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gin-gonic/gin v1.9.1
	github.com/smartystreets/goconvey v1.7.2
	gorm.io/driver/sqlite v1.4.2
	gorm.io/gorm v1.24.1
	huawei.com/mindx/common/checker v0.0.3
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/k8stool v0.0.4
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/utils v0.1.13
	huawei.com/mindx/common/websocketmgr v0.0.2
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindx/common/xcrypto v0.0.2
	huawei.com/mindxedge/base v0.0.1
	k8s.io/api v0.25.3
	k8s.io/apimachinery v0.25.3
	k8s.io/client-go v0.25.3
)

replace (
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/checker => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/checker v0.0.3
	huawei.com/mindx/common/envutils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/envutils v0.0.8
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.10.6
	huawei.com/mindx/common/k8stool => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/k8stool v0.0.4
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.1.4
	huawei.com/mindx/common/modulemgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/modulemgr v0.0.2
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.1
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.5
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.19
	huawei.com/mindx/common/websocketmgr => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/websocketmgr v0.0.4
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.18
	huawei.com/mindx/common/xcrypto => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/xcrypto v0.0.2
	huawei.com/mindx/mef/common/cmsverify => ./../MEF_Utils/cmsverify
	huawei.com/mindxedge/base v0.0.1 => ./../
	k8s.io/api => szv-open.codehub.huawei.com/OpenSourceCenter/kubernetes/kubernetes.git/staging/src/k8s.io/api v0.0.0-20230316115657-3ccff71419e0
	k8s.io/apimachinery => szv-open.codehub.huawei.com/OpenSourceCenter/kubernetes/kubernetes.git/staging/src/k8s.io/apimachinery v0.0.0-20230316115657-3ccff71419e0
	k8s.io/client-go => szv-open.codehub.huawei.com/OpenSourceCenter/kubernetes/kubernetes.git/staging/src/k8s.io/client-go v0.0.0-20230316115657-3ccff71419e0

)
