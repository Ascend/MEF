// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package modelchecker for check model file param
package modelchecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

const (
	modelFileUpdateLen    = 1
	modelFileDeleteMinLen = 0
	modelFileDeleteMaxLen = 256
)

var modelFileSuffixList = []string{".om", ".tar.gz", ".zip"}
var modelFileMsgCheckMap map[string]func(updateInfo types.ModelFileInfo) checker.CheckResult

func init() {
	modelFileMsgCheckMap = make(map[string]func(updateInfo types.ModelFileInfo) checker.CheckResult)
	modelFileMsgCheckMap[constants.OptUpdate] = checkUpdate
	modelFileMsgCheckMap[constants.OptDelete] = checkDelete
}

type modelUpdateMsgChecker struct {
	modelChecker checker.ModelChecker
}

type modelDeleteTempMsgChecker struct {
	modelChecker checker.ModelChecker
}

type modelDeleteAllMsgChecker struct {
	modelChecker checker.ModelChecker
}

// CheckModelFileMsg check and modify model file msg
func CheckModelFileMsg(content []byte) error {
	var err error
	var updateInfo types.ModelFileInfo
	if err = json.Unmarshal(content, &updateInfo); err != nil {
		hwlog.RunLog.Error("unmarshal update model file message failed")
		return errors.New("unmarshal update model file message failed")
	}
	defer func() {
		iterationCount := 1
		for i := range updateInfo.ModelFiles {
			if iterationCount > modelFileUpdateLen {
				break
			}
			iterationCount++
			utils.ClearStringMemory(updateInfo.ModelFiles[i].FileServer.PassWord)
		}
	}()
	if err = checkModelFileMsg(updateInfo); err != nil {
		return err
	}

	return nil
}

func checkModelFileMsg(updateInfo types.ModelFileInfo) error {
	checkFunc, ok := modelFileMsgCheckMap[updateInfo.Operation]
	if !ok {
		return errors.New("no check func for model file")
	}
	if checkResult := checkFunc(updateInfo); !checkResult.Result {
		hwlog.RunLog.Errorf("model file check failed, error: %s", checkResult.Reason)
		return errors.New(checkResult.Reason)
	}
	return nil
}

func checkUpdate(updateInfo types.ModelFileInfo) checker.CheckResult {
	return check(updateInfo, (&modelUpdateMsgChecker{}).init().modelChecker)
}

func checkDelete(updateInfo types.ModelFileInfo) checker.CheckResult {
	switch updateInfo.Target {
	case constants.TargetTypeTemp, "":
		return check(updateInfo, (&modelDeleteTempMsgChecker{}).init().modelChecker)
	case constants.TargetTypeAll:
		return check(updateInfo, (&modelDeleteAllMsgChecker{}).init().modelChecker)
	default:
		return checker.NewFailedResult(
			fmt.Sprintf("model check target failed"))
	}
}

func (m *modelUpdateMsgChecker) init() *modelUpdateMsgChecker {
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Target", []string{"all"}, true),
		checker.GetRegChecker("Uuid", "^"+constants.UUIDRegex+"$", true),
		checker.GetListChecker("ModelFiles", &modelFileUpdateChecker{},
			modelFileUpdateLen, modelFileUpdateLen, true),
	)
	return m
}

func (m *modelDeleteTempMsgChecker) init() *modelDeleteTempMsgChecker {
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Uuid", "^"+constants.UUIDRegex+"$", true),
		checker.GetListChecker("ModelFiles", &modelFileDeleteChecker{},
			modelFileDeleteMinLen, modelFileDeleteMaxLen, true),
	)
	return m
}

func (m *modelDeleteAllMsgChecker) init() *modelDeleteAllMsgChecker {
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Uuid", "^"+constants.UUIDRegex+"$", true),
		checker.GetListChecker("ModelFiles", &modelFileDeleteChecker{},
			modelFileDeleteMinLen, modelFileDeleteMaxLen, true),
	)
	return m
}

