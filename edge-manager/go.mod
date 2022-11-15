module edge-manager

go 1.16

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/gorilla/websocket v1.5.0
	github.com/stretchr/testify v1.7.1
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.0
	huawei.com/mindx/common/hwlog v0.0.0
	huawei.com/mindx/common/k8stool v0.0.0
	huawei.com/mindx/common/utils v0.0.0
	k8s.io/client-go v0.19.11
	k8s.io/api v0.0.0
    k8s.io/apimachinery v0.0.0
)

replace (
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.2
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.0.10
	huawei.com/mindx/common/k8stool => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/k8stool v0.0.3
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.0.3
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.1.3
	k8s.io/api => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/apimachinery => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/client-go => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
)
