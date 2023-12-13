// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInstallPkg(t *testing.T) {
	Convey("test install pkg", t, func() {
		Convey("test install mgr file", DoInstallMgrTest)
		Convey("test cert mgr file", CertMgrTest)
		Convey("test working dir mgr file", WorkingDirMgrTest)
		Convey("test yaml mgr file", YamlMgrTest)
	})
}
