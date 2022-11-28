// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strings"

	"huawei.com/mindx/common/hwlog"
)

const edgeCore = "edgecore"
const edgeInstaller = "edgeinstaller"

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkFields(contentType string, version string) bool {
	if contentType != edgeCore && contentType != edgeInstaller {
		return false
	}
	match, err := regexp.MatchString("^v+.+.+$", version)
	if err != nil {
		hwlog.RunLog.Error("Regexp match error")
		return false
	}
	return match
}

func checkFile(file *multipart.FileHeader) bool {
	strslice := strings.Split(file.Filename, ".")
	if strslice[len(strslice)-1] != "zip" {
		return false
	}
	if float64(file.Size)/kbToMB > maximumSize {
		return false
	}
	return true
}

func checkSoftwareExist(contentType string, version string) string {
	var softwareRecords []softwareRecord
	gormDB.Where("content_type=? and version=?", contentType, version).Find(&softwareRecords)
	if len(softwareRecords) == 0 || len(softwareRecords) != 1 {
		return ""
	}
	dst := RepositoryFilesPath + "/" + contentType + "/" + contentType + "_" + version
	return dst + "/" + contentType + ".zip"

}

func creatDir(contentType string, version string) string {
	dst := RepositoryFilesPath + "/" + contentType + "/" + contentType + "_" + version
	b, err := pathExists(dst)
	if err != nil {
		hwlog.RunLog.Error("Path checking error")
		return ""
	}
	if !b {
		err := os.MkdirAll(dst, os.ModePerm)
		if err != nil {
			hwlog.RunLog.Error("Create directory error")
			return ""
		}
	}
	return dst
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, dBFileMode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}
