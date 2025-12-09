module huawei.com/mindx/common/cmsverify

go 1.17

require huawei.com/mindx/common/fileutils v0.0.1


replace (
	huawei.com/mindx/common/rand => ../rand
	huawei.com/mindx/common/fileutils => ../fileutils
)
