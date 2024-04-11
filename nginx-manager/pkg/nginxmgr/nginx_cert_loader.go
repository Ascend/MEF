// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509/certutils"

	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	maxRetry           = 10
	waitTime           = 5 * time.Second
	httpReqTryInterval = time.Second * 30
	httpReqTryMaxTime  = 5
)

func prepareCert() error {
	if err := prepareServerCert(nginxcom.ServerCertKeyFile, nginxcom.ServerCertFile, common.NginxCertName); err != nil {
		return err
	}
	if err := prepareRootCert(nginxcom.NorthernCertFile, common.NorthernCertName, false); err != nil {
		hwlog.RunLog.Errorf("get root ca(%s) failed: %v", common.NorthernCertName, err)
		return err
	}
	if err := prepareServerCert(nginxcom.SouthAuthCertKeyFile, nginxcom.SouthAuthCertFile, common.WsSerName); err != nil {
		return err
	}
	if err := prepareServerCert(nginxcom.WebsocketCertKeyFile, nginxcom.WebsocketCertFile, common.WsSerName); err != nil {
		return err
	}
	if err := prepareRootCert(nginxcom.SouthernCertFile, common.WsCltName, true); err != nil {
		hwlog.RunLog.Errorf("get root ca(%s) failed: %v", common.WsCltName, err)
		return err
	}
	return nil
}

func prepareRootCert(certPath, certName string, retry bool) error {
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: nginxcom.RootCaPath,
			CertPath:   nginxcom.ClientCertFile,
			KeyPath:    nginxcom.ClientCertKeyFile,
			WithBackup: true,
		},
	}
	var rootCaStr string
	var err error
	for i := 0; i < maxRetry; i++ {
		rootCaStr, err = reqCertParams.GetRootCa(certName)
		if err != nil && retry {
			time.Sleep(waitTime)
			continue
		}
		break
	}
	if rootCaStr == "" {
		return err
	}

	// To ensure the latest root certificate, overwritten root each restarted.
	err = fileutils.WriteData(certPath, []byte(rootCaStr))
	if err != nil {
		hwlog.RunLog.Errorf("save cert for %s service cert failed: %s", certName, err.Error())
		return err
	}
	return nil
}

func prepareServerCert(keyPath string, certPath string, server string) error {
	certStr, err := getServerCert(keyPath, server)
	if err != nil {
		return err
	}
	err = fileutils.WriteData(certPath, []byte(certStr))
	if err != nil {
		hwlog.RunLog.Errorf("save cert for %s service cert failed: %s", server, err.Error())
		return err
	}
	hwlog.RunLog.Infof("create cert for %s service success", server)
	return nil
}

func getServerCert(keyPath string, server string) (string, error) {
	ips, err := common.GetHostIpV4()
	if err != nil {
		return "", err
	}
	san := certutils.CertSan{IpAddr: ips}
	csr, err := certutils.CreateCsr(keyPath, common.MefCertCommonNamePrefix, nil, san)
	if err != nil {
		hwlog.RunLog.Errorf("create %s service cert csr failed: %s", server, err.Error())
		return "", err
	}
	reqCertParams := requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: nginxcom.RootCaPath,
			CertPath:   nginxcom.ClientCertFile,
			KeyPath:    nginxcom.ClientCertKeyFile,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
	var certStr string
	certStr, err = reqCertParams.ReqIssueSvrCert(server, csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue certStr for nginx service cert failed: %s", err.Error())
		return "", err
	}
	return certStr, nil
}

// PreparePipe Check and create a pipeline file
func PreparePipe(pipePath string) error {
	var err error
	err = fileutils.MakeSureDir(pipePath)
	if err != nil {
		return err
	}
	if fileutils.IsExist(pipePath) {
		return nil
	}
	err = createPipe(pipePath)
	if err != nil {
		return err
	}
	return nil
}

// WritePipe load the apigw secret key file and write into pipe
func WritePipe(keyPath, pipePath string, deletePipeAfterUse bool) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	if !fileutils.IsExist(pipePath) {
		common.ClearSliceByteMemory(keyContent)
		return fmt.Errorf("the file: %s does not exist", pipePath)
	}

	go writeKeyToPipe(pipePath, keyContent, deletePipeAfterUse)
	return nil
}

