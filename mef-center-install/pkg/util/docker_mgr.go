// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

// DockerDealer is a struct to handle docker images
type DockerDealer struct {
	imageName string
	version   string
}

// GetDockerDealer inits a DockerDealer struct
func GetDockerDealer(imageName string, version string) DockerDealer {
	return DockerDealer{
		imageName: imageName,
		version:   version,
	}
}

// LoadImage is used to build a docker image by Dockerfile
func (dd *DockerDealer) LoadImage(buildPath string) error {
	absPath, err := utils.CheckPath(buildPath)
	if err != nil {
		return err
	}
	// imageName is fixed name.
	// version is read from file or filename, the verification will be added in setVersion.
	// absPath has been verified
	cmdStr := "docker build -t " + ImagePrefix + dd.imageName + ":" + dd.version + " " + absPath + "/."
	if _, err = common.RunCommand("sh", false, "-c", cmdStr); err != nil {
		hwlog.RunLog.Errorf("load docker image [%s] failed:%s", dd.imageName, err)
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
	cmdStr := fmt.Sprintf("docker save %s:%s > %s", ImagePrefix+dd.imageName, dd.version, savePath)
	if _, err = common.RunCommand("sh", false, "-c", cmdStr); err != nil {
		hwlog.RunLog.Errorf("save docker image [%s:%s] failed:%s", dd.imageName, dd.version, err)
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
