// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgchecker for Secret
package msgchecker

import (
	"errors"
	"fmt"

	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/database"
)

const (
	fdConfigMapDataKeyFormat = "^[a-zA-Z-]([a-zA-Z0-9-_.]){0,62}$"
	maxAllowedCmNumber       = 256
)

func (mv *MsgValidator) auxCheckCm(cm *types.ConfigMap) error {
	if err := validateStruct(cm); err != nil {
		return err
	}
	checkFuncs := []func(c *types.ConfigMap) error{
		checkCmDataKey,
		checkCmNumber,
	}

	for _, check := range checkFuncs {
		if err := check(cm); err != nil {
			return err
		}

	}

	return nil
}
func checkCmDataKey(c *types.ConfigMap) error {
	for key := range c.Data {
		if !checker.RegexStringChecker(key, fdConfigMapDataKeyFormat) {
			return errors.New("configmap data key check failed")
		}
	}
	return nil
}
func checkCmNumber(c *types.ConfigMap) error {
	// allow using config's update message
	if isUsingConfigmap(c.Name) {
		return nil
	}

	existingCount, err := database.GetMetaRepository().CountByType(constants.ResourceTypeConfigMap)
	if err != nil {
		return fmt.Errorf("get existing configmap count failed, %v", err)
	}

	if existingCount+1 > maxAllowedCmNumber {
		return fmt.Errorf("out of allowed max configmap limit")
	}
	return nil
}

func isUsingConfigmap(name string) bool {
	key := fmt.Sprintf("websocket/configmap/%s", name)
	_, err := database.GetMetaRepository().GetByKey(key)
	if err != nil {
		return false
	}
	return true
}
