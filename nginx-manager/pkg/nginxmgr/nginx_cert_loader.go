// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"fmt"
	"os"
	"syscall"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"

	"nginx-manager/pkg/nginxcom"
)

func prepareServerCert() error {
	keyPath := nginxcom.ServerCertKeyFile
	certPath := nginxcom.ServerCertFile
	if utils.IsExist(keyPath) && utils.IsExist(certPath) {
		hwlog.RunLog.Info("check nginx server certs success")
		return nil
	}
	hwlog.RunLog.Warn("check nginx server certs failed, start to create")
	certStr, err := getServerCert(keyPath)
	if err != nil {
		return err
	}
	err = common.WriteData(certPath, []byte(certStr))
	if err != nil {
		hwlog.RunLog.Errorf("save cert for nginx service cert failed: %s", err.Error())
		return err
	}
	hwlog.RunLog.Info("create cert for nginx service success")
	return nil
}

func getServerCert(keyPath string) (string, error) {
	ips, err := common.GetHostIpV4()
	if err != nil {
		return "", err
	}
	san := certutils.CertSan{IpAddr: ips}
	csr, err := certutils.CreateCsr(keyPath, common.NginxCertName, nil, san)
	if err != nil {
		hwlog.RunLog.Errorf("create nginx service cert csr failed: %s", err.Error())
		return "", err
	}
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    nginxcom.RootCaPath,
			CertPath:      nginxcom.ClientCertFile,
			KeyPath:       nginxcom.ClientCertKeyFile,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	var certStr string
	certStr, err = reqCertParams.ReqIssueSvrCert(common.NginxCertName, csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue certStr for nginx service cert failed: %s", err.Error())
		return "", err
	}
	return certStr, nil
}

// Load load the apigw secret key file and write into pipe
func Load(keyPath, pipePath string) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	err = utils.MakeSureDir(pipePath)
	if err != nil {
		return err
	}
	if utils.IsExist(pipePath) {
		return nil
	}
	err = createPipe(pipePath)
	if err != nil {
		return err
	}
	go writeKeyToPipe(pipePath, keyContent)
	return nil
}

// LoadForClient load the secret key file which used for inner communication and write into pipe
func LoadForClient(keyPath, pipeDir string, pipeCount int) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	var pipePaths []string
	for i := 0; i < pipeCount; i++ {
		pipePath := fmt.Sprintf("%s%s_%d", pipeDir, nginxcom.ClientPipePrefix, i)
		pipePaths = append(pipePaths, pipePath)
	}

	for _, pipePath := range pipePaths {
		if utils.IsExist(pipePath) {
			return nil
		}
		err = createPipe(pipePath)
		if err != nil {
			return err
		}
	}
	for _, pipePath := range pipePaths {
		go writeKeyToPipe(pipePath, keyContent)
	}
	return nil
}

func loadKey(path string) ([]byte, error) {
	encryptKeyContent, err := utils.LoadFile(path)
	if encryptKeyContent == nil {
		hwlog.RunLog.Errorf("load key file failed: %s" + err.Error())
		return nil, fmt.Errorf("load key file failed: %s" + err.Error())
	}
	decryptKeyByte, err := common.DecryptContent(encryptKeyContent, common.GetDefKmcCfg())
	if err != nil {
		hwlog.RunLog.Errorf("decrypt key content failed: %s" + err.Error())
		return nil, fmt.Errorf("decrypt key content failed: %s" + err.Error())
	}
	return decryptKeyByte, nil
}

func writeKeyToPipe(pipeFile string, content []byte) {
	pipe, err := os.OpenFile(pipeFile, os.O_WRONLY|os.O_SYNC, os.ModeNamedPipe)
	if err != nil {
		hwlog.RunLog.Error("open pipe failed")
		return
	}
	defer func() {
		err := pipe.Close()
		if err != nil {
			hwlog.RunLog.Error("pipe close error")
		}
		err = os.Remove(pipeFile)
		if err != nil {
			hwlog.RunLog.Error("pipe remove error")
		}
	}()
	_, err = pipe.Write(content)
	if err != nil {
		hwlog.RunLog.Errorf("pass key to pipe failed:%v", err)
		return
	}
	_, err = pipe.WriteString("\n")
	if err != nil {
		hwlog.RunLog.Errorf("pass key to pipe failed:%v", err)
		return
	}
	hwlog.RunLog.Infof("write pipe %s success", pipeFile)
}

func createPipe(pipeFile string) error {
	_, err := os.Stat(pipeFile)
	if os.IsExist(err) {
		hwlog.RunLog.Infof("pipe %v exists", pipeFile)
		return nil
	}
	err = syscall.Mkfifo(pipeFile, nginxcom.FifoPermission)
	if err != nil {
		hwlog.RunLog.Errorf("make pipe %s error: %s", pipeFile, err.Error())
		return fmt.Errorf("make pipe %s error: %s", pipeFile, err.Error())
	}
	hwlog.RunLog.Infof("make pipe %v success", pipeFile)
	return nil
}
