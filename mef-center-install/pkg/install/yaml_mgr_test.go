// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"os"

	. "github.com/smartystreets/goconvey/convey"
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
	So(err, ShouldBeNil)
	var yamlDealer = GetYamlDealer(pathMgr, "edge-manager", "", "", yamlPath)
	yamlContent := getTestYaml()
	writer, err := os.OpenFile(yamlPath, os.O_WRONLY|os.O_CREATE, common.Mode700)
	So(err, ShouldBeNil)
	defer func() {
		err = writer.Close()
		So(err, ShouldBeNil)
	}()
	defer func() {
		err = fileutils.DeleteAllFileWithConfusion("./test.yaml")
		So(err, ShouldBeNil)
	}()
	_, err = writer.Write([]byte(yamlContent))
	So(err, ShouldBeNil)

	err = yamlDealer.EditSingleYaml([]string{"edge-manager"})
	So(err, ShouldBeNil)
}

func YamlEditTest() {
	Convey("test yaml edit function", testGetYamlPath)
}
