// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package downloadmgr

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
)

const (
	downloadPackageType = "package"
	downloadCrlType     = "crlFile"
	downloadSignType    = "signFile"
	defaultFileSize     = 1 * constants.MB
	softwarePackageSize = constants.InstallerTarGzSizeMaxInMB * constants.MB
)

type downloadParams struct {
	savePath    string
	sizeLimit   int64
	downloadUrl string
	accountName string
	password    []byte
	caContent   []byte
	crlContent  []byte
}

func (dp *downloadProcess) downloadSoftware() error {
	urls, err := getValidUrls(dp.sfwDownloadInfo)
	if err != nil {
		hwlog.RunLog.Errorf("parse url failed, error: %v", err)
		return err
	}

	for packageType, url := range urls {
		filePath, err := getTargetFilePath(dp.sfwDownloadInfo.SoftwareName, packageType)
		if err != nil {
			return err
		}
		var limit int64 = defaultFileSize
		if packageType == downloadPackageType {
			limit = softwarePackageSize
		}

		if err := envutils.CheckDiskSpace(filepath.Dir(filePath), uint64(limit)); err != nil {
			return err
		}

		info := downloadParams{
			savePath:    filePath,
			sizeLimit:   limit,
			downloadUrl: url,
			accountName: dp.sfwDownloadInfo.DownloadInfo.UserName,
			password:    dp.sfwDownloadInfo.DownloadInfo.Password,
			caContent:   dp.cert,
			crlContent:  dp.crlContent,
		}

		if err = createHttpsReqAndSaveToFile(info); err != nil {
			return err
		}
	}
	dp.progress = progressVerifying
	return nil
}

func parseUrlInfo(url string) (string, error) {
	if url == "" {
		return "", errors.New("url is nil")
	}

	segments := strings.Split(url, " ")
	if len(segments) != constants.URLFieldNum {
		return "", errors.New("url filed num invalid")
	}

	if segments[constants.LocationMethod] != http.MethodGet && segments[constants.LocationMethod] != http.MethodPost {
		return "", errors.New("url method invalid")
	}

	return segments[1], nil
}

func getValidUrls(downloadRequire util.SoftwareDownloadInfo) (map[string]string, error) {
	var urls = make(map[string]string)
	var url string
	var err error
	if url, err = parseUrlInfo(downloadRequire.DownloadInfo.Package); err != nil {
		return urls, fmt.Errorf("package url is invalid because %v", err)
	}

	urls[downloadPackageType] = url

	if url, err = parseUrlInfo(downloadRequire.DownloadInfo.CrlFile); err == nil {
		urls[downloadCrlType] = url
	} else {
		return urls, fmt.Errorf("crl file url is invalid because %v", err)
	}

	if url, err = parseUrlInfo(downloadRequire.DownloadInfo.SignFile); err == nil {
		urls[downloadSignType] = url
	} else {
		return urls, fmt.Errorf("sign file url is invalid because %v", err)
	}

	return urls, nil
}

func createHttpsReqAndSaveToFile(params downloadParams) error {
	if _, err := fileutils.CheckOriginPath(params.savePath); err != nil {
		hwlog.RunLog.Error("create file from https req failed, path is no a file")
		return errors.New("path is no a file")
	}
	fileWriter, err := os.OpenFile(params.savePath, os.O_RDWR|os.O_CREATE, fileutils.Mode600)
	if err != nil {
		return err
	}
	defer func() {
		err = fileWriter.Close()
		if err != nil {
			hwlog.RunLog.Error("close file error")
			return
		}
	}()

	tlsCfg := certutils.TlsCertInfo{
		RootCaContent: params.caContent,
		CrlContent:    params.crlContent,
		RootCaOnly:    true,
		WithBackup:    true,
	}

	cloudUsrAndPwd := append([]byte(params.accountName+":"), params.password...)
	enCodeCloudUsrAndPwdStr := base64.StdEncoding.EncodeToString(cloudUsrAndPwd)
	authorization := "Basic " + enCodeCloudUsrAndPwdStr

	reqHeaders := map[string]interface{}{
		"Authorization": authorization,
		"nodeID":        configpara.GetInstallerConfig().SerialNumber,
	}

	defer utils.ClearSliceByteMemory(cloudUsrAndPwd)
	defer utils.ClearStringMemory(enCodeCloudUsrAndPwdStr)
	defer utils.ClearStringMemory(authorization)

	return httpsmgr.GetHttpsReq(params.downloadUrl, tlsCfg, reqHeaders).
		GetRespToFileWithLimit(fileWriter, params.sizeLimit)
}

func getTargetFilePath(softwareName string, packageType string) (string, error) {
	packageDir := constants.EdgeDownloadPath

	var fileName string
	switch packageType {
	case downloadPackageType:
		fileName = fmt.Sprintf("%s%s", softwareName, constants.TarGzExt)
	case downloadCrlType:
		fileName = fmt.Sprintf("%s%s", softwareName, constants.CrlExt)
	case downloadSignType:
		fileName = fmt.Sprintf("%s%s", softwareName, constants.SignExt)
	default:
		return "", errors.New("invalid package type")

	}

	return filepath.Join(packageDir, fileName), nil

}
