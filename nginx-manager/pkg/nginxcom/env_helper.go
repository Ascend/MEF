// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package nginxcom this file is for common constant or method
package nginxcom

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// NginxConfItem nginx replace item info
type NginxConfItem struct {
	Key  string
	From string
	To   string
}

// confItemsTemplate the items needed to replace into nginx.conf
var confItemsTemplate = []NginxConfItem{
	{Key: EdgePortKey, From: KeyPrefix + EdgePortKey},
	{Key: CertPortKey, From: KeyPrefix + CertPortKey},
	{Key: AuthPortKey, From: KeyPrefix + AuthPortKey},
	{Key: WebsocketPortKey, From: KeyPrefix + WebsocketPortKey},
	{Key: NginxSslPortKey, From: KeyPrefix + NginxSslPortKey},
	{Key: PodIpKey, From: KeyPrefix + PodIpKey},
}

// GetConfigItemTemplate get the template of config replace items
func GetConfigItemTemplate() []NginxConfItem {
	return confItemsTemplate
}

// environmentMgr manager environment var
type environmentMgr struct {
	valuers map[string]*environmentValuer
}

type environmentValuer struct {
	Key          string
	Value        string
	DefaultValue string
	Require      bool
	Checker      ObjChecker
}

var envMgr = newEnvironmentMgr()

// GetEnvManager env manager is for check, load, get environment vars
func GetEnvManager() *environmentMgr {
	return envMgr
}

func newEnvironmentMgr() *environmentMgr {
	valuers := map[string]*environmentValuer{
		EdgePortKey: {EdgePortKey, "", "", true,
			createIntChecker(common.MinPort, common.MaxPort)},
		CertPortKey: {CertPortKey, "", "", true,
			createIntChecker(common.MinPort, common.MaxPort)},
		AuthPortKey: {AuthPortKey, "", "", true,
			createIntChecker(common.MinPort, common.MaxPort)},
		WebsocketPortKey: {WebsocketPortKey, "", "", true,
			createIntChecker(common.MinPort, common.MaxPort)},
		NginxSslPortKey: {NginxSslPortKey, "", "", true,
			createIntChecker(common.MinPort, common.MaxPort)},
		PodIpKey: {PodIpKey, "", "", true,
			createIpChecker()},
		EnableResolverKey: {EnableResolverKey, "", "false", false,
			createBoolChoiceChecker()},
	}
	return &environmentMgr{valuers: valuers}
}

func createIntChecker(min int64, max int64) ObjChecker {
	return ObjChecker{
		Checker:  checker.GetAndChecker(checker.GetIntChecker("", min, max, true)),
		DataType: reflect.Int,
	}
}

func createBoolChoiceChecker() ObjChecker {
	choices := []string{"true", "false"}
	return ObjChecker{
		Checker:  checker.GetAndChecker(checker.GetStringChoiceChecker("", choices, true)),
		DataType: reflect.String,
	}
}

func createIpChecker() ObjChecker {
	return ObjChecker{
		Checker:  checker.GetAndChecker(checker.GetIpV4Checker("", true)),
		DataType: reflect.String,
	}
}

// Load load method to load all environments needed
func (m *environmentMgr) Load() error {
	for _, v := range m.valuers {
		v.load()
		if !v.check() {
			hwlog.RunLog.Errorf("load env error, key: %s, %s, %s", v.Key, v.Value, v.DefaultValue)
			return fmt.Errorf("load env error, key: %s, val: %s, default: %s", v.Key, v.Value, v.DefaultValue)
		}
	}
	return nil
}

// Get the string value of the environment specified by the key
func (m *environmentMgr) Get(key string) (string, error) {
	valuer, ok := m.valuers[key]
	if !ok {
		return "", fmt.Errorf("no valuer for key: %s", key)
	}
	return valuer.get()
}

// GetInt get the int value of the environment specified by the key
func (m *environmentMgr) GetInt(key string) (int, error) {
	val, err := m.Get(key)
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

func (e *environmentValuer) check() bool {
	if e.Require && e.Value == "" {
		hwlog.RunLog.Errorf("env %s is required, but value not found", e.Key)
		return false
	}
	if !e.Require && e.DefaultValue == "" {
		hwlog.RunLog.Errorf("env %s not required, but defaultValue not found", e.Key)
		return false
	}
	if e.Require {
		return e.Checker.Check(e.Value).Result
	}
	if e.Value == "" {
		return e.Checker.Check(e.DefaultValue).Result
	} else {
		return e.Checker.Check(e.Value).Result
	}
}

func (e *environmentValuer) load() {
	e.Value = os.Getenv(e.Key)
}

func (e *environmentValuer) get() (string, error) {
	if e.Value != "" {
		return e.Value, nil
	}
	if e.DefaultValue != "" {
		return e.DefaultValue, nil
	}
	return "", fmt.Errorf("cannot get string value or defaultValue for key: %s", e.Key)
}
