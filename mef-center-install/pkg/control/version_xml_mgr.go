// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package control

import (
	"encoding/xml"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
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
	xmlIns := new(VersionXmlTemplate)
	content, err := utils.LoadFile(xmlPath)
	if err != nil {
		hwlog.RunLog.Errorf("read version.xml failed: %s", err.Error())
		return nil, errors.New("read version.xml failed")
	}

	err = xml.Unmarshal(content, xmlIns)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal version.xml failed: %s", err.Error())
		return nil, errors.New("unmarshal version.xml failed")
	}

	return xmlIns, nil
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

	if len(xmlIns.Package) == 0 {
		hwlog.RunLog.Error("unmarshal version.xml failed: cannot find package data")
		return "", errors.New("unmarshal version.xml failed")
	}

	return xmlIns.Package[0].InnerVersion, nil
}
