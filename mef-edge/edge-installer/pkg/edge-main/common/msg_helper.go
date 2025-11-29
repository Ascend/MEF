// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

const (
	ipKey             = "ip"
	portKey           = "port"
	maxMsgOptLogCache = 600
)

// messages from MindXOM to MEF. key: resource.  value: operation.
var mindXomInnerMsgWhiteList = map[string]string{
	// operation log of update /edge/system/image-cert-info recorded in edge-om proc, ignore here
	constants.ResImageCertInfo: constants.OptUpdate,
	constants.QueryAllAlarm:    constants.OptQuery,
}

// messages no need to write opt log. key: resource.  value: operation.
var msgSkipOptLogging = map[string]string{
	constants.ResPodStatus: constants.OptUpdate,
	constants.ResStatic:    constants.OptUpdate,
}

var fdOptMessages []*model.Message

// NewFDPodStatusMsg create a msg to FD, parameters will be injected to kubeEdgeRouter
func NewFDPodStatusMsg(source, group, operation, resource string) (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return nil, fmt.Errorf("create to FD message error: %v", err)
	}
	msg.Header.ID = msg.Header.Id
	msg.SetKubeEdgeRouter(source, group, operation, resource)
	allPodStatus, err := statusmanager.LoadPodsDataForFd()
	if err != nil {
		return nil, err
	}
	if err = msg.FillContent(allPodStatus); err != nil {
		return nil, fmt.Errorf("fill all pod status into content failed: %v", err)
	}
	msg.SetRouter(source, constants.ModDeviceOm, operation, resource)
	return msg, nil
}

// NewFDNodeStatusMsg create a msg to FD
func NewFDNodeStatusMsg(source, group, option, resource string) (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return nil, fmt.Errorf("create msg error: %v", err)
	}
	nodeStatus, err := statusmanager.LoadNodeDataForFd()
	if err != nil {
		return nil, fmt.Errorf("get node status error: %v", err)
	}
	msg.Header.ID = msg.Header.Id
	msg.Header.Sync = true
	msg.SetRouter(source, group, option, resource)
	msg.SetKubeEdgeRouter(
		constants.EdgedModule, "meta", constants.OptUpdate, constants.ModifiedNodePrefix+nodeStatus.Status.Content.Name)
	if err = msg.FillContent(nodeStatus); err != nil {
		return nil, fmt.Errorf("fill node status into content failed: %v", err)
	}
	return msg, nil
}

// MsgOutProcess process come out message
func MsgOutProcess(msg *model.Message) (*model.Message, error) {
	if msg == nil {
		return nil, fmt.Errorf("invalid msg data")
	}
	// 1.process headers
	if msg.Header.ID == "" && msg.Header.Id != "" {
		msg.Header.ID = msg.Header.Id
	}
	if msg.Header.ParentID == "" && msg.Header.ParentId != "" {
		msg.Header.ParentID = msg.Header.ParentId
	}
	if msg.Header.ResourceVersion == "" && msg.Header.Version != "" {
		msg.Header.ResourceVersion = msg.Header.Version
	}
	if msg.Header.Sync == false && msg.Header.IsSync == true {
		msg.Header.Sync = true
	}

	// 2.process route
	if msg.KubeEdgeRouter.Source == "" && msg.Router.Source != "" {
		msg.KubeEdgeRouter.Source = msg.Router.Source
	}
	if msg.KubeEdgeRouter.Operation == "" && msg.Router.Option != "" {
		msg.KubeEdgeRouter.Operation = msg.Router.Option
	}
	if msg.KubeEdgeRouter.Group == "" && msg.Router.Destination != "" {
		msg.KubeEdgeRouter.Group = msg.Router.Destination
	}
	if msg.KubeEdgeRouter.Resource == "" && msg.Router.Resource != "" {
		msg.KubeEdgeRouter.Resource = msg.Router.Resource
	}
	return msg, nil
}

// MsgInProcess convert come in message
func MsgInProcess(msg *model.Message) (*model.Message, error) {
	if msg == nil {
		return nil, fmt.Errorf("invalid msg data")
	}
	// 1.process headers
	if msg.Header.ID != "" {
		msg.Header.Id = msg.Header.ID
	}
	if msg.Header.ParentID != "" {
		msg.Header.ParentId = msg.Header.ParentID
	}
	if msg.Header.ResourceVersion != "" {
		msg.Header.Version = msg.Header.ResourceVersion
	}
	if msg.Header.Sync == true {
		msg.Header.IsSync = true
	}

	// 2.process route
	if msg.KubeEdgeRouter.Source != "" {
		msg.Router.Source = msg.KubeEdgeRouter.Source
	}
	if msg.KubeEdgeRouter.Operation != "" {
		msg.Router.Option = msg.KubeEdgeRouter.Operation
	}
	if msg.KubeEdgeRouter.Resource != "" {
		msg.Router.Resource = msg.KubeEdgeRouter.Resource
	}

	return msg, nil
}

