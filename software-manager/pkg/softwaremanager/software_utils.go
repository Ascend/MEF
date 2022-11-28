// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"fmt"
	"huawei.com/mindx/common/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"huawei.com/mindx/common/hwlog"
)

const edgeCore = "edgecore"
const edgeInstaller = "edgeinstaller"

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
	dst := filepath.Join(RepositoryFilesPath, contentType, fmt.Sprintf("%s%s%s", contentType, "_", version))
	return filepath.Join(dst, fmt.Sprintf("%s%s", contentType, ".zip"))
}

func creatDir(contentType string, version string) string {
	dst := filepath.Join(RepositoryFilesPath, contentType, contentType+"_"+version)
	b := utils.IsExist(dst)
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
