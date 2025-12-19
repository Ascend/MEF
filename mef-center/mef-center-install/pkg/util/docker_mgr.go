// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

// DockerDealer is a struct to handle docker images
type DockerDealer struct {
	imageName string
	tag       string
}

// GetAscendDockerDealer inits a DockerDealer struct with ascend prefix on image name
func GetAscendDockerDealer(componentName string, tag string) DockerDealer {
	return DockerDealer{
		imageName: ImagePrefix + componentName,
		tag:       tag,
	}
}

// GetDockerDealer inits a DockerDealer with iamge name and tag
func GetDockerDealer(imageName, tag string) DockerDealer {
	return DockerDealer{
		imageName: imageName,
		tag:       tag,
	}
}

// LoadImage is used to build a docker image by Dockerfile
func (dd *DockerDealer) LoadImage(buildPath string) error {
	absPath, err := fileutils.CheckOriginPath(buildPath)
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
	if len(lines) > common.MaxLoopNum {
		hwlog.RunLog.Error("the number of images exceed the upper limit")
		return false, errors.New("the number of images exceed the upper limit")
	}
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

// CheckDependentImage is the func to check the existence of dependent image
func CheckDependentImage() error {
	var dependentImageMap = map[string]string{
		common.UbuntuImageName: common.UbuntuImageTag,
	}

	for name, version := range dependentImageMap {
		dockerDealerIns := GetDockerDealer(name, version)
		exists, err := dockerDealerIns.CheckImageExists()
		if err != nil {
			hwlog.RunLog.Errorf("check if %s:%s image exists failed: %v", name, version, err)
			return errors.New("check image existence failed")
		}

		if !exists {
			fmt.Printf("docker images %s:%s does not exist, please prepare it first\n", name, version)
			hwlog.RunLog.Errorf("image %s:%s does not exist", name, version)
			return errors.New("dependent image does node exist")
		}
	}

	return nil
}
