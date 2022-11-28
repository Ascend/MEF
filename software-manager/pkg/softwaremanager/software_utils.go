// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

package softwaremanager

import (
	"fmt"
	"huawei.com/mindx/common/utils"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"software-manager/pkg/restfulservice"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
)

const (
	edgeCore      = "edgecore"
	edgeInstaller = "edgeinstaller"
	userLength    = 8
	psdLength     = 16
	maxLength     = 30
)

const (
	numSet    = "0123456789"
	charSet   = "abcdefghijklkmnopqrstuvwxyzABCDEFGHIJKLKMNOPQRSTUVWXYZ"
	symbolSet = "!@#$%^&*()-."
)

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

func usrgenerate() []byte {
	return generateRandomString(fmt.Sprintf("%s%s", numSet, charSet), userLength)
}

func psdgenerate() []byte {
	return generateRandomString(fmt.Sprintf("%s%s%s", numSet, charSet, symbolSet), psdLength)
}

func downloadRight(userName []byte, password []byte, nodeID string) bool {
	userInfo := restfulservice.UserInfoMap[nodeID]
	if userInfo == nil {
		return false
	}
	return checkByteArr(userName, userInfo[restfulservice.UserName]) &&
		checkByteArr(password, userInfo[restfulservice.Password])
}

func generateRandomString(source string, length int) []byte {
	if length < maxLength {
		rand.Seed(time.Now().UnixNano())
		randomByteArr := make([]byte, length, length)
		for i := 0; i < length; i++ {
			index := rand.Intn(len(source))
			randomByteArr[i] = source[index]
		}
		return randomByteArr
	}
	hwlog.RunLog.Error("length is too long")
	return nil
}

func checkNodeID(nodeID string) bool {
	return nodeID != ""
}

func checkByteArr(arr1 []byte, arr2 []byte) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for index, val := range arr1 {
		if val != arr2[index] {
			return false
		}
	}
	return true
}