func check(data interface{}, mChecker checker.ModelChecker) checker.CheckResult {
	checkResult := mChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("model info checker check failed, error: %s", checkResult.Reason))
	}
	hwlog.RunLog.Info("model info checker check success")
	return checker.NewSuccessResult()
}

type modelFileUpdateChecker struct {
	modelChecker checker.ModelChecker
}

type modelFileDeleteChecker struct {
	modelChecker checker.ModelChecker
}

type modelStringChecker struct {
	checkFunc func(data interface{}) checker.CheckResult
}

func (m *modelFileUpdateChecker) init() {
	fileChecker := &checker.ModelChecker{Field: "FileServer", Required: true,
		Checker: checker.GetAndChecker(&fileServerChecker{})}
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetAndChecker(
			checker.GetRegChecker("Name", constants.ModelFileNameReg, true),
			checker.GetStringExcludeChecker("Name", []string{".."}, true),
			&modelStringChecker{checkFunc: checkModelName},
		),
		&modelStringChecker{checkFunc: checkModelSize},
		checker.GetAndChecker(
			checker.GetRegChecker("Version", constants.VersionLenReg, true),
		),
		checker.GetStringChoiceChecker("CheckType", []string{"sha256"}, true),
		checker.GetRegChecker("CheckCode", constants.CheckCodeReg, true),
		fileChecker,
	)
}

func (m *modelFileDeleteChecker) init() {
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetAndChecker(
			checker.GetRegChecker("Name", constants.ModelFileNameReg, true),
			checker.GetStringExcludeChecker("Name", []string{".."}, true),
			&modelStringChecker{checkFunc: checkModelName},
		),
		checker.GetAndChecker(
			checker.GetRegChecker("Version", constants.VersionLenReg, true),
		),
	)
}

func (m *modelFileUpdateChecker) Check(data interface{}) checker.CheckResult {
	m.init()
	checkResult := m.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("model file update checker check failed, error: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (m *modelFileDeleteChecker) Check(data interface{}) checker.CheckResult {
	m.init()
	checkResult := m.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("model file delete checker check failed, error: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func (s *modelStringChecker) Check(data interface{}) checker.CheckResult {
	return s.checkFunc(data)
}

func checkModelName(data interface{}) checker.CheckResult {
	stringValuer := valuer.StringValuer{}
	name, err := stringValuer.GetValue(data, "Name")
	if err != nil {
		return checker.NewFailedResult(
			fmt.Sprintf("model name check failed:%v", err))
	}
	for _, v := range modelFileSuffixList {
		if strings.HasSuffix(name, v) {
			return checker.NewSuccessResult()
		}
	}
	return checker.NewFailedResult(fmt.Sprintf("name suffix not valid"))
}

func checkModelSize(data interface{}) checker.CheckResult {
	stringValuer := valuer.StringValuer{}
	size, err := stringValuer.GetValue(data, "Size")
	if err != nil {
		return checker.NewFailedResult(
			fmt.Sprintf("model size check failed:%v", err))
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return checker.NewFailedResult(
			fmt.Sprintf("model size check failed:%v", err))
	}
	if sizeInt <= 0 || sizeInt > constants.ModelFileMaxSize {
		return checker.NewFailedResult(
			fmt.Sprintf("model file size is invalid"))
	}
	return checker.NewSuccessResult()
}

type fileServerChecker struct {
	modelChecker checker.ModelChecker
}

func (f *fileServerChecker) init() {
	f.modelChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Protocol", []string{"https"}, true),
		checker.GetRegChecker("UserName", constants.ModelFileUserNameReg, true),
		checker.GetRegChecker("PassWord", constants.ModeFilePwdReg, true),
		checker.GetHttpsUrlChecker("Path", true, true),
	)
}

func (f *fileServerChecker) Check(data interface{}) checker.CheckResult {
	f.init()
	checkResult := f.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("file server checker check failed, error: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
