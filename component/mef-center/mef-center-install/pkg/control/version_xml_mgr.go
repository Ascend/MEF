// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package control

import (
	"encoding/xml"
	"errors"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
)

// VersionXmlTemplate is the struct to deal with version.xml
type VersionXmlTemplate struct {
	XmlName    xml.Name  `xml:"SoftwarePackage"`
	PkgVersion string    `xml:"version,attr"`
	Package    []Package `xml:"Package"`
}

// Package is the struct to deal with the Package session in version.xml
type Package struct {
	XmlName       xml.Name `xml:"Package"`
	FileName      string   `xml:"FileName"`
	OuterName     string   `xml:"OuterName"`
	Version       string   `xml:"Version"`
	InnerVersion  string   `xml:"InnerVersion"`
	FileType      string   `xml:"FileType"`
	Vendor        string   `xml:"Vendor"`
	SupportModel  string   `xml:"SupportModel"`
	ProcessorArch string   `xml:"ProcessorArchitecture"`
}

func unmarshalXml(xmlPath string) (*VersionXmlTemplate, error) {
	content, err := fileutils.LoadFile(xmlPath)
	if err != nil {
		hwlog.RunLog.Errorf("read version.xml failed: %s", err.Error())
		return nil, errors.New("read version.xml failed")
	}

	var xmlIns VersionXmlTemplate
	if err = xml.Unmarshal(content, &xmlIns); err != nil {
		hwlog.RunLog.Errorf("unmarshal version.xml failed: %s", err.Error())
		return nil, errors.New("unmarshal version.xml failed")
	}

	return &xmlIns, nil
}

// GetVersion func is used to get the inner version in version.xml
func GetVersion(xmlPath string) (string, error) {
	xmlIns, err := unmarshalXml(xmlPath)
	if err != nil {
		return "", err
	}
	if xmlIns == nil {
		return "", errors.New("marshaled version.xml is nil")
	}

	if len(xmlIns.Package) < 1 {
		hwlog.RunLog.Error("unmarshal version.xml failed: cannot find package data")
		return "", errors.New("unmarshal version.xml failed")
	}

	return xmlIns.Package[0].InnerVersion, nil
}
