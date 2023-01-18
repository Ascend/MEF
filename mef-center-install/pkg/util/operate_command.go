// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindxedge/base/common"
)

// CtlComponent is a struct used to do a specific command to a single component
type CtlComponent struct {
	Name           string
	Operation      string
	InstallPathMgr *InstallDirPathMgr
}

func (cc *CtlComponent) startComponent(yamlPath string) error {
	if err := cc.prepareNameSpace(); err != nil {
		return err
	}

	if _, err := common.RunCommand(CommandKubectl, true, "apply", "-f", yamlPath); err != nil {
		hwlog.RunLog.Errorf("exec kubectl apply failed: %s", err.Error())
		return fmt.Errorf("exec kubectl apply failed: %s", err.Error())
	}

	if err := cc.checkIfComponentReady(); err != nil {
		return err
	}
	return nil
}

func (cc *CtlComponent) stopComponent(yamlPath string) error {
	ret, err := cc.checkNameSpaceExist()
	if err != nil {
		return err
	}
	if !ret {
		hwlog.RunLog.Warnf("namespace %s does not exist, no component could start", MefNamespace)
		return nil
	}

	status, err := cc.getComponentStatus()
	if err != nil {
		return err
	}
	if status == "" {
		hwlog.RunLog.Infof("component %s does not start now, no need to stop", cc.Name)
		return nil
	}

	if _, err = common.RunCommand(CommandKubectl, true,
		"scale", "deployment", cc.Name, "-n", MefNamespace, "--replicas=0"); err != nil {
		hwlog.RunLog.Errorf("exec kubectl delete failed: %s", err.Error())
		return fmt.Errorf("exec kubectl delete failed: %s", err.Error())
	}
	return nil
}

func (cc *CtlComponent) checkNameSpaceExist() (bool, error) {
	checkCmd := fmt.Sprintf("%s get namespaces | grep -w %s", CommandKubectl, MefNamespace)
	ret, err := common.RunCommand("sh", false, "-c", checkCmd)
	if err != nil {
		hwlog.RunLog.Errorf("check namespace command exec failed: %s", err.Error())
		return false, errors.New("check namespace command exec failed")
	}

	if ret == "" {
		return false, nil
	}

	return true, nil
}

func (cc *CtlComponent) getComponentStatus() (string, error) {
	checkCmd := fmt.Sprintf("%s get deployment -n %s | grep -w %s",
		CommandKubectl, MefNamespace, cc.Name)
	ret, err := common.RunCommand("sh", false, "-c", checkCmd)

	if err != nil && err.Error() != "" {
		hwlog.RunLog.Warnf("check components %s's status failed: %s", cc.Name, err.Error())
		return "", errors.New("check components status failed")
	}

	return ret, nil
}

func (cc *CtlComponent) checkIfStatusReady(status string) bool {
	if !strings.Contains(status, ReadyFlag) {
		hwlog.RunLog.Warn("the component pod is not active yet")
		return false
	}
	return true
}

func (cc *CtlComponent) checkIfComponentReady() error {
	for i := 0; i < CheckStatusTimes; i++ {
		status, err := cc.getComponentStatus()
		if err != nil {
			time.Sleep(CheckStatusInterval)
			continue
		}
		if cc.checkIfStatusReady(status) {
			return nil
		}
		time.Sleep(CheckStatusInterval)
	}
	hwlog.RunLog.Errorf("componentFlag [%s] is not running", cc.Name)
	return fmt.Errorf("componentFlag [%s] is not running", cc.Name)
}

func (cc *CtlComponent) prepareNameSpace() error {
	hwlog.RunLog.Info("start to prepare namespace")
	checkCmd := fmt.Sprintf("%s get namespaces | grep -w %s | awk '{print$2}'", CommandKubectl, MefNamespace)
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
		_, err = common.RunCommand(CommandKubectl, true,
			"delete", CommandNamespace, MefNamespace)
		if err != nil {
			hwlog.RunLog.Errorf("the namespace exists but not active, delete it failed: %s", err.Error())
			return errors.New("the namespace exists but not active, delete it failed")
		}
	}

	// namespace does not exist, then create
	hwlog.RunLog.Info("start to create namespace")
	_, err = common.RunCommand(CommandKubectl, true, "create", CommandNamespace, MefNamespace)
	if err != nil {
		hwlog.RunLog.Errorf("create namespace failed: %s", err.Error())
		return fmt.Errorf("create namespace failed")
	}

	hwlog.RunLog.Info("prepare namespace successful")
	return nil
}

// Operate is used to start an operate to a single component
func (cc *CtlComponent) Operate() error {
	hwlog.RunLog.Infof("start to %s module %s", cc.Operation, cc.Name)
	yamlPath := cc.InstallPathMgr.WorkPathMgr.GetRelativeYamlPath(cc.Name)
	yamlRealPath, err := filepath.EvalSymlinks(yamlPath)
	if err != nil {
		hwlog.RunLog.Errorf("get real path of component [%s]'s yaml failed: %s", cc.Name, err.Error())
		return fmt.Errorf("get real path of component [%s]'s yaml failed", cc.Name)
	}

	filePath, err := utils.CheckPath(yamlRealPath)
	if err != nil {
		hwlog.RunLog.Errorf("check real path of component [%s]'s yaml failed: %s", cc.Name, err.Error())
		return fmt.Errorf("check real path of component [%s]'s yaml failed", cc.Name)
	}

	switch cc.Operation {
	case "start":
		if err = cc.startComponent(filePath); err != nil {
			return err
		}

	case "stop":
		if err = cc.stopComponent(filePath); err != nil {
			return err
		}

	case "restart":
		if err = cc.stopComponent(filePath); err != nil {
			return err
		}
		if err = cc.startComponent(filePath); err != nil {
			return err
		}
	default:
		hwlog.RunLog.Errorf("unsupported Operate type")
		return errors.New("unsupported Operate type")
	}
	hwlog.RunLog.Infof("%s module %s successful", cc.Operation, cc.Name)
	return nil
}
