module Ascend-device-plugin

go 1.18

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/smartystreets/goconvey v1.7.2
	google.golang.org/grpc v1.57.2
	huawei.com/mindx/common/hwlog v0.0.0
	huawei.com/mindx/common/limiter v0.0.0
	k8s.io/apimachinery v0.28.1
	k8s.io/kubelet v0.28.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	golang.org/x/net v0.13.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	huawei.com/mindx/common/cache v0.0.0 // indirect
	huawei.com/mindx/common/fileutils v0.0.0 // indirect
	huawei.com/mindx/common/rand v0.0.0 // indirect
	huawei.com/mindx/common/utils v0.0.0 // indirect
)

replace (
	huawei.com/mindx/common/cache => ./../common-utils/cache
	huawei.com/mindx/common/fileutils => ./../common-utils/fileutils
	huawei.com/mindx/common/hwlog => ./../common-utils/hwlog
	huawei.com/mindx/common/limiter => ./../common-utils/limiter
	huawei.com/mindx/common/rand => ./../common-utils/rand
	huawei.com/mindx/common/utils => ./../common-utils/utils
)
