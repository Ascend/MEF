// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package install

import (
	"fmt"
	"path"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// YamlMgr is a struct to manage a single component's yaml file
// used on installation procedure to edit the mount path on those yaml file
type YamlMgr struct {
	component    string
	pathMgr      *util.InstallDirPathMgr
	yamlPath     string
	logDir       string
	logBackupDir string
}

type modifier struct {
	component      string
	content        string
	mark           string
	modifiedString string
}

// GetYamlDealer returns a YamlMgr for a components
func GetYamlDealer(pathMgr *util.InstallDirPathMgr, component, logDir, logBackupDir, yamlPath string) *YamlMgr {
	return &YamlMgr{
		component:    component,
		pathMgr:      pathMgr,
		logDir:       logDir,
		logBackupDir: logBackupDir,
		yamlPath:     yamlPath,
	}
}

func (yd *YamlMgr) modifyLogDir(content string) (string, error) {
	modifierIns := modifier{
		component:      yd.component,
		content:        content,
		mark:           util.LogFlag,
		modifiedString: path.Join(yd.logDir, util.ModuleLogName, yd.component),
	}

	return modifierIns.modifyMntDir()
}

func (yd *YamlMgr) modifyLogBackupDir(content string) (string, error) {
	modifierIns := modifier{
		component:      yd.component,
		content:        content,
		mark:           util.LogBackupFlag,
		modifiedString: path.Join(yd.logBackupDir, util.ModuleLogBackupName, yd.component),
	}

	return modifierIns.modifyMntDir()
}

func (yd *YamlMgr) modifyRootCaDir(content string) (string, error) {
	modifierIns := modifier{
		component:      yd.component,
		content:        content,
		mark:           util.RootCaFlag,
		modifiedString: yd.pathMgr.ConfigPathMgr.GetRootCaCertDirPath(),
	}
	return modifierIns.modifyMntDir()
}

func (yd *YamlMgr) modifyModuleDir(content string) (string, error) {
	modifierIns := modifier{
		component:      yd.component,
		content:        content,
		mark:           util.ConfigFlag,
		modifiedString: yd.pathMgr.ConfigPathMgr.GetComponentConfigPath(yd.component),
	}
	return modifierIns.modifyMntDir()
}

func (yd *YamlMgr) modifyInstalledModule(content string, installedModule []string) (string, error) {
	modifierIns := modifier{
		component:      yd.component,
		content:        content,
		mark:           util.InstalledModuleFlag,
		modifiedString: yd.getModuleString(installedModule),
	}
	return modifierIns.modifyEnv()
}

func (yd *YamlMgr) getModuleString(installedModule []string) string {
	result := `"{[`
	for idx, module := range installedModule {
		result = fmt.Sprintf(`%s'%s'`, result, module)
		if idx < len(installedModule)-1 {
			result = result + ", "
		}
	}
	result = result + `]}"`
	return result
}

func (yd *YamlMgr) modifyContent(content string, installedModule []string) (string, error) {
	var err error
	content, err = yd.modifyModuleDir(content)
	if err != nil {
		return "", err
	}

	content, err = yd.modifyLogDir(content)
	if err != nil {
		return "", err
	}

	content, err = yd.modifyLogBackupDir(content)
	if err != nil {
		return "", err
	}

	content, err = yd.modifyRootCaDir(content)
	if err != nil {
		return "", err
	}

	content, err = yd.modifyInstalledModule(content, installedModule)
	if err != nil {
		return "", err
	}

	return content, nil
}

// EditSingleYaml is used to edit a single yaml file on installation
func (yd *YamlMgr) EditSingleYaml(installedModule []string) error {
	hwlog.RunLog.Infof("start to modify %s's yaml", yd.component)
	ret, err := utils.LoadFile(yd.yamlPath)
	if err != nil {
		hwlog.RunLog.Errorf("reading yaml [%s] meets error: %v", yd.yamlPath, err)
		return err
	}

	content := string(ret)
	content, err = yd.modifyContent(content, installedModule)
	if err != nil {
		return err
	}

	err = common.WriteData(yd.yamlPath, []byte(content))
	if err != nil {
		hwlog.RunLog.Errorf("write yaml [%s] meets error: %v", yd.yamlPath, err)
		return err
	}
	return nil
}

func (md *modifier) modifyMntDir() (string, error) {
	var retString string
	subStrings := strings.SplitN(md.content, md.mark, util.SplitCount)
	if len(subStrings) < util.SplitCount {
		hwlog.RunLog.Errorf("split [%s]'s yaml by [%s] failed, not enough substrings", md.component, md.mark)
		return "", fmt.Errorf("modify [%s]'s yaml failed", md.component)
	}
	retString = subStrings[0] + md.modifiedString

	subStrings = strings.SplitN(subStrings[1], util.LineSplitter, util.SplitCount)
	if len(subStrings) < util.SplitCount {
		hwlog.RunLog.Errorf("split [%s]'s yaml by [%s] failed, not enough substrings",
			md.component, util.LineSplitter)
		return "", fmt.Errorf("modify [%s]'s yaml failed", md.component)
	}
	retString = retString + util.LineSplitter + subStrings[1]

	return retString, nil
}

func (md *modifier) modifyEnv() (string, error) {
	var retString string
	subStrings := strings.SplitN(md.content, md.mark, util.SplitCount)
	if len(subStrings) < util.SplitCount {
		hwlog.RunLog.Errorf("split [%s]'s yaml by [%s] failed, not enough substrings", md.component, md.mark)
		return "", fmt.Errorf("modify [%s]'s yaml failed", md.component)
	}
	retString = subStrings[0] + md.modifiedString

	subStrings = strings.SplitN(subStrings[1], util.LineSplitter, util.SplitCount)
	if len(subStrings) < util.SplitCount {
		hwlog.RunLog.Errorf("split [%s]'s yaml by [%s] failed, not enough substrings",
			md.component, util.LineSplitter)
		return "", fmt.Errorf("modify [%s]'s yaml failed", md.component)
	}
	retString = retString + util.LineSplitter + subStrings[1]

	return retString, nil
}
