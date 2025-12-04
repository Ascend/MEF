// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/database"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

type updateModelFileHandler struct {
}

var lockResError = errors.New("cannot perform this operation, other operation is working")

func (u *updateModelFileHandler) checkModeFileNum(modelFileInfo types.ModelFileInfo) error {
	fileList := modeltask.GetModelMgr().GetFileList()
	var uuidInfo = make(map[string]*utils.Set)

	for _, file := range fileList {
		if _, ok := uuidInfo[file.Uuid]; !ok {
			uuidInfo[file.Uuid] = utils.NewSet()
		}
		uuidInfo[file.Uuid].Add(file.Name)
	}

	for _, modelFile := range modelFileInfo.ModelFiles {
		if _, ok := uuidInfo[modelFileInfo.Uuid]; !ok {
			uuidInfo[modelFileInfo.Uuid] = utils.NewSet()
		}
		uuidInfo[modelFileInfo.Uuid].Add(modelFile.Name)
	}

	perPodModelFileNumber := configpara.GetPodConfig().ContainerModelFileNumber
	totalPodModelFileNumber := configpara.GetPodConfig().TotalModelFileNumber
	if len(uuidInfo[modelFileInfo.Uuid].List()) > perPodModelFileNumber {
		hwlog.RunLog.Error("model file number of per pod up to limit")
		return fmt.Errorf("model file number of per pod up to limit")
	}

	var totalNumber int
	for uuid := range uuidInfo {
		totalNumber += len(uuidInfo[uuid].List())
	}
	if totalNumber > totalPodModelFileNumber {
		hwlog.RunLog.Error("total model file number up to limit")
		return fmt.Errorf("total model file number up to limit")
	}

	return nil
}

func (u *updateModelFileHandler) Handle(msg *model.Message) error {
	modeltask.SendOkResponse(msg)

	var info types.ModelFileInfo
	var err error
	if err = msg.ParseContent(&info); err != nil {
		modeltask.SendFailResponse(constants.ActionModelFiles, "model file update param error")
		return fmt.Errorf("updateModelfileHandler failed, model file update param error: %v", err)
	}
	if info.Operation == constants.OptUpdate {
		err = u.update(info)
		if err != nil {
			for i := range info.ModelFiles {
				modeltask.GetModelMgr().AddFailTask(info.Uuid, info.ModelFiles[i], "operate model file failed")
			}
		}
	} else if info.Operation == constants.OptDelete {
		err = u.delete(msg, info)
	}

	if err != nil {
		hwlog.RunLog.Errorf("operate model file failed: %v", err)
		modeltask.SendFailResponse(constants.ActionModelFiles, "operate model file failed")
		return errors.New("operate model file failed")
	}
	modeltask.SendConfigResult(constants.ActionModelFiles)
	return nil
}

func (u *updateModelFileHandler) deleteNotActive(msg *model.Message, info types.ModelFileInfo) error {
	if !modeltask.GetModelMgr().LockUuid(info.Uuid) {
		return lockResError
	}
	defer modeltask.GetModelMgr().UnLockUuid(info.Uuid)
	modeltask.GetModelMgr().DelNotActiveTasks(info.Uuid, info.ModelFiles)
	err := u.sendAsyncMsgToEdgeOm(constants.OptUpdate, constants.ActionModelFiles, msg.Content, true)
	if err != nil {
		return err
	}
	return nil
}

func (u *updateModelFileHandler) deleteActiveAndNotActive(msg *model.Message, info types.ModelFileInfo) error {
	if !modeltask.GetModelMgr().LockUuid(info.Uuid) {
		return lockResError
	}
	defer modeltask.GetModelMgr().UnLockUuid(info.Uuid)
	modeltask.GetModelMgr().DelActiveAndNotActiveTasks(info.Uuid, info.ModelFiles)
	err := u.sendAsyncMsgToEdgeOm(constants.OptUpdate, constants.ActionModelFiles, msg.Content, true)
	if err != nil {
		return err
	}
	return nil
}

func (u *updateModelFileHandler) deleteByUuid(msg *model.Message, info types.ModelFileInfo) error {
	if !modeltask.GetModelMgr().LockUuid(info.Uuid) {
		return lockResError
	}
	defer modeltask.GetModelMgr().UnLockUuid(info.Uuid)
	modeltask.GetModelMgr().DelTasksByUuid(info.Uuid)

	err := u.sendAsyncMsgToEdgeOm(constants.OptUpdate, constants.ActionModelFiles, msg.Content, true)
	if err != nil {
		return err
	}
	return nil
}

