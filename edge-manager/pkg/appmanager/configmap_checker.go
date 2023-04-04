// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager checker related configmap
package appmanager

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"edge-manager/pkg/util"
)

const configMapContentMaxCount = 64
const configmapContentValueMaxLen = 1024

type configmapParaChecker struct {
	req *ConfigmapReq
}

type configmapParaPattern struct {
	patterns map[string]string
}

var configmapPattern = configmapParaPattern{patterns: map[string]string{
	configmapName:        "^[a-zA-Z0-9][a-zA-Z0-9-_]{0,61}[a-zA-Z0-9]$",
	configmapDescription: `^[\S ]{0,255}$`,
	configmapContentKey:  "^[a-zA-Z-]([a-zA-Z0-9-_.]){0,62}$",
},
}

func (cpp *configmapParaPattern) getPatternFromMap(mapKey string) (string, bool) {
	pattern, ok := cpp.patterns[mapKey]
	return pattern, ok
}

func (cpc *configmapParaChecker) Check() error {
	var checkItems = []func() error{
		cpc.checkConfigmapNameValid,
		cpc.checkConfigmapDescriptionValid,
		cpc.checkConfigmapContentValid,
	}

	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}

	return nil
}

func (cpc *configmapParaChecker) checkConfigmapNameValid() error {
	configmapNamePattern, ok := configmapPattern.getPatternFromMap(configmapName)
	if !ok {
		hwlog.RunLog.Error("configmapName regex is not exist")
		return errors.New("configmapName regex is not exist")
	}

	if !util.RegexStringChecker(cpc.req.ConfigmapName, configmapNamePattern) {
		hwlog.RunLog.Error("configmap name doesn't match regex")
		return errors.New("configmap name doesn't match regex")
	}

	hwlog.RunLog.Info("check configmap name success")
	return nil
}

func (cpc *configmapParaChecker) checkConfigmapDescriptionValid() error {
	configmapDescriptionPattern, ok := configmapPattern.getPatternFromMap(configmapDescription)
	if !ok {
		hwlog.RunLog.Error("configmapDescription regex is not exist")
		return errors.New("configmapDescription regex is not exist")
	}

	if !util.RegexStringChecker(cpc.req.Description, configmapDescriptionPattern) {
		hwlog.RunLog.Error("configmap description doesn't match regex")
		return errors.New("configmap description doesn't match regex")
	}

	hwlog.RunLog.Info("check configmap description success")
	return nil
}

func (cpc *configmapParaChecker) checkConfigmapContentValid() error {
	if len(cpc.req.ConfigmapContent) > configMapContentMaxCount {
		hwlog.RunLog.Error("configmap content key-value count is invalid")
		return errors.New("configmap content key-value count  is invalid")
	}

	var configmapContentKeys []string
	for idx := range cpc.req.ConfigmapContent {
		configmapContentKeys = append(configmapContentKeys, cpc.req.ConfigmapContent[idx].Name)
		var checker = configmapContentParaChecker{
			configmapContent: &cpc.req.ConfigmapContent[idx],
		}
		if err := checker.check(); err != nil {
			return err
		}
	}

	if err := checkConfigmapContentKeyUniqueValid(configmapContentKeys); err != nil {
		return err
	}

	hwlog.RunLog.Info("check configmap content success")
	return nil
}

type configmapContentParaChecker struct {
	configmapContent *ConfigmapContent
}

func (ccpc *configmapContentParaChecker) check() error {
	var checkItems = []func() error{
		ccpc.checkConfigmapContentKeyValid,
		ccpc.checkConfigmapContentValueValid,
	}

	for _, checkItem := range checkItems {
		if err := checkItem(); err != nil {
			return err
		}
	}

	return nil
}

func (ccpc *configmapContentParaChecker) checkConfigmapContentKeyValid() error {
	configmapContentKeyPattern, ok := configmapPattern.getPatternFromMap(configmapContentKey)
	if !ok {
		hwlog.RunLog.Error("configmapContentKey regex is not exist")
		return errors.New("configmapContentKey regex is not exist")
	}

	if !util.RegexStringChecker(ccpc.configmapContent.Name, configmapContentKeyPattern) {
		hwlog.RunLog.Errorf("configmap content key [%s] doesn't match regex", ccpc.configmapContent.Name)
		return fmt.Errorf("configmap content key [%s] doesn't match regex", ccpc.configmapContent.Name)
	}

	hwlog.RunLog.Infof("check configmap content key [%s] success", ccpc.configmapContent.Name)
	return nil
}

func (ccpc *configmapContentParaChecker) checkConfigmapContentValueValid() error {
	if len(ccpc.configmapContent.Value) > configmapContentValueMaxLen {
		hwlog.RunLog.Errorf("configmap content value [%s] length is invalid", ccpc.configmapContent.Value)
		return fmt.Errorf("configmap content value [%s] length is invalid", ccpc.configmapContent.Value)
	}

	hwlog.RunLog.Infof("check configmap content value [%s] success", ccpc.configmapContent.Value)
	return nil
}

func checkConfigmapContentKeyUniqueValid(configmapContentKeys []string) error {
	cmContentKeysMap := make(map[string]int)
	for _, cmContentKey := range configmapContentKeys {
		cmContentKeysMap[cmContentKey] = 1
	}

	var cmContentKeysAfterDeduplicated []interface{}
	for cmContentKeyAfterDeduplicated := range cmContentKeysMap {
		cmContentKeysAfterDeduplicated = append(cmContentKeysAfterDeduplicated, cmContentKeyAfterDeduplicated)
	}

	if len(cmContentKeysAfterDeduplicated) != len(configmapContentKeys) {
		hwlog.RunLog.Error("configmap content key is duplicated")
		return errors.New("configmap content key is duplicated")
	}

	hwlog.RunLog.Info("check configmap content key unique success")
	return nil
}

func checkItemCountInDB() error {
	total, err := database.GetItemCount(ConfigmapInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get table configmap_infos num failed, error: %v", err)
		return fmt.Errorf("get table configmap_infos num failed, error: %s", err.Error())
	}

	if total >= maxConfigmapItemNum {
		hwlog.RunLog.Error("table configmap_infos item num is enough, can't be created")
		return errors.New("table configmap_infos item num is enough, can't be created")
	}

	hwlog.RunLog.Info("check item count in database success")
	return nil
}
