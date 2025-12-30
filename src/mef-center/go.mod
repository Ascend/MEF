module huawei.com/mindxedge/base

go 1.20

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gin-gonic/gin v1.10.0
	github.com/smartystreets/goconvey v1.7.2
	gorm.io/gorm v1.25.4
	huawei.com/mindx/common/backuputils v0.0.1
	huawei.com/mindx/common/checker v0.0.2
	huawei.com/mindx/common/database v0.0.2
	huawei.com/mindx/common/envutils v0.0.4
	huawei.com/mindx/common/fileutils v0.0.14
	huawei.com/mindx/common/httpsmgr v0.0.2
	huawei.com/mindx/common/hwlog v0.10.12
	huawei.com/mindx/common/kmc v0.1.0
	huawei.com/mindx/common/limiter v0.0.0
	huawei.com/mindx/common/modulemgr v0.0.1
	huawei.com/mindx/common/rand v0.0.1
	huawei.com/mindx/common/test v0.0.1
	huawei.com/mindx/common/utils v0.1.13
	huawei.com/mindx/common/x509 v0.0.12
	huawei.com/mindx/mef/common/cmsverify v0.0.1
)

require (
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/sqlite v1.5.3 // indirect
	huawei.com/mindx/common/cache v0.0.0 // indirect
	huawei.com/mindx/common/terminal v0.0.5 // indirect
)

replace (
	huawei.com/mindx/common/backuputils => ./../common-utils/backuputils
	huawei.com/mindx/common/cache => ./../common-utils/cache
	huawei.com/mindx/common/checker => ./../common-utils/checker
	huawei.com/mindx/common/database => ./../common-utils/database
	huawei.com/mindx/common/envutils => ./../common-utils/envutils
	huawei.com/mindx/common/fileutils => ./../common-utils/fileutils
	huawei.com/mindx/common/httpsmgr => ./../common-utils/httpsmgr
	huawei.com/mindx/common/hwlog => ./../common-utils/hwlog
	huawei.com/mindx/common/kmc => ./../common-utils/kmc
	huawei.com/mindx/common/limiter => ./../common-utils/limiter
	huawei.com/mindx/common/modulemgr => ./../common-utils/modulemgr
	huawei.com/mindx/common/rand => ./../common-utils/rand
	huawei.com/mindx/common/terminal => ./../common-utils/terminal
	huawei.com/mindx/common/test => ./../common-utils/test
	huawei.com/mindx/common/utils => ./../common-utils/utils
	huawei.com/mindx/common/x509 => ./../common-utils/x509
	huawei.com/mindx/mef/common/cmsverify => ./../common-utils/cmsverify
)
