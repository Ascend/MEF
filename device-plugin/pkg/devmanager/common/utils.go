// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common this for util method
package common

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// IsGreaterThanOrEqualInt32 check num range
func IsGreaterThanOrEqualInt32(num int64) bool {
	if num >= int64(math.MaxInt32) {
		return true
	}

	return false
}

// IsValidChipInfo valid chip info is or not empty
func IsValidChipInfo(chip *ChipInfo) bool {
	return chip.Name != "" || chip.Type != "" || chip.Version != ""
}

// IsValidCardID valid card id
func IsValidCardID(cardID int32) bool {
	// for cardID, please watch the maximum value of the driver is changed in the future version
	return cardID >= 0 && cardID < HiAIMaxCardID
}

// IsValidDeviceID valid device id
func IsValidDeviceID(deviceID int32) bool {
	return deviceID >= 0 && deviceID < HiAIMaxDeviceNum
}

// IsValidLogicIDOrPhyID valid logic id
func IsValidLogicIDOrPhyID(id int32) bool {
	return id >= 0 && id < HiAIMaxCardNum*HiAIMaxDeviceNum
}

// IsValidCardIDAndDeviceID check two params both needs meet the requirement
func IsValidCardIDAndDeviceID(cardID, deviceID int32) bool {
	if !IsValidCardID(cardID) {
		return false
	}

	return IsValidDeviceID(deviceID)
}

// IsValidDevNumInCard valid devNum in card
func IsValidDevNumInCard(num int32) bool {
	return num > 0 && num <= HiAIMaxDeviceNum
}

// GetDeviceTypeByChipName get device type by chipName
func GetDeviceTypeByChipName(chipName string) string {
	if strings.Contains(chipName, "310P") {
		return Ascend310P
	}
	if strings.Contains(chipName, "310B") {
		return Ascend310B
	}
	if strings.Contains(chipName, "310") {
		return Ascend310
	}
	return ""
}

func get310PTemplateNameList() map[string]struct{} {
	return map[string]struct{}{"vir04": {}, "vir02": {}, "vir01": {}, "vir04_3c": {}, "vir02_1c": {},
		"vir04_4c_dvpp": {}, "vir04_3c_ndvpp": {}}
}

// IsValidTemplateName check template name meet the requirement
func IsValidTemplateName(devType, templateName string) bool {
	isTemplateNameValid := false
	switch devType {
	case Ascend310P:
		_, isTemplateNameValid = get310PTemplateNameList()[templateName]
	default:
	}
	return isTemplateNameValid
}

// RemoveDuplicate remove duplicate device
func RemoveDuplicate(list *[]string) []string {
	listValueMap := make(map[string]string, len(*list))
	var rmDupValueList []string
	for _, value := range *list {
		listValueMap[value] = value
	}
	for _, value := range listValueMap {
		rmDupValueList = append(rmDupValueList, value)
	}
	return rmDupValueList
}

// GetDriverLibPath get driver lib path from ld config
func GetDriverLibPath(libraryName string) (string, error) {
	var libPath string
	var err error
	if libPath, err = getLibFromEnv(libraryName); err == nil {
		return libPath, nil
	}
	if libPath, err = getLibFromLdCmd(libraryName); err == nil {
		return libPath, nil
	}
	return "", fmt.Errorf("cannot found valid driver lib, %#v", err)
}

func getLibFromEnv(libraryName string) (string, error) {
	ldLibraryPath := os.Getenv(ldLibPath)
	if len(ldLibraryPath) > maxPathLength {
		return "", fmt.Errorf("invalid library path env")
	}
	libraryPaths := strings.Split(ldLibraryPath, ":")
	return checkLibsPath(libraryPaths, libraryName)
}

func getLibFromLdCmd(libraryName string) (string, error) {
	libraryAbsName, err := parseLibFromLdCmd(libraryName)
	if err != nil {
		return "", err
	}
	if absLibPath, err := checkAbsPath(libraryAbsName); err == nil {
		return absLibPath, nil
	}
	return "", fmt.Errorf("driver lib is not exist or it's permission is invalid")
}

