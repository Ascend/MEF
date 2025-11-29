// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package config this file for version xml file manager
package config

import (
	"errors"
	"regexp"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

const matchedLen = 2

// VersionXmlMgr version xml file manager
type VersionXmlMgr struct {
	versionPath string
}

// NewVersionXmlMgr create a version xml file manager
func NewVersionXmlMgr(versionPath string) *VersionXmlMgr {
	return &VersionXmlMgr{
		versionPath: versionPath,
	}
}

// GetSftPkgName get the value of package name in version xml file
func (vxm VersionXmlMgr) GetSftPkgName() (string, error) {
	return vxm.getFieldValue("OutterName")
}

// GetInnerVersion get the value of InnerVersion in version xml file
func (vxm VersionXmlMgr) GetInnerVersion() (string, error) {
	return vxm.getFieldValue("InnerVersion")
}

// GetVersion get the value of Version in version xml file
func (vxm VersionXmlMgr) GetVersion() (string, error) {
	return vxm.getFieldValue("Version")
}

func (vxm VersionXmlMgr) checkAndGetXmlData() ([]byte, error) {
	if _, err := fileutils.RealFileCheck(vxm.versionPath, false, false, constants.MaxXmlSizeTimes); err != nil {
		hwlog.RunLog.Errorf("check version xml path [%s] failed, error: %v", vxm.versionPath, err)
		return nil, err
	}

	xmlData, err := fileutils.LoadFile(vxm.versionPath)
	if err != nil {
		hwlog.RunLog.Errorf("load version xml file [%s] failed, error: %v", vxm.versionPath, err)
		return nil, err
	}

	return xmlData, nil
}

func (vxm VersionXmlMgr) getRegexStr(fieldName string) (string, error) {
	fieldRegexMap := map[string]string{
		"InnerVersion": "<InnerVersion>([0-9.]{1,16})</InnerVersion>",
		"Version":      "<Version>([A-Za-z0-9.]{1,16})</Version>",
		"OutterName":   "<OutterName>([A-Za-z_]{1,16})</OutterName>",
	}
	regexStr, ok := fieldRegexMap[fieldName]
	if !ok {
		hwlog.RunLog.Errorf("get regex string for field name [%s] failed", fieldName)
		return "", errors.New("get regex string failed")
	}
	return regexStr, nil
}

func (vxm VersionXmlMgr) getFieldValue(fieldName string) (string, error) {
	xmlData, err := vxm.checkAndGetXmlData()
	if err != nil {
		return "", err
	}

	regexStr, err := vxm.getRegexStr(fieldName)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(regexStr)
	if err != nil {
		hwlog.RunLog.Errorf("invalid regular expression, error: %v", err)
		return "", err
	}

	result := re.FindSubmatch(xmlData)
	if len(result) != matchedLen {
		hwlog.RunLog.Errorf("the field name [%s] finds value from version xml data failed", fieldName)
		return "", errors.New("find value from version xml data failed")
	}
	return string(result[1]), nil
}