// PrepareForClient creates a pipe file for internal communication.
func PrepareForClient(pipeDir string, pipeCount int) error {
	var err error
	var pipePaths []string
	for i := 0; i < pipeCount; i++ {
		pipePath := fmt.Sprintf("%s%s_%d", pipeDir, nginxcom.ClientPipePrefix, i)
		pipePaths = append(pipePaths, pipePath)
	}

	for _, pipePath := range pipePaths {
		if fileutils.IsExist(pipePath) {
			return nil
		}
		err = createPipe(pipePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// WritePipeForClient load the secret key file which used for inner communication and write into pipe
func WritePipeForClient(keyPath, pipeDir string, pipeCount int, deletePipeAfterUse bool) error {
	var keyContent []byte
	var err error
	if keyContent, err = loadKey(keyPath); err != nil {
		return err
	}
	defer common.ClearSliceByteMemory(keyContent)

	var pipePaths []string
	for i := 0; i < pipeCount; i++ {
		pipePath := fmt.Sprintf("%s%s_%d", pipeDir, nginxcom.ClientPipePrefix, i)
		pipePaths = append(pipePaths, pipePath)
	}

	for _, pipePath := range pipePaths {
		if !fileutils.IsExist(pipePath) {
			return fmt.Errorf("the file: %s does not exist", pipePath)
		}
		content := make([]byte, len(keyContent))
		copy(content, keyContent)
		go writeKeyToPipe(pipePath, content, deletePipeAfterUse)
	}
	return nil
}

func loadKey(path string) ([]byte, error) {
	encryptKeyContent, err := fileutils.LoadFile(path)
	if err != nil {
		hwlog.RunLog.Errorf("load key file failed: %s", err.Error())
		return nil, fmt.Errorf("load key file failed: %s", err.Error())
	}
	if encryptKeyContent == nil {
		hwlog.RunLog.Error("load key file returns empty content")
		return nil, errors.New("load key file returns empty content")
	}
	decryptKeyByte, err := kmc.DecryptContent(encryptKeyContent, kmc.GetDefKmcCfg())
	if err != nil {
		hwlog.RunLog.Errorf("decrypt key content failed: %s", err.Error())
		return nil, fmt.Errorf("decrypt key content failed: %s", err.Error())
	}
	return decryptKeyByte, nil
}

func writeKeyToPipe(pipeFile string, content []byte, deletePipeAfterUse bool) {
	defer func() {
		common.ClearSliceByteMemory(content)
		if deletePipeAfterUse {
			err := fileutils.DeleteFile(pipeFile)
			if err != nil {
				hwlog.RunLog.Error("pipe remove error")
			}
		}
	}()
	pipe, err := os.OpenFile(pipeFile, os.O_WRONLY|os.O_SYNC, os.ModeNamedPipe)
	if err != nil {
		hwlog.RunLog.Error("open pipe failed")
		return
	}
	defer func() {
		err = pipe.Close()
		if err != nil {
			hwlog.RunLog.Error("pipe close error")
		}
	}()
	var nginxStatus bool
	for i := 0; i < maxRetry; i++ {
		nginxStatus = isNginxRunning()
		if nginxStatus {
			break
		}
		time.Sleep(waitTime)
	}
	if !nginxStatus {
		hwlog.RunLog.Warn("nginx is not running, skip writing tls cert key")
		return
	}
	_, err = pipe.Write(content)
	if err != nil {
		hwlog.RunLog.Errorf("write key to pipe failed: %v", err)
		return
	}
	_, err = pipe.WriteString("\n")
	if err != nil {
		hwlog.RunLog.Errorf("write `\\n` to pipe failed: %v", err)
		return
	}
	hwlog.RunLog.Infof("write pipe %s success", pipeFile)
}

func isNginxRunning() bool {
	pid, err := envutils.RunCommand("pgrep", envutils.DefCmdTimeoutSec, "-o", "-x", "nginx")
	if err != nil || pid == "" {
		return false
	}
	nginxPID := strings.TrimSpace(pid)
	ppid, err := envutils.RunCommand("ps", envutils.DefCmdTimeoutSec, "-o", "ppid=", "-p", nginxPID)
	if err != nil || ppid == "" {
		return false
	}
	nginxPPID := strings.TrimSpace(ppid)
	// To prevent forgery
	if nginxPPID != "1" {
		return false
	}
	return true
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

// PrepareServiceCert prepare cert and key files with force flag: if true, delete them first then create new
func PrepareServiceCert(keyPath string, certPath string, caName string, forceFlag bool, locker *sync.Mutex) error {
	if forceFlag {
		// key and cert files will be read when nginx conf is reloading
		// try to get the lock before deleting them
		// get the lock success means reload operation is finished, it will be safe to delete them
		locker.Lock()
		defer locker.Unlock()
		if err := fileutils.DeleteAllFileWithConfusion(keyPath); err != nil {
			return fmt.Errorf("remove old key file error: %v", err)
		}
		hwlog.RunLog.Infof("key file [%v] is deleted", keyPath)
		if err := fileutils.DeleteFile(certPath); err != nil {
			return fmt.Errorf("remove old cert file error: %v", err)
		}
		hwlog.RunLog.Infof("cert file [%v] is deleted", certPath)
	}
	var tryCnt int
	// try multiple times until cert-manager https server is ready
	for tryCnt = 0; tryCnt < httpReqTryMaxTime; tryCnt++ {
		if err := prepareServerCert(keyPath, certPath, caName); err != nil {
			hwlog.RunLog.Errorf("retrieve service cert [%v], get error: %v, try request for next time", certPath, err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		break
	}
	if tryCnt == httpReqTryMaxTime {
		return fmt.Errorf("retrieve service cert [%v] failed, please check network connection", certPath)
	}
	return nil
}