func cleanKubeedgeMessage(msg *model.Message) {
	msg.Router.Option = ""
	msg.Router.Resource = ""
	msg.Router.Source = ""
	msg.Router.Destination = ""

	msg.Header.Id = ""
	msg.Header.ParentId = ""
	msg.Header.IsSync = false
	msg.Header.Version = ""
	msg.Header.NodeId = ""
	msg.Header.PeerInfo = model.MsgPeerInfo{}
}

// UnmarshalKubeedgeMessage unmarshals a kubeedge message and ignores inner fields
func UnmarshalKubeedgeMessage(dataBytes []byte, msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("invalid msg data")
	}

	var newMsg model.Message
	if err := json.Unmarshal(dataBytes, &newMsg); err != nil {
		return err
	}
	cleanKubeedgeMessage(&newMsg)
	*msg = newMsg
	return nil
}

// MarshalKubeedgeMessage marshals a kubeedge message and ignores inner fields
func MarshalKubeedgeMessage(msg *model.Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("invalid msg data")
	}

	copiedMsg := *msg
	cleanKubeedgeMessage(&copiedMsg)
	return json.Marshal(copiedMsg)
}

// MsgOptLog write operation log for downstream msg
func MsgOptLog(msg *model.Message) {
	// skip minXOM inner message
	if msg == nil || IsValidInnerMindXomMsg(msg) {
		return
	}
	fdIp, err := GetFdIp()
	if err != nil {
		hwlog.RunLog.Warnf("get fd ip error: %v", err)
	}
	operation := msg.KubeEdgeRouter.Operation
	resource := msg.KubeEdgeRouter.Resource
	id := msg.Header.ID
	if strings.HasPrefix(resource, constants.ActionPod) && operation == constants.OptRestart {
		recordOpLog(fdIp, operation, resource, constants.Start, id)
		return
	}

	var savedMsg model.Message
	savedMsg.Header.ID = msg.Header.ID
	savedMsg.KubeEdgeRouter.Operation = msg.KubeEdgeRouter.Operation
	savedMsg.KubeEdgeRouter.Resource = msg.KubeEdgeRouter.Resource
	if len(fdOptMessages) >= maxMsgOptLogCache {
		var temp []*model.Message
		temp = append(temp, fdOptMessages[1:]...)
		fdOptMessages = temp
	}
	fdOptMessages = append(fdOptMessages, &savedMsg)

	recordOpLog(fdIp, operation, resource, constants.Start, id)
}

// MsgResultOptLog write operation log for result of downstream msg
func MsgResultOptLog(msg *model.Message) {
	if ignoreMsgOptLog(msg) {
		return
	}
	fdIp, err := GetFdIp()
	if err != nil {
		hwlog.RunLog.Warnf("get fd ip error: %v", err)
	}

	if msg.KubeEdgeRouter.Resource == constants.ResConfigResult {
		oplogByContent(msg, fdIp)
		return
	}

	originalOpt, originalRes, err := getOriginalOptAndRes(msg)
	if err != nil {
		hwlog.RunLog.Warn(err.Error())
		return
	}
	originalId := msg.Header.ParentID

	switch msg.KubeEdgeRouter.Operation {
	case constants.OptError:
		recordOpLog(fdIp, originalOpt, originalRes, constants.Failed, originalId)
	case constants.OptResponse:
		contentValue := string(model.UnformatMsg(msg.Content))
		if contentValue != constants.OK {
			recordOpLog(fdIp, originalOpt, originalRes, constants.Failed, originalId)
			return
		}
		recordOpLog(fdIp, originalOpt, originalRes, constants.Success, originalId)
	default:
		hwlog.RunLog.Error("error msg operation type for operation log")
	}
}

