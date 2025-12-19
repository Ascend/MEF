// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package innercommands

import (
	"errors"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

// ExchangeCaFlow is used to exchange root ca with OM
type ExchangeCaFlow struct {
	configPathMgr *pathmgr.ConfigPathMgr
	importPath    string
	exportPath    string
}

// NewExchangeCaFlow an ExchangeCaFlow struct
func NewExchangeCaFlow(importPath, exportPath string, configPathMgr *pathmgr.ConfigPathMgr) *ExchangeCaFlow {
	return &ExchangeCaFlow{configPathMgr: configPathMgr, importPath: importPath, exportPath: exportPath}
}

// RunTasks run mef config task
func (ecf ExchangeCaFlow) RunTasks() error {
	checkParam := checkParamTask{
		importPath: ecf.importPath,
		exportPath: ecf.exportPath,
	}
	if err := checkParam.runTask(); err != nil {
		hwlog.RunLog.Errorf("check exchange ca param failed, error: %v", err)
		return errors.New("check exchange ca param failed")
	}

	uid, gid, err := util.GetMefId()
	if err != nil {
		return err
	}

	savePath := ecf.configPathMgr.GetOMCertDir()
	importTask := common.InitImportCaTask(ecf.importPath, savePath, constants.RootCaName, uid, gid)
	if err = importTask.RunTask(); err != nil {
		hwlog.RunLog.Errorf("import ca failed: %s", err.Error())
		return errors.New("import ca failed")
	}

	installRootDir, err := path.GetInstallRootDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install root dir failed, error: %v", err)
		return errors.New("get install root dir failed")
	}
	generateCertsTask, err := common.NewGenerateCertsTask(installRootDir)
	if err != nil {
		hwlog.RunLog.Errorf("get generate certs task failed, error: %v", err)
		return errors.New("get generate certs task failed")
	}
	if err = generateCertsTask.MakeSureEdgeCerts(); err != nil {
		hwlog.RunLog.Errorf("make sure certs failed: %s", err.Error())
		return errors.New("make sure certs failed")
	}

	exportTask := exportCaTask{configPathMgr: ecf.configPathMgr, exportPath: ecf.exportPath}
	if err = exportTask.runTask(); err != nil {
		hwlog.RunLog.Error("export ca failed")
		return errors.New("export ca failed")
	}
	return nil
}

type checkParamTask struct {
	importPath string
	exportPath string
}

func (cpt *checkParamTask) runTask() error {
	if !(strings.HasPrefix(cpt.importPath, constants.DefaultInstallDir) &&
		strings.HasPrefix(cpt.exportPath, constants.DefaultInstallDir)) {
		hwlog.RunLog.Error("importPath or exportPath check failed")
		return errors.New("importPath or exportPath check failed")
	}
	if _, err := fileutils.RealFileCheck(cpt.importPath, true, false, constants.MaxCertSize); err != nil {
		hwlog.RunLog.Errorf("importPath file check failed: %s", err.Error())
		return errors.New("importPath file check failed")
	}

	exportDir := filepath.Dir(cpt.exportPath)
	if _, err := fileutils.RealDirCheck(exportDir, true, false); err != nil {
		hwlog.RunLog.Errorf("exportPath dir check failed: %s", err.Error())
		return errors.New("exportPath dir check failed")
	}

	if !fileutils.IsExist(cpt.exportPath) {
		return nil
	}

	if _, err := fileutils.RealFileCheck(cpt.exportPath, true, false, constants.MaxCertSize); err != nil {
		hwlog.RunLog.Errorf("exportPath file check failed: %s", err.Error())
		return errors.New("exportPath file check failed")
	}
	return nil
}

type exportCaTask struct {
	configPathMgr *pathmgr.ConfigPathMgr
	exportPath    string
}

func (ect *exportCaTask) runTask() error {
	var checkFunc = []func() error{
		ect.export,
	}
	for _, function := range checkFunc {
		if err := function(); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("export root_ca success")
	return nil
}

func (ect *exportCaTask) export() error {
	hwlog.RunLog.Info("start to export ca")

	srcPath := ect.configPathMgr.GetCompInnerRootCertPath(constants.EdgeMain)
	if err := fileutils.CopyFile(srcPath, ect.exportPath); err != nil {
		hwlog.RunLog.Errorf("export ca failed: %s", err.Error())
		return errors.New("export ca failed")
	}

	hwlog.RunLog.Info("export ca success")
	return nil
}
