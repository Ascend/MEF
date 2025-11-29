// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msglistchecker
package msglistchecker

import (
	"regexp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/common/constants"
)

// MsgHeaderValidatorIntf [interface] for msg validator interface
type MsgHeaderValidatorIntf interface {
	Check(msg *model.Message) bool
}

// MsgHeaderValidator [struct] to check msg is in whitelist or not
type MsgHeaderValidator struct {
	whiteList map[messageRoute][]resourceInfo
}

// MefMsgHeaderValidator [struct] to check mef msg is in whitelist or not
type MefMsgHeaderValidator struct {
	MsgHeaderValidator
}

// NewCloudCoreMsgHeaderValidator [method] new cloud core checker object
func NewCloudCoreMsgHeaderValidator(upstream bool) MsgHeaderValidator {
	if upstream {
		return MsgHeaderValidator{whiteList: cloudCoreUpstreamAllowedRoutes}
	}
	return MsgHeaderValidator{whiteList: cloudCoreDownstreamAllowedRoutes}
}

// NewFdMsgHeaderValidator [method] new fd checker object
func NewFdMsgHeaderValidator() MsgHeaderValidator {
	return MsgHeaderValidator{whiteList: fdDownstreamAllowedRoutes}
}

// NewMefMsgHeaderValidator [method] new mef checker object
func NewMefMsgHeaderValidator() MefMsgHeaderValidator {
	return MefMsgHeaderValidator{MsgHeaderValidator: MsgHeaderValidator{whiteList: mefUpstreamAllowedRoutes}}
}

// Check [method] Check msg valid
func (mhv MsgHeaderValidator) Check(msg *model.Message) bool {
	if msg == nil {
		return false
	}

	checkItems := []func(msg *model.Message) bool{
		mhv.checkHeaderId,
		mhv.checkHeaderParentId,
		mhv.checkHeaderResourceVersion,
		mhv.checkRoute,
	}

	for _, check := range checkItems {
		if !check(msg) {
			return false
		}
	}

	return true
}

func (mhv MsgHeaderValidator) checkHeaderId(msg *model.Message) bool {
	if !checker.RegexStringChecker(msg.Header.ID, "^"+constants.UUIDRegex+"$") {
		hwlog.RunLog.Error("UUID check failed")
		return false
	}
	return true
}

func (mhv MsgHeaderValidator) checkHeaderParentId(msg *model.Message) bool {
	if msg.Header.ParentID == "" {
		return true
	}

	if !checker.RegexStringChecker(msg.Header.ParentID, "^"+constants.UUIDRegex+"$") {
		hwlog.RunLog.Error("parent id check failed")
		return false
	}

	return true
}

func (mhv MsgHeaderValidator) checkHeaderResourceVersion(msg *model.Message) bool {
	if !checker.RegexStringChecker(msg.Header.ResourceVersion, constants.ResourceVersionRegex) {
		hwlog.RunLog.Error("ResourceVersion check failed")
		return false
	}
	return true
}
func (mhv MsgHeaderValidator) checkRoute(message *model.Message) bool {
	route := messageRoute{source: message.KubeEdgeRouter.Source,
		group: message.KubeEdgeRouter.Group, operation: message.KubeEdgeRouter.Operation}

	resourceInfos, exists := mhv.whiteList[route]
	if !exists {
		hwlog.RunLog.Errorf("not supported message route, Router: [%+v]", message.KubeEdgeRouter)
		return false
	}
	for _, res := range resourceInfos {
		if mhv.checkResource(res, message) {
			return true
		}
	}
	hwlog.RunLog.Errorf("not supported message, Header: [%+v]; Router: [%+v]", message.Header,
		message.KubeEdgeRouter)
	return false
}

func (mhv MsgHeaderValidator) checkResource(resInfo resourceInfo, message *model.Message) bool {
	if !mhv.checkParentID(resInfo, message.Header.ParentID) {
		return false
	}
	if !mhv.checkSyncFlag(resInfo, message.Header.Sync) {
		return false
	}
	if !mhv.checkResourceName(resInfo, message.KubeEdgeRouter.Resource) {
		return false
	}
	return true
}

func (mhv MsgHeaderValidator) checkParentID(resInfo resourceInfo, parentID string) bool {
	if parentID == "" && resInfo.hasParentID {
		return false
	}
	if parentID != "" && !resInfo.hasParentID {
		return false
	}
	return true
}

func (mhv MsgHeaderValidator) checkSyncFlag(resInfo resourceInfo, syncFlag bool) bool {
	return syncFlag == resInfo.sync
}

func (mhv MsgHeaderValidator) checkResourceName(resInfo resourceInfo, resource string) bool {
	switch resInfo.resource.(type) {
	case *regexp.Regexp:
		reg := (resInfo.resource).(*regexp.Regexp)
		return reg.MatchString(resource)
	case string:
		return resource == resInfo.resource
	default:
		hwlog.RunLog.Error("not supported resource type")
		return false
	}
}

// Check [method] check mef msg valid
func (mhv MefMsgHeaderValidator) Check(msg *model.Message) bool {
	if msg == nil {
		return false
	}

	checkItems := []func(msg *model.Message) bool{
		mhv.checkHeaderId,
		mhv.checkHeaderParentId,
		mhv.checkHeaderResourceVersion,
		mhv.checkRoute,
	}

	for _, checkItem := range checkItems {
		if !checkItem(msg) {
			return false
		}
	}

	return true
}

func (mhv MefMsgHeaderValidator) checkRoute(message *model.Message) bool {
	route := messageRoute{source: "", group: "", operation: message.KubeEdgeRouter.Operation}

	resourceInfos, exists := mhv.whiteList[route]
	if !exists {
		hwlog.RunLog.Errorf("not supported message route, Router: [%+v]", message.KubeEdgeRouter)
		return false
	}
	for _, res := range resourceInfos {
		if mhv.checkResource(res, message) {
			return true
		}
	}
	hwlog.RunLog.Errorf("not supported message, Header: [%+v]; Router: [%+v]", message.Header,
		message.KubeEdgeRouter)
	return false
}

func (mhv MefMsgHeaderValidator) checkHeaderResourceVersion(msg *model.Message) bool {
	if !checker.RegexStringChecker(msg.Header.ResourceVersion, constants.MefResourceVersionRegex) {
		hwlog.RunLog.Error("ResourceVersion check failed")
		return false
	}
	return true
}