func (u *updateModelFileHandler) delete(msg *model.Message, info types.ModelFileInfo) error {
	if info.Target == constants.TargetTypeTemp {
		return u.deleteNotActive(msg, info)
	} else if info.Target == constants.TargetTypeAll {
		return u.deleteByUuid(msg, info)
	} else if info.Target == "" {
		return u.deleteActiveAndNotActive(msg, info)
	}
	return nil
}

func (u *updateModelFileHandler) update(info types.ModelFileInfo) error {
	err := u.syncFile(info.Uuid)
	if err != nil {
		return err
	}

	if err = u.checkModeFileNum(info); err != nil {
		return err
	}

	caData, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:      constants.ModHandlerMgr,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptGet,
		Resource:    constants.ResConfig,
		Content:     constants.SoftwareCert,
	})
	if err != nil {
		return err
	}
	if caData == constants.Failed {
		return errors.New("get cert failed when update model file")
	}

	content := types.OperateModelFileContent{
		Operate:     "check",
		OperateInfo: nil,
	}
	checkResult, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:                constants.ModHandlerMgr,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptRaw,
		Resource:              constants.ActionModelFiles,
		Content:               content,
		TransferStructIntoStr: true,
	})
	if err != nil {
		return err
	}
	if checkResult != constants.Success {
		return errors.New("check docker path failed")
	}

	for _, m := range info.ModelFiles {
		if !modeltask.GetModelMgr().Lock(info.Uuid, m.Name) {
			return lockResError
		}
		if err = modeltask.GetModelMgr().AddTask(info.Uuid, m, []byte(caData)); err != nil {
			modeltask.GetModelMgr().UnLock(info.Uuid, m.Name)
			return err
		}
		modeltask.GetModelMgr().UnLock(info.Uuid, m.Name)
	}
	return nil
}

func (u *updateModelFileHandler) syncFile(currentUuid string) error {
	if !modeltask.GetModelMgr().LockGlobal() {
		return lockResError
	}
	defer modeltask.GetModelMgr().UnLockGlobal()
	modeltask.GetModelMgr().CancelTasks()
	fileList := modeltask.GetModelMgr().GetFileList()
	fileListBytes, err := json.Marshal(fileList)
	if err != nil {
		hwlog.RunLog.Errorf("cannot marshal fileList: %v", err)
		return fmt.Errorf("cannot marshal fileList: %v", err)
	}

	usedFiles, err := u.getUsedModelFiles()
	if err != nil {
		return err
	}

	opInfo := map[string]string{"fileList": string(fileListBytes)}
	content := types.OperateModelFileContent{
		Operate:     "sync",
		OperateInfo: opInfo,
		UsedFiles:   usedFiles.List(),
		CurrentUuid: currentUuid,
	}
	result, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:                constants.ModHandlerMgr,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptRaw,
		Resource:              constants.ActionModelFiles,
		Content:               content,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("send sync list failed: %v", err)
		return fmt.Errorf("send sync list failed: %v", err)
	}

	var toDelList modeltask.SyncList
	if err = json.Unmarshal([]byte(result), &toDelList); err != nil {
		hwlog.RunLog.Errorf("cannot unmarshal syncDelList: %v", err)
		return fmt.Errorf("cannot unmarshal syncDelList: %v", err)
	}
	modeltask.GetModelMgr().DelTaskByBriefs(toDelList.FileList)
	return nil
}

func (u *updateModelFileHandler) getUsedModelFiles() (*utils.Set, error) {
	metas, err := database.GetMetaRepository().GetByType(constants.ResourceTypePod)
	if err != nil {
		hwlog.RunLog.Errorf("get used pod id failed: %v", err)
		return nil, fmt.Errorf("get used pod id failed: %v", err)
	}
	set := utils.NewSet()
	for _, meta := range metas {
		var pod v1.Pod
		if err = json.Unmarshal([]byte(meta.Value), &pod); err != nil {
			hwlog.RunLog.Errorf("unmarshal pod failed: %v", err)
			return nil, fmt.Errorf("unmarshal pod failed: %v", err)
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.HostPath == nil {
				continue
			}
			if !strings.HasPrefix(volume.HostPath.Path, constants.ModeFileActiveDir) {
				continue
			}
			set.Add(volume.HostPath.Path)
		}
	}
	return set, nil
}

func (u *updateModelFileHandler) sendAsyncMsgToEdgeOm(operation, res string,
	content interface{}, transferStructIntoStr bool) error {
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.ModHandlerMgr,
		Destination:           constants.ModEdgeOm,
		Operation:             operation,
		Resource:              res,
		Content:               content,
		TransferStructIntoStr: transferStructIntoStr,
	})
	if err != nil {
		return fmt.Errorf("new message error: %v", err)
	}

	err = modulemgr.SendAsyncMessage(msg)
	if err != nil {
		return fmt.Errorf("send msg to edge om failed: %v", err)
	}

	return nil
}
