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
	"os"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

func YamlMgrTest() {
	YamlEditTest()
}

func getTestYaml() string {
	return `
spec:
  template:
    spec:
      containers:
          env:
            - name: installed-module
              value: ${installed_module}
      volumes:
        - name: edge-manager-log
          hostPath:
            path: ${log}
            type: Directory
        - name: edge-manager-log-backup
          hostPath:
            path: ${log-backup}
            type: Directory
        - name: edge-manager-config
          hostPath:
            path: ${config}
        - name: root-ca
          hostPath:
            path: ${root-ca}
        - name: public-config
          hostPath:
            path: ${public-config}
`
}

func testGetYamlPath() {
	yamlPath := "./test.yaml"
	pathMgr, err := util.InitInstallDirPathMgr(yamlPath)
	convey.So(err, convey.ShouldBeNil)
	var yamlDealer = GetYamlDealer(pathMgr, "edge-manager", "", "", yamlPath)
	yamlContent := getTestYaml()
	writer, err := os.OpenFile(yamlPath, os.O_WRONLY|os.O_CREATE, common.Mode700)
	convey.So(err, convey.ShouldBeNil)
	defer func() {
		err = writer.Close()
		convey.So(err, convey.ShouldBeNil)
	}()
	defer func() {
		err = fileutils.DeleteFile("./test.yaml")
		convey.So(err, convey.ShouldBeNil)
	}()
	_, err = writer.Write([]byte(yamlContent))
	convey.So(err, convey.ShouldBeNil)

	err = yamlDealer.EditSingleYaml([]string{"edge-manager"})
	convey.So(err, convey.ShouldBeNil)
}

func YamlEditTest() {
	convey.Convey("test yaml edit function", testGetYamlPath)
}
