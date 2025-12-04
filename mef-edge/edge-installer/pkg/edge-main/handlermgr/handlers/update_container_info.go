// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers
package handlers

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

const (
	noModelFile = "none"
	hotUpdate   = "hot_update"
	coldUpdate  = "cold_update"
)

// UpdateContainerInfo update container info struct
type UpdateContainerInfo struct {
	containerInfo types.UpdateContainerInfo
}

// EffectModelFile effect model file
func (u *UpdateContainerInfo) EffectModelFile() error {
	switch u.getActiveType() {
	case noModelFile:
		return nil
	case hotUpdate:
		return u.dealHotUpdate()
	case coldUpdate:
		return u.dealColdUpdate()
	default:
		return errors.New("unknown active type")
	}
}

func (u *UpdateContainerInfo) getActiveType() string {
	activeType := noModelFile
	for _, container := range u.containerInfo.Container {
		for _, modelFile := range container.ModelFile {
			activeType = modelFile.ActiveType
			break
		}
	}
	return activeType
}

func (u *UpdateContainerInfo) dealHotUpdate() error {
	effectAllSuccess := true
	for _, container := range u.containerInfo.Container {
		for _, modelFile := range container.ModelFile {
			if err := u.effectModelFile(modelFile); err != nil {
				effectAllSuccess = false
				hwlog.RunLog.Errorf("model file takes effect failed, uuid:[%s], name:[%s], version:[%s], "+
					"error: %v", u.containerInfo.Uuid, modelFile.Name, modelFile.Version, err)
				continue
			}
			hwlog.RunLog.Infof("model file takes effect success, uuid:[%s], name:[%s], version:[%s]",
				u.containerInfo.Uuid, modelFile.Name, modelFile.Version)
		}
	}

	if !effectAllSuccess {
		return errors.New("not all model files take effect successfully")
	}
	return nil
}

func (u *UpdateContainerInfo) dealColdUpdate() error {
	podResource := constants.ActionPod + u.containerInfo.PodName
	if err := CheckPodRestartPolicy(podResource); err != nil {
		return fmt.Errorf("check pod restart policy failed, %v", err)
	}

	if err := u.dealHotUpdate(); err != nil {
		return err
	}

	if err := u.restartPodByEdgeOm(); err != nil {
		return errors.New("restart pod for model file effect failed")
	}
	return nil
}

func (u *UpdateContainerInfo) effectModelFile(modelFile types.ModelFileEffectInfo) error {
	if !modeltask.GetModelMgr().Lock(u.containerInfo.Uuid, modelFile.Name) {
		hwlog.RunLog.Error("lock model file database failed")
		return errors.New("lock model file database failed")
	}
	defer modeltask.GetModelMgr().UnLock(u.containerInfo.Uuid, modelFile.Name)

	activeTask := modeltask.GetModelMgr().GetActiveTask(u.containerInfo.Uuid, modelFile.Name)
	if activeTask != nil && activeTask.ModelFile.Version == modelFile.Version {
		hwlog.RunLog.Info("model file has already taken effect")
		return nil
	}

	notActiveTask := modeltask.GetModelMgr().GetNotActiveTask(u.containerInfo.Uuid, modelFile.Name)
	if notActiveTask == nil || notActiveTask.GetStatusType() != types.StatusInactive {
		hwlog.RunLog.Error("no model file to be activated")
		return errors.New("no model file to be activated")
	}

	if notActiveTask.ModelFile.Version != modelFile.Version {
		hwlog.RunLog.Error("the model file version is not correct")
		return errors.New("the model file version is not correct")
	}

	effectMapInfo := map[string]string{"uuid": u.containerInfo.Uuid, "name": modelFile.Name}
	content := types.OperateModelFileContent{
		Operate:     constants.OptUpdate,
		OperateInfo: effectMapInfo,
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
		hwlog.RunLog.Errorf("send effect model file message to edge om failed, error: %v", err)
		return errors.New("send effect model file message to edge om failed")
	}
	if result == constants.Failed {
		hwlog.RunLog.Error("model file takes effect failed by edge om")
		return errors.New("model file takes effect failed by edge om")
	}

	if err = modeltask.GetModelMgr().Active(u.containerInfo.Uuid, modelFile.Name, u.containerInfo.PodUid); err != nil {
		hwlog.RunLog.Errorf("update model file status to active failed, error: %v", err)
		return errors.New("update model file status to active failed")
	}
	return nil
}

func (u *UpdateContainerInfo) restartPodByEdgeOm() error {
	result, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:      constants.ModHandlerMgr,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptRestart,
		Resource:    constants.ActionPod,
		Content:     u.containerInfo.PodName,
	})
	if err != nil {
		hwlog.RunLog.Errorf("send restart pod message to edge om failed, error: %v", err)
		return errors.New("send restart pod message to edge om failed")
	}
	if result == constants.Failed {
		hwlog.RunLog.Error("restart pod failed by edge om")
		return errors.New("restart pod failed by edge om")
	}
	hwlog.RunLog.Info("restart pod by edge om success")
	return nil
}