func parseLibFromLdCmd(libraryName string) (string, error) {
	ldCmd := exec.Command(ldCommand, ldParam)
	grepCmd := exec.Command(grepCommand, libraryName)
	ldCmdStdout, err := ldCmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("command exec failed")
	}
	grepCmd.Stdin = ldCmdStdout
	stdout, err := grepCmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("command exec failed")
	}
	if err := grepCmd.Start(); err != nil {
		return "", fmt.Errorf("command exec failed")
	}
	if err := ldCmd.Run(); err != nil {
		return "", fmt.Errorf("command exec failed")
	}
	defer func() {
		if err := grepCmd.Wait(); err != nil {
			log.Printf("command exec failed, %#v", err)
		}
	}()
	reader := bufio.NewReader(stdout)
	count := 0
	for {
		if count >= maxPathLength {
			break
		}
		count++
		line, err := reader.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		if libPath := parserLibPath(line, libraryName); libPath != "" {
			return libPath, nil
		}
	}
	return "", fmt.Errorf("can't find valid lib")
}

func parserLibPath(line, libraryName string) string {
	ldInfo := strings.Split(line, "=>")
	if len(ldInfo) < ldSplitLen {
		return ""
	}
	libNames := strings.Split(ldInfo[ldLibNameIndex], " ")
	for index, libName := range libNames {
		if index >= maxPathDepth {
			break
		}
		if len(libName) == 0 {
			continue
		}
		if name := trimSpaceTable(libName); name != libraryName {
			continue
		}
		return trimSpaceTable(ldInfo[ldLibPathIndex])
	}
	return ""
}

func checkLibsPath(libraryPaths []string, libraryName string) (string, error) {
	for _, libraryPath := range libraryPaths {
		libraryAbsName := path.Join(libraryPath, libraryName)
		if len(libraryAbsName) > maxPathLength {
			continue
		}
		if absLibPath, err := checkAbsPath(libraryAbsName); err == nil {
			return absLibPath, nil
		}
	}
	return "", fmt.Errorf("driver lib is not exist or it's permission is invalid")
}

func trimSpaceTable(data string) string {
	data = strings.Replace(data, " ", "", -1)
	data = strings.Replace(data, "\t", "", -1)
	data = strings.Replace(data, "\n", "", -1)
	return data
}

func checkAbsPath(libPath string) (string, error) {
	absLibPath, err := CheckOwnerAndPermission(libPath, DefaultWriteFileMode, rootUID)
	if err != nil {
		return "", err
	}
	count := 0
	fPath := absLibPath
	for {
		if count >= maxPathDepth {
			break
		}
		count++
		if fPath == "/" {
			return absLibPath, nil
		}
		fPath = filepath.Dir(fPath)
		if _, err := CheckOwnerAndPermission(fPath, DefaultWriteFileMode, rootUID); err != nil {
			return "", err
		}
	}
	return "", errors.New("absolute path check failed")
}

// CheckOwnerAndPermission check path  owner and permission
func CheckOwnerAndPermission(verifyPath string, mode os.FileMode, uid uint32) (string, error) {
	if verifyPath == "" {
		return verifyPath, errors.New("empty path")
	}
	absPath, err := filepath.Abs(verifyPath)
	if err != nil {
		return "", fmt.Errorf("abs failed %#v", err)
	}
	resoledPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("evalSymlinks failed %#v", err)
	}
	// if symlinks
	if absPath != resoledPath {
		// check symlinks its self owner
		pathInfo, err := os.Lstat(absPath)
		if err != nil {
			return "", fmt.Errorf("lstat failed, %#v", err)
		}
		stat, ok := pathInfo.Sys().(*syscall.Stat_t)
		if !ok || stat.Uid != uid {
			return "", errors.New("symlinks owner may not root")
		}
	}
	pathInfo, err := os.Stat(resoledPath)
	if err != nil {
		return "", fmt.Errorf("stat failed %#v", err)
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uid || !CheckMode(pathInfo.Mode(), mode) {
		return "", errors.New("check uid or mode failed")
	}
	return resoledPath, nil
}

// CheckMode check input file mode whether includes invalid mode.
// For example, if read operation of group and other is forbidden, then call CheckMode(inputFileMode, 0044).
// All operations are forbidden for group and other, then call CheckMode(inputFileMode, 0077).
// Write operation is forbidden for group and other by default, with calling CheckMode(inputFileMode)
func CheckMode(mode os.FileMode, optional ...os.FileMode) bool {
	var targetMode os.FileMode
	if len(optional) > 0 {
		targetMode = optional[0]
	} else {
		targetMode = DefaultWriteFileMode
	}
	checkMode := uint32(mode) & uint32(targetMode)
	return checkMode == 0
}
