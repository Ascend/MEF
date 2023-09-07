// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

// CtlComponent is a struct used to do a specific command to a single component
type CtlComponent struct {
	Name           string
	Operation      string
	InstallPathMgr WorkPathItf
}

func (cc *CtlComponent) startComponent(yamlPath string) (bool, error) {
	nsMgr := NewNamespaceMgr(MefNamespace)
	if err := nsMgr.prepareNameSpace(); err != nil {
		return false, err
	}

	status, err := cc.getComponentStatus()
	if cc.checkIfStatusReady(status) && err == nil {
		return true, nil
	}

	if err = backuputils.InitConfig(yamlPath, cc.tryStartComponent); err != nil {
		hwlog.RunLog.Errorf("exec kubectl apply failed: %s", err.Error())
		return false, fmt.Errorf("exec kubectl apply failed: %s", err.Error())
	}

	if err = cc.checkIfComponentReady(); err != nil {
		return false, err
	}
	return false, nil
}

func (cc *CtlComponent) tryStartComponent(yamlPath string) error {
	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "apply", "-f",
		yamlPath); err != nil {
		return err
	}
	return nil
}

func (cc *CtlComponent) stopComponent() error {
	nsMgr := NewNamespaceMgr(MefNamespace)
	ret, err := nsMgr.checkNameSpaceExist()
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
		hwlog.RunLog.Warnf("component %s does not start now, no need to stop", cc.Name)
		return nil
	}

	if err = cc.setReplicas(StopReplicasNum); err != nil {
		return err
	}
	return nil
}

func (cc *CtlComponent) getComponentStatus() (string, error) {
	ret, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "get", "deployment", "-n", MefNamespace)

	NoNamespaceErr := fmt.Sprintf("No resources found in %s namespace.", MefNamespace)
	if err != nil && strings.Contains(err.Error(), NoNamespaceErr) {
		return ret, err
	}
	if err != nil {
		hwlog.RunLog.Warnf("check components %s's status failed: %s", cc.Name, err.Error())
		return "", errors.New("check components status failed")
	}

	deploymentReg := fmt.Sprintf("^%s\\s", AscendPrefix+cc.Name)
	lines := strings.Split(ret, "\n")
	for _, line := range lines {
		found, err := regexp.MatchString(deploymentReg, line)
		if err != nil {
			hwlog.RunLog.Errorf("check components %s's status on reg match failed: %s", cc.Name, err.Error())
			return "", errors.New("check components status failed")
		}
		if found {
			return line, nil
		}
	}

	return "", nil
}

func (cc *CtlComponent) checkIfStatusStopped(status string) bool {
	if !strings.Contains(status, StopFlag) {
		hwlog.RunLog.Warn("the component pod is not active yet")
		return false
	}
	return true
}

func (cc *CtlComponent) setReplicas(num int) error {
	deploymentName := AscendPrefix + cc.Name
	if _, err := envutils.RunCommand(CommandKubectl, envutils.DefCmdTimeoutSec, "scale", "deployment", deploymentName,
		"-n", MefNamespace, fmt.Sprintf("--replicas=%d", num)); err != nil {
		hwlog.RunLog.Errorf("exec kubectl delete failed: %s", err.Error())
		return fmt.Errorf("exec kubectl delete failed: %s", err.Error())
	}

	return nil
}

func (cc *CtlComponent) checkIfStatusReady(status string) bool {
	if !strings.Contains(status, ReadyFlag) {
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
		if cc.checkIfStatusStopped(status) {
			if err = cc.setReplicas(StartReplicasNum); err != nil {
				hwlog.RunLog.Warn("the component pod is stop, start it now")
				time.Sleep(CheckStatusInterval)
				continue
			}
		}
		if cc.checkIfStatusReady(status) {
			return nil
		}
		time.Sleep(CheckStatusInterval)
	}
	hwlog.RunLog.Errorf("componentFlag [%s] is not running", cc.Name)
	return fmt.Errorf("componentFlag [%s] is not running", cc.Name)
}

// CheckStarted is used to check if a single component starts
func (cc *CtlComponent) CheckStarted() (bool, error) {
	status, err := cc.getComponentStatus()
	if err != nil {
		hwlog.RunLog.Errorf("get component %s's status failed: %s", cc.Name, err.Error())
		return false, fmt.Errorf("get component %s's status failed", cc.Name)
	}
	if cc.checkIfStatusReady(status) {
		return true, nil
	}
	return false, nil
}

// Operate is used to start an operate to a single component
func (cc *CtlComponent) Operate() error {
	hwlog.RunLog.Infof("start to %s module %s", cc.Operation, cc.Name)
	fmt.Printf("start to %s module %s\n", cc.Operation, cc.Name)
	if cc.InstallPathMgr == nil {
		hwlog.RunLog.Error("pointer WorkPathItf is nil or invalid")
		return errors.New("pointer WorkPathItf is nil or invalid")
	}

	yamlPath := cc.InstallPathMgr.GetComponentYamlPath(cc.Name)
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
	case StartOperateFlag:
		started, err := cc.startComponent(filePath)
		if err != nil {
			return err
		}

		if started {
			fmt.Printf("module %s's status unchanged\n", cc.Name)
			hwlog.RunLog.Infof("%s module %s, the component's status unchanged", cc.Operation, cc.Name)
			return nil
		}

	case StopOperateFlag:
		if err = cc.stopComponent(); err != nil {
			return err
		}

	case RestartOperateFlag:
		if err = cc.stopComponent(); err != nil {
			return err
		}
		if _, err = cc.startComponent(filePath); err != nil {
			return err
		}
	default:
		hwlog.RunLog.Errorf("unsupported Operate type")
		return errors.New("unsupported Operate type")
	}
	fmt.Printf("%s module %s successful\n", cc.Operation, cc.Name)
	hwlog.RunLog.Infof("%s module %s successful", cc.Operation, cc.Name)
	return nil
}
