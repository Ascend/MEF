// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
