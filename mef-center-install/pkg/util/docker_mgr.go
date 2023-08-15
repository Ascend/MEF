// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
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

	ret, err := dd.checkImageExist()
	if err != nil {
		return err
	}

	if ret {
		hwlog.RunLog.Errorf("same docker image [%s] exists,cannot reload it", dd.imageName)
		return errors.New("load docker image failed")
	}

	mefUid, mefGid, err := GetMefId()
	if err != nil {
		hwlog.RunLog.Error("Get MEFCenter uid/gid failed")
		return errors.New("get MEFCenter uid/gid failed")
	}
	uidArg := fmt.Sprintf("UID=%d", mefUid)
	gidArg := fmt.Sprintf("GID=%d", mefGid)

	// imageName is fixed name.
	// tag is read from file or filename, the verification will be added in setVersion.
	// absPath has been verified
	if _, err = envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "build", "--build-arg",
		uidArg, "--build-arg", gidArg, "-t", dd.imageName+":"+dd.tag, absPath+"/."); err != nil {
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
	fullImage := fmt.Sprintf("%s:%s", dd.imageName, dd.tag)
	if _, err = envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "save", fullImage,
		"-o", savePath); err != nil {
		hwlog.RunLog.Errorf("save docker image [%s:%s] failed:%s", dd.imageName, dd.tag, err.Error())
		return errors.New("save docker image failed")
	}

	return nil

}

// CheckImageExists checks if the docker image exists
func (dd *DockerDealer) CheckImageExists() (bool, error) {
	const (
		imageNameColumn  = 0
		imageTagColumn   = 1
		imagesMinColumns = 2
	)
	images, err := envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "images")
	if err != nil {
		hwlog.RunLog.Errorf("get all docker images failed:%s", err.Error())
		return false, errors.New("get all docker images failed")
	}

	lines := strings.Split(images, "\n")
	for _, line := range lines {
		columns := strings.Fields(line)
		if len(columns) < imagesMinColumns {
			continue
		}

		if columns[imageNameColumn] == dd.imageName && columns[imageTagColumn] == dd.tag {
			return true, nil
		}
	}

	return false, nil
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
	ret, err := envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "image", "ls", dd.imageName+":"+dd.tag)
	if err != nil {
		hwlog.RunLog.Errorf("check %s's docker image command exec failed: %s", dd.imageName, err.Error())
		return false, fmt.Errorf("check %s's docker image command exec failed", dd.imageName)
	}

	lines := strings.Split(ret, "\n")
	if len(lines) == DockerImageExist {
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

	_, err = envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "rmi", dd.imageName+":"+dd.tag)
	if err != nil {
		hwlog.RunLog.Errorf("delete %s's docker image command exec failed: %s", dd.imageName, err.Error())
		return fmt.Errorf("delete %s's docker image command exec failed", dd.imageName)
	}

	return nil
}

// ReloadImage is used to reload docker image from saved file
func (dd *DockerDealer) ReloadImage(imageDirPath string) error {
	imageName, err := dd.getImageTarName()
	if err != nil {
		return fmt.Errorf("get image tar name failed: %s", err.Error())
	}

	imagePath := filepath.Join(imageDirPath, imageName)
	_, err = envutils.RunCommand(CommandDocker, envutils.DefCmdTimeoutSec, "load", "-i", imagePath)
	if err != nil {
		return fmt.Errorf("reload docker image %s failed: %s", imagePath, err.Error())
	}
	return nil
}
