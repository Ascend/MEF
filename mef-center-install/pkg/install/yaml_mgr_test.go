// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/smartystreets/goconvey/convey"
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
        - name: edge-manager-config
          hostPath:
            path: ${config}
        - name: root-ca
          hostPath:
            path: ${root-ca}
`
}

func testGetYamlPath() {
	var yamlDealers = GetYamlDealers([]string{"edge-manager"},
		util.InitInstallDirPathMgr("./test_path"), "")
	yamlPath := yamlDealers[0].getYamlPath()
	yamlContent := getTestYaml()
	fmt.Printf(yamlPath)
	err := os.MkdirAll(filepath.Dir(yamlPath), common.Mode700)
	defer func() {
		err = os.RemoveAll("./test_path")
		So(err, ShouldBeNil)
	}()
	So(err, ShouldBeNil)
	writer, err := os.OpenFile(yamlPath, os.O_WRONLY|os.O_CREATE, common.Mode700)
	So(err, ShouldBeNil)
	defer func() {
		err = writer.Close()
		So(err, ShouldBeNil)
	}()
	_, err = writer.Write([]byte(yamlContent))
	So(err, ShouldBeNil)
	for _, dealers := range yamlDealers {
		err = dealers.EditSingleYaml([]string{"edge-manager"})
		So(err, ShouldBeNil)
	}

}

func YamlEditTest() {
	Convey("test yaml edit function", testGetYamlPath)
}
