// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestInstallPkg(t *testing.T) {
	convey.Convey("test install pkg", t, func() {
		convey.Convey("test install mgr file", DoInstallMgrTest)
		convey.Convey("test cert mgr file", CertMgrTest)
		convey.Convey("test working dir mgr file", WorkingDirMgrTest)
		convey.Convey("test yaml mgr file", YamlMgrTest)
	})
}
