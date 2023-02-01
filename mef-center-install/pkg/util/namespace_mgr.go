// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"strconv"

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
	checkCmd := fmt.Sprintf("%s get namespaces | grep -w %s | awk '{print$2}'", CommandKubectl, namespaceReg)
	status, err := common.RunCommand("sh", false, "-c", checkCmd)
	if err != nil {
		hwlog.RunLog.Errorf("check namespace failed: %s", err.Error())
		return errors.New("get namespace failed")
	}
	if status == ActiveFlag {
		hwlog.RunLog.Info("the namespace has already existed")
		return nil
	}

	if status != "" && status != ActiveFlag {
		_, err = common.RunCommand(CommandKubectl, true, "delete", CommandNamespace, nm.namespace)
		if err != nil {
			hwlog.RunLog.Errorf("the namespace exists but not active, delete it failed: %s", err.Error())
			return errors.New("the namespace exists but not active, delete it failed")
		}
	}

	// namespace does not exist, then create
	hwlog.RunLog.Info("start to create namespace")
	_, err = common.RunCommand(CommandKubectl, true, "create", CommandNamespace, nm.namespace)
	if err != nil {
		hwlog.RunLog.Errorf("create namespace failed: %s", err.Error())
		return fmt.Errorf("create namespace failed")
	}

	hwlog.RunLog.Info("prepare namespace successful")
	return nil
}

func (nm *NamespaceMgr) checkNameSpaceExist() (bool, error) {
	namespaceReg := fmt.Sprintf("'^%s\\s'", nm.namespace)
	checkCmd := fmt.Sprintf("%s get namespaces | grep -w %s | wc -l", CommandKubectl, namespaceReg)
	ret, err := common.RunCommand("sh", false, "-c", checkCmd)
	if err != nil {
		hwlog.RunLog.Errorf("check namespace command exec failed: %s", err.Error())
		return false, errors.New("check namespace command exec failed")
	}

	if ret == strconv.Itoa(NamespaceExist) {
		return true, nil
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

	_, err = common.RunCommand(CommandKubectl, true, "delete", "namespace", nm.namespace)
	if err != nil {
		hwlog.RunLog.Errorf("delete %s namespace command exec failed: %s", nm.namespace, err.Error())
		return errors.New("delete namespace command exec failed")
	}
	return nil
}
