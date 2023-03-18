// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// NamespaceMgr is the struct that manages the func related to namespace operation
type NamespaceMgr struct {
	namespace string
}

// NewNamespaceMgr is the func that used to create a NamespaceMgr struct
func NewNamespaceMgr(namespace string) *NamespaceMgr {
	return &NamespaceMgr{namespace: namespace}
}

func (nm *NamespaceMgr) prepareNameSpace() error {
	hwlog.RunLog.Info("start to prepare namespace")
	namespaceReg := fmt.Sprintf("'^%s\\s'", nm.namespace)
	ret, err := common.RunCommand(CommandKubectl, true, common.DefCmdTimeoutSec, "get", "namespace")
	if err != nil {
		hwlog.RunLog.Errorf("check namespace failed: %s", err.Error())
		return errors.New("get namespace failed")
	}

	var status string
	lines := strings.Split(ret, "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(namespaceReg, line)
		if err != nil {
			hwlog.RunLog.Errorf("check namespace %s's status on reg match failed: %s", nm.namespace, err.Error())
			return errors.New("check namespace status failed")
		}
		if found {
			data := strings.Split(line, " ")
			if len(data) < NamespaceStatusLoc+1 {
				hwlog.RunLog.Errorf("split namespace ret failed")
				return errors.New("split namespace ret failed")
			}
			status = data[NamespaceStatusLoc]
		}
	}
	if status == ActiveFlag {
		hwlog.RunLog.Info("the namespace has already existed")
		return nil
	}

	if status != "" && status != ActiveFlag {
		_, err = common.RunCommand(CommandKubectl, true, DeleteNsTimeoutSec, "delete", CommandNamespace, nm.namespace)
		if err != nil {
			hwlog.RunLog.Errorf("the namespace exists but not active, delete it failed: %s", err.Error())
			return errors.New("the namespace exists but not active, delete it failed")
		}
	}

	// namespace does not exist, then create
	hwlog.RunLog.Info("start to create namespace")
	_, err = common.RunCommand(CommandKubectl, true, common.DefCmdTimeoutSec,
		"create", CommandNamespace, nm.namespace)
	if err != nil {
		hwlog.RunLog.Errorf("create namespace failed: %s", err.Error())
		return fmt.Errorf("create namespace failed")
	}

	hwlog.RunLog.Info("prepare namespace successful")
	return nil
}

func (nm *NamespaceMgr) checkNameSpaceExist() (bool, error) {
	namespaceReg := fmt.Sprintf("'^%s\\s'", nm.namespace)
	ret, err := common.RunCommand(CommandKubectl, true, common.DefCmdTimeoutSec, "get", "namespace")
	if err != nil {
		hwlog.RunLog.Errorf("check namespace command exec failed: %s", err.Error())
		return false, errors.New("check namespace command exec failed")
	}

	lines := strings.Split(ret, "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(namespaceReg, line)
		if err != nil {
			hwlog.RunLog.Errorf("check if namespace %s's exists on reg match failed: %s", nm.namespace, err.Error())
			return false, errors.New("check if namespace exists failed")
		}
		if found {
			return true, nil
		}
	}

	return false, nil
}

// ClearNamespace is the func that used to clear the namespace
func (nm *NamespaceMgr) ClearNamespace() error {
	ret, err := nm.checkNameSpaceExist()
	if err != nil {
		return err
	}
	if !ret {
		hwlog.RunLog.Warnf("namespace %s does not exist, no need to clear", nm.namespace)
		return nil
	}

	_, err = common.RunCommand(CommandKubectl, true, DeleteNsTimeoutSec,
		"delete", "namespace", nm.namespace)
	if err != nil {
		hwlog.RunLog.Errorf("delete %s namespace command exec failed: %s", nm.namespace, err.Error())
		return errors.New("delete namespace command exec failed")
	}
	return nil
}
