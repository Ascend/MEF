// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msgchecker
package msgchecker

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	ctypes "edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/checker/containerinfochecker"
	"edge-installer/pkg/edge-main/common/checker/modelchecker"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

var msgResourceMap = make(map[interface{}]interface{})

// Thresholds for system res check before process update message
const (
	minRemainingStorageThreshold = uint64(50 * constants.MB)
	minRemainingMemoryThreshold  = uint64(200 * constants.MB)
	maxCPUAverageUsageThreshold  = float64(0.7)
)

func init() {
	msgResourceMap[regexp.MustCompile("^mef-user/pod/"+constants.MefPodNameRegex+"$")] = &types.Pod{}
	msgResourceMap[regexp.MustCompile("^mef-user/podpatch/"+constants.MefPodNameRegex+"$")] = &types.PodPatch{}
	msgResourceMap[regexp.MustCompile("^default/node/"+constants.NodeNameRegx+"$")] = &types.Node{}
	msgResourceMap[regexp.MustCompile("^default/nodepatch/"+constants.NodeNameRegx+"$")] = &types.NodePatch{}
	msgResourceMap[regexp.MustCompile("^kube-node-lease/lease/"+constants.NodeNameRegx+"$")] = &types.Lease{}
	msgResourceMap["mef-user/secret/image-pull-secret"] = &types.Secret{}

	msgResourceMap["websocket/secret/fusion-director-docker-registry-secret"] = &types.Secret{}
	msgResourceMap[regexp.MustCompile("^websocket/pod/"+constants.FdPodNameRegex+"$")] = &types.Pod{}
	msgResourceMap[regexp.MustCompile("^websocket/configmap/"+constants.ConfigmapNameRegex+"$")] = &types.ConfigMap{}
	msgResourceMap["websocket/npu_sharing"] = &types.NpuSharingInfo{}
	msgResourceMap["websocket/container_info"] = &ctypes.UpdateContainerInfo{}
	msgResourceMap["websocket/modelfiles"] = &ctypes.ModelFileInfo{}

	msgResourceMap["websocket/pods_data"] = struct{}{}
	msgResourceMap["/edge/system/image-cert-info"] = struct{}{}
	msgResourceMap["/edge/system/all-alarm"] = struct{}{}
}

func isResourceMatched(resource string, value interface{}) bool {
	switch value.(type) {
	case *regexp.Regexp:
		reg := value.(*regexp.Regexp)
		return reg.MatchString(resource)
	case string:
		return resource == value.(string)
	default:
		return false
	}
}

// MsgValidator [struct] to check msg is valid or not
type MsgValidator struct {
	netType   string
	resource  string
	operation string
	data      []byte

	headerValidator msglistchecker.MsgHeaderValidatorIntf
}

// NewMsgValidator [method] creat msg check obj
func NewMsgValidator(headerValidator msglistchecker.MsgHeaderValidatorIntf) MsgValidator {
	return MsgValidator{headerValidator: headerValidator}
}

func (mv *MsgValidator) check() error {
	g := gin.Context{}

	for k, obj := range msgResourceMap {
		if !isResourceMatched(mv.resource, k) {
			continue
		}

		if obj == struct{}{} {
			return nil
		}

		if reflect.TypeOf(obj).Kind() != reflect.Pointer {
			return errors.New("msg obj type is not pointer")
		}

		newObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()

		g.Request = &http.Request{Body: ioutil.NopCloser(bytes.NewReader(mv.data))}
		if err := g.ShouldBindJSON(newObj); err != nil {
			return err
		}

		return mv.auxCheck(newObj)
	}

	return fmt.Errorf("resource: %s not matched", mv.resource)
}

// Check [method] to check msg valid
func (mv *MsgValidator) Check(msg *model.Message) error {
	if mv.headerValidator != nil && !mv.headerValidator.Check(msg) {
		return fmt.Errorf("check msg header failed")
	}

	return mv.checkMsgContent(msg)
}

func (mv *MsgValidator) checkMsgContent(msg *model.Message) error {
	data := model.UnformatMsg(msg.Content)

	if len(data) == 0 || string(data) == "OK" || string(data) == "null" {
		return nil
	}
	netType, err := configpara.GetNetType()
	if err != nil {
		return err
	}
	var validator = MsgValidator{
		netType:   netType,
		data:      data,
		operation: msg.KubeEdgeRouter.Operation,
		resource:  msg.KubeEdgeRouter.Resource,
	}
	return validator.check()
}

func (mv *MsgValidator) auxCheck(obj interface{}) error {
	if err := mv.checkSystemResources(obj); err != nil {
		return err
	}

	switch value := obj.(type) {
	case *types.Pod:
		return mv.auxCheckPod(value)
	case *types.PodPatch:
		return mv.auxCheckPodPatch(value)
	case *types.ConfigMap:
		return mv.auxCheckCm(value)
	case *types.Secret:
		return mv.auxCheckSecret(value)
	case *ctypes.UpdateContainerInfo:
		return containerinfochecker.CheckContainerInfo(mv.data)
	case *ctypes.ModelFileInfo:
		return modelchecker.CheckModelFileMsg(mv.data)
	default:
		return nil
	}
}

func skipCheckResource(obj interface{}, operation string) bool {
	var gracefulDeleteFlag bool
	switch value := obj.(type) {
	case *types.Pod:
		gracefulDeleteFlag = isPodGraceDelete(value.DeletionTimestamp)
	case *types.PodPatch:
		gracefulDeleteFlag = isPodGraceDelete(value.Object.DeletionTimestamp)
	default:
		gracefulDeleteFlag = false
	}

	// check system resource availability only when data is updated,
	// no need to check system resource availability when delete pod  gracefully.
	if operation == constants.OptDelete || (operation == constants.OptUpdate && gracefulDeleteFlag) {
		return true
	}

	return false
}

func (mv *MsgValidator) checkSystemResources(obj interface{}) error {
	if skipCheckResource(obj, mv.operation) {
		return nil
	}

	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Error("get edge main config dir error")
		return err
	}
	dbPath := filepath.Join(cfgDir, constants.DbEdgeMainPath)

	if ok, err := util.IsSystemStorageEnough(dbPath, minRemainingStorageThreshold); err != nil || !ok {
		return errors.New("system storage available space not enough")
	}
	if ok, err := util.IsSystemMemoryEnough(minRemainingMemoryThreshold); err != nil || !ok {
		return errors.New("system memory available space not enough")
	}
	if !util.IsSystemCPUAvailable(maxCPUAverageUsageThreshold) {
		return errors.New("system cpu is busy")
	}
	return nil
}
