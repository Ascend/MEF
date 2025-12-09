// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cfgrestore
package cfgrestore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/database"
)

const (
	percent25  = 25
	percent50  = 50
	percent75  = 75
	percent100 = 100

	syncMsgTimeout = time.Second * 5
)

type worker struct {
	ctx  context.Context
	busy int32
}

type resourceInfo struct {
	Metadata v1.ObjectMeta `json:"metadata"`
}

// DeletePodsData delete pods data
func (w *worker) DeletePodsData() {
	if !atomic.CompareAndSwapInt32(&w.busy, 0, 1) {
		feedbackFailed(constants.ResourceTypePodsData, "cfgRestore busy")
		atomic.AddInt32(&goRoutineCount, -1)
		return
	}
	defer atomic.StoreInt32(&w.busy, 0)

	defer atomic.AddInt32(&goRoutineCount, -1)
	deletePodsData()
	deleteAllModelFiles()
}

func deletePodsData() {
	if err := deleteAllPods(); err != nil {
		hwlog.RunLog.Errorf("delete pods failed, %v", err)
		feedbackFailed(constants.ResourceTypePodsData, "delete pods failed")
		return
	}

	hwlog.RunLog.Info("delete pods successful")
	feedbackPodsDataProcessing(percent25)

	if err := deleteAllConfigMaps(); err != nil {
		hwlog.RunLog.Errorf("delete configMap failed, %v", err)
		feedbackFailed(constants.ResourceTypePodsData, "delete configMap failed")
		return
	}

	hwlog.RunLog.Info("delete configMap successful")
	feedbackPodsDataProcessing(percent50)

	if err := deleteAllSecrets(); err != nil {
		hwlog.RunLog.Error("delete secret failed")
		feedbackFailed(constants.ResourceTypePodsData, "delete secret failed")
		return
	}

	hwlog.RunLog.Info("delete secret successful")
	feedbackPodsDataProcessing(percent75)

}

func deleteAllPods() error {
	return deleteResources(constants.ResourceTypePod)
}

func deleteAllConfigMaps() error {
	return deleteResources(constants.ResourceTypeConfigMap)
}

func deleteAllSecrets() error {
	if err := deleteResources(constants.ResourceTypeSecret); err != nil {
		return err
	}
	resource := resourceInfo{Metadata: v1.ObjectMeta{UID: constants.ActionSecret}}
	return deleteResource(constants.ActionSecret, resource)
}

func deleteResources(resourceType string) error {
	resourceMetas, err := database.GetMetaRepository().GetByType(resourceType)
	if err != nil {
		return err
	}

	for _, meta := range resourceMetas {

		var resource resourceInfo
		if err := json.Unmarshal([]byte(meta.Value), &resource); err != nil {
			return fmt.Errorf("unmarshal resource failed, resource=%s, error=%v", meta.Key, err)
		}

		if err := deleteResource(meta.Key, resource); err != nil {
			return err
		}
	}
	return nil
}

func deleteResource(resourceName string, resource resourceInfo) error {
	req, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create message failed, resource=%s, error=%v", resourceName, err)
	}

	req.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.CfgRestore,
		Group:     constants.ResourceModule,
		Operation: constants.OptDelete,
		Resource:  resourceName,
	}
	req.Header.Timestamp = time.Now().UnixMilli()
	req.Header.ID = req.Header.Id
	req.Header.Sync = false

	now := v1.NewTime(time.Now())
	resource.Metadata.DeletionTimestamp = &now
	resource.Metadata.DeletionGracePeriodSeconds = new(int64)
	if err = req.FillContent(resource, false); err != nil {
		return fmt.Errorf("fill resource info into content failed: %v", err)
	}
	req.SetRouter(constants.CfgRestore, constants.ModEdgeCore, constants.OptDelete, resourceName)

	resp, err := modulemgr.SendSyncMessage(req, syncMsgTimeout)
	if err != nil {
		return fmt.Errorf("send message failed, resource=%s, error=%v", resourceName, err)
	}
	var respCntStr string
	if err = resp.ParseContent(&respCntStr); err != nil {
		return fmt.Errorf("get delete resource resp failed: %v", err)
	}
	if respCntStr != constants.OK {
		return fmt.Errorf("edgecore respond error, resource=%s, message=%v", resourceName, respCntStr)
	}

	if err := database.GetMetaRepository().DeleteByKey(resourceName); err != nil {
		return fmt.Errorf("delete resource from database failed, resource=%s", resourceName)
	}
	return nil
}

func feedbackFailed(topic, reason string) {
	feedbackProgress(topic, percent100, constants.ResultFailed, reason)
}

func feedbackPodsDataProcessing(percentage int) {
	result := constants.ResultProcessing
	if percentage == percent100 {
		result = constants.Success
	}
	feedbackProgress(constants.ResourceTypePodsData, percentage, result, "")
}

func feedbackProgress(topic string, percentage int, result string, reason string) {
	content := config.ProgressTip{
		Topic:      topic,
		Result:     result,
		Percentage: fmt.Sprintf("%d%%", percentage),
		Reason:     reason,
	}

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		return
	}

	respMsg.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.SourceHardware,
		Group:     constants.GroupHub,
		Operation: constants.OptUpdate,
		Resource:  constants.ResConfigResult,
	}
	respMsg.Header.Timestamp = time.Now().UnixMilli()
	respMsg.Header.ID = respMsg.Header.Id
	respMsg.Header.Sync = false
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptUpdate, constants.ResConfigResult)
	if err = respMsg.FillContent(content, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}

	if err = modulemgr.SendAsyncMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("send message failed, %v", err)
	}
}

func deleteAllModelFiles() {
	delMsg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:      constants.CfgRestore,
		Destination: constants.ModHandlerMgr,
		Operation:   constants.OptDelete,
		Resource:    constants.ActionPodsData,
		Content:     nil,
	})
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		feedbackFailed(constants.ResourceTypePodsData, "delete all model file failed")
		return
	}
	if err = modulemgr.SendAsyncMessage(delMsg); err != nil {
		hwlog.RunLog.Errorf("send delete all model file message failed, %v", err)
		feedbackFailed(constants.ResourceTypePodsData, "delete all model file failed")
	}
}