func getOriginalOptAndRes(msg *model.Message) (string, string, error) {
	if msg.Header.ParentID != "" {
		// fd删除pod时三条消息ID相同，所以需要按照最近的一条消息进行匹配
		for i := len(fdOptMessages) - 1; i >= 0; i-- {
			if msg.Header.ParentID == fdOptMessages[i].Header.ID {
				return fdOptMessages[i].KubeEdgeRouter.Operation, fdOptMessages[i].KubeEdgeRouter.Resource, nil
			}
		}
	}
	if msg.KubeEdgeRouter.Resource == constants.ResConfigResult {
		return msg.KubeEdgeRouter.Operation, msg.KubeEdgeRouter.Resource, nil
	}
	return "", "", errors.New("get original operate and resource failed")
}

func recordOpLog(fdIp, operation, resource, result, msgId string) {
	switch result {
	case constants.Start, constants.Success:
		hwlog.OpLog.Infof("[%s@%s] %s %s %s, the message(id:%s) is forwarded from [%s:%s]",
			constants.DeviceOmModule, constants.LocalIp, operation, resource, result, msgId, constants.FD, fdIp)
	case constants.Failed:
		hwlog.OpLog.Errorf("[%s@%s] %s %s %s, the message(id:%s) is forwarded from [%s:%s]",
			constants.DeviceOmModule, constants.LocalIp, operation, resource, result, msgId, constants.FD, fdIp)
	default:
		hwlog.RunLog.Error("error operation log result")
	}
}

// if no need to write operation log, return true
func ignoreMsgOptLog(msg *model.Message) bool {
	if msg == nil {
		return true
	}
	operation := msg.KubeEdgeRouter.Operation
	resource := msg.KubeEdgeRouter.Resource
	if operation == constants.OptQuery || operation == constants.OptReport {
		return true
	}
	if strings.HasPrefix(resource, constants.ModifiedNodePrefix) || resource == constants.ResAlarm ||
		resource == constants.ActionModelFileInfo || resource == constants.QueryAllAlarm {
		return true
	}

	if opt, ok := msgSkipOptLogging[resource]; ok && opt == operation {
		return true
	}
	return false
}

func oplogByContent(message *model.Message, fdIp string) {
	var configResult config.ProgressTip
	if err := message.ParseContent(&configResult); err != nil {
		hwlog.RunLog.Errorf("parse config result failed: %s", err.Error())
		return
	}
	if configResult.Result == "" || configResult.Result == constants.ResultProcessing {
		return
	}
	msgId := message.Header.ID
	if configResult.Topic == constants.ResourceTypePodsData {
		recordOpLog(fdIp, constants.OptDelete, constants.ActionPodsData, configResult.Result, msgId)
	} else {
		recordOpLog(fdIp, message.KubeEdgeRouter.Operation, configResult.Topic, configResult.Result, msgId)
	}
}

// IsValidInnerMindXomMsg check is mindXOM inner message is valid
func IsValidInnerMindXomMsg(msg *model.Message) bool {
	operation := getMsgOperation(msg)
	resource := getMsgResource(msg)
	if opt, ok := mindXomInnerMsgWhiteList[resource]; ok && opt == operation {
		return true
	}
	return false
}

// UpdateFdAddrInfo update fd ip info when receiving the first message from mindXOM
func UpdateFdAddrInfo(msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("invalid message")
	}
	operation := getMsgOperation(msg)
	resource := getMsgResource(msg)
	if operation != constants.OptUpdate || resource != constants.ResImageCertInfo {
		return nil
	}
	content := make(map[string]string)
	var jsonData string
	if err := msg.ParseContent(&jsonData); err != nil {
		return fmt.Errorf("get json data failed: %v", err)
	}

	if err := json.Unmarshal([]byte(jsonData), &content); err != nil {
		return err
	}
	ipStr, ok := content[ipKey]
	if !ok {
		return errors.New("content ip key does not exist")
	}
	portStr, ok := content[portKey]
	if !ok {
		return errors.New("content port key does not exist")
	}
	if ipStr == "" || portStr == "" {
		return fmt.Errorf("invalid fd ip or port")
	}
	addr := ipStr + ":" + portStr
	if err := SetFdIp(constants.ModDeviceOm, addr); err != nil {
		return err
	}
	return nil
}

func getMsgOperation(msg *model.Message) string {
	operation := msg.KubeEdgeRouter.Operation
	if operation == "" {
		operation = msg.Router.Option
	}
	if operation == constants.OptResponse {
		operation = constants.OptRaw
	}
	return operation
}

func getMsgResource(msg *model.Message) string {
	resource := msg.KubeEdgeRouter.Resource
	if resource == "" {
		resource = msg.Router.Resource
	}
	return resource
}
