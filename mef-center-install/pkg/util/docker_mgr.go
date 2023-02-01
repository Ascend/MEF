// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path"
	"strconv"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

// DockerDealer is a struct to handle docker images
type DockerDealer struct {
	imageName string
	tag       string
}

// GetDockerDealer inits a DockerDealer struct
func GetDockerDealer(componentName string, tag string) DockerDealer {
	return DockerDealer{
		imageName: ImagePrefix + componentName,
		tag:       tag,
	}
}

// LoadImage is used to build a docker image by Dockerfile
func (dd *DockerDealer) LoadImage(buildPath string) error {
	absPath, err := utils.CheckPath(buildPath)
	if err != nil {
		return err
	}
	// imageName is fixed name.
	// tag is read from file or filename, the verification will be added in setVersion.
	// absPath has been verified
	cmdStr := "docker build -t " + dd.imageName + ":" + dd.tag + " " + absPath + "/."
	if _, err = common.RunCommand("sh", false, "-c", cmdStr); err != nil {
		hwlog.RunLog.Errorf("load docker image [%s] failed:%s", dd.imageName, err.Error())
		return errors.New("load docker image failed")
	}

	return nil
}

// SaveImage is used to save docker image to a specific path
func (dd *DockerDealer) SaveImage(savePath string) error {
	imageTarName, err := dd.getImageTarName()
	if err != nil {
		return err
	}

	savePath = path.Join(savePath, imageTarName)
	cmdStr := fmt.Sprintf("docker save %s:%s > %s", dd.imageName, dd.tag, savePath)
	if _, err = common.RunCommand("sh", false, "-c", cmdStr); err != nil {
		hwlog.RunLog.Errorf("save docker image [%s:%s] failed:%s", dd.imageName, dd.tag, err.Error())
		return errors.New("save docker image failed")
	}

	return nil

}

func (dd *DockerDealer) getImageTarName() (string, error) {
	arch, err := GetArch()
	if err != nil {
		hwlog.RunLog.Errorf("get system arch failed: %s", err.Error())
		return "", errors.New("get system arch failed")
	}

	return fmt.Sprintf(ImageTarNamePattern, dd.imageName, arch), nil
}

func (dd *DockerDealer) checkImageExist() (bool, error) {
	checkCmd := fmt.Sprintf("docker image ls %s | wc -l", dd.imageName+":"+dd.tag)
	ret, err := common.RunCommand("sh", false, "-c", checkCmd)
	if err != nil {
		hwlog.RunLog.Errorf("check %s's docker image command exec failed: %s", dd.imageName, err.Error())
		return false, fmt.Errorf("check %s's docker image command exec failed", dd.imageName)
	}

	if ret == strconv.Itoa(DockerImageExist) {
		return true, nil
	}

	return false, nil
}

// DeleteImage is used to delete the docker images
func (dd *DockerDealer) DeleteImage() error {
	ret, err := dd.checkImageExist()
	if err != nil {
		return err
	}
	if !ret {
		hwlog.RunLog.Warnf(" %s's docker image does not exist, no need to delete", dd.imageName)
		return nil
	}

	_, err = common.RunCommand("docker", true, "rmi", dd.imageName+":"+dd.tag)
	if err != nil {
		hwlog.RunLog.Errorf("delete %s's docker image command exec failed: %s", dd.imageName, err.Error())
		return fmt.Errorf("delete %s's docker image command exec failed", dd.imageName)
	}

	return nil
}
