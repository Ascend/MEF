module huawei.com/mindx/common/k8stool

go 1.18

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.7.2
	huawei.com/mindx/common/fileutils v0.0.4
	huawei.com/mindx/common/hwlog v0.10.5
	huawei.com/mindx/common/kmc v0.0.0
	huawei.com/mindx/common/rand v0.0.0
	huawei.com/mindx/common/utils v0.1.5
	huawei.com/mindx/common/x509 v0.0.0
	k8s.io/client-go v0.28.1
)

replace (
	huawei.com/mindx/common/backuputils => ../backuputils
	huawei.com/mindx/common/cache => ../cache
	huawei.com/mindx/common/envutils => ../envutils
	huawei.com/mindx/common/fileutils => ../fileutils
	huawei.com/mindx/common/hwlog => ../hwlog
	huawei.com/mindx/common/kmc => ../kmc
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/terminal => ../terminal
	huawei.com/mindx/common/utils => ../utils
	huawei.com/mindx/common/x509 => ../x509
)
