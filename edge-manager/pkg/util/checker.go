// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base checker used
package util

import (
	"regexp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// CheckInnerMsg checks internal messages
func CheckInnerMsg(msg *model.Message) bool {
	if msg == nil {
		return false
	}
	msgId := msg.GetId()
	msgParentId := msg.GetParentId()
	timestamp := msg.GetTimestamp()

	if msgId == "" || timestamp == 0 {
		hwlog.RunLog.Error("check message id or timestamp failed")
		return false
	}
	if (msg.GetIsSync() && msgParentId == "") || (!msg.GetIsSync() && msgParentId != "") {
		hwlog.RunLog.Error("sync message does not match parent id")
		return false
	}
	return true
}

// CheckInt checks whether the value is within the range
func CheckInt(value, min, max int) bool {
	return value >= min && value <= max
}

// CheckNameFormat checks whether the name format is valid
func CheckNameFormat(name string) bool {
	if !RegexStringChecker(name, "^[0-9a-zA-Z]{1,32}$") {
		return false
	}
	return true
}

// RegexStringChecker checks whether the str is valid
func RegexStringChecker(str, matchStr string) bool {
	strSlice := regexp.MustCompile(matchStr)
	return strSlice.MatchString(str)
}
