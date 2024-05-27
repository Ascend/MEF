// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nginxmgr this package is for manager the nginx
package nginxmgr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"nginx-manager/pkg/msgutil"
	"nginx-manager/pkg/nginxcom"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	retryInterval        = 3 * time.Second
	maxGetNorthCertTimes = 15
	startCommand         = "./nginx"
	accessLogFile        = "/home/MEFCenter/logs/access.log"
	errorLogFile         = "/home/MEFCenter/logs/error.log"
	pidFile              = "/home/MEFCenter/nginx.pid"
)

// initResource initial the resources needed by nginx
func initResource() error {
	if err := updateConf(); err != nil {
		hwlog.RunLog.Errorf("update conf content error: %v", err)
		return err
	}

	if err := prepareCert(); err != nil {
		return err
	}

	if err := prepareCrlFile(); err != nil {
		return err
	}

	if err := prepareFilesMode(); err != nil {
		return err
	}

	return nil
}

func prepareCrlFile() error {
	const (
		crlConfig = "ssl_crl /home/data/config/mef-certs/northern-root.crl;"
	)
	// remove old 3rd north crl
	if err := fileutils.DeleteFile(nginxcom.NorthernCrlFile); err != nil {
		return err
	}
	hwlog.RunLog.Info("start to get north crl from cert manager")
	exist, err := getNorthCrl()
	if err != nil {
		hwlog.RunLog.Errorf("get north crl from cert manager failed: %s", err.Error())
		return err
	}
	content, err := loadConf(nginxcom.NginxConfigPath)
	if err != nil {
		return err
	}
	var modifiedCrlConfig string
	if exist {
		hwlog.RunLog.Info("get north crl from cert manager success")
		modifiedCrlConfig = crlConfig
	} else {
		hwlog.RunLog.Info("get north crl was not imported, nginx with run without ssl crl")
	}
	content = bytes.ReplaceAll(content, []byte(nginxcom.KeyPrefix+nginxcom.CrlConfigKey), []byte(modifiedCrlConfig))
	if err := fileutils.WriteData(nginxcom.NginxConfigPath, content); err != nil {
		hwlog.RunLog.Errorf("writeFile failed. error:%s", err.Error())
		return fmt.Errorf("writeFile failed. error:%s", err.Error())
	}
	hwlog.RunLog.Info("prepare nginx crl success")
	return nil
}

func prepareFilesMode() error {
	fileList := []string{
		accessLogFile,
		errorLogFile,
		pidFile,
	}

	for _, file := range fileList {
		if err := prepareOneFileMode(file); err != nil {
			return err
		}
	}

	return nil
}

func prepareOneFileMode(file string) error {
	if !fileutils.IsExist(file) {
		if err := fileutils.CreateFile(file, fileutils.Mode600); err != nil {
			hwlog.RunLog.Errorf("create %s failed: %v", file, err)
			return fmt.Errorf("create %s failed", file)
		}
	} else {
		if err := fileutils.SetPathPermission(file, fileutils.Mode600, false, false); err != nil {
			hwlog.RunLog.Errorf("chmod %s failed, cause by: {%v}", file, err)
			return fmt.Errorf("set %s's mode failed", file)
		}
	}

	return nil
}

func updateConf() error {
	items, err := CreateConfItems()
	if err != nil {
		return err
	}
	updater, err := NewNginxConfUpdater(items)
	if err != nil {
		return err
	}
	return updater.Update()
}

// LoadKeysDataToPipes decrypt cert keys then write them to pipes
func LoadKeysDataToPipes(deletePipeAfterUse bool) error {
	if err := WritePipe(nginxcom.ServerCertKeyFile, nginxcom.PipePath, deletePipeAfterUse); err != nil {
		return err
	}
	if err := WritePipe(nginxcom.SouthAuthCertKeyFile, nginxcom.AuthPipePath, deletePipeAfterUse); err != nil {
		return err
	}
	if err := WritePipe(nginxcom.WebsocketCertKeyFile, nginxcom.WebsocketPipePath, deletePipeAfterUse); err != nil {
		return err
	}

	thirdPipeCount, err := calculatePipeCount(nginxcom.ThirdPipePrefix)
	if err != nil {
		return err
	}
	if thirdPipeCount != 0 {
		if err := WritePipeForClient(nginxcom.ThirdPartyServiceKeyPath, nginxcom.ClientPipeDir,
			nginxcom.ThirdPipePrefix, thirdPipeCount, deletePipeAfterUse); err != nil {
			return err
		}
	}
	pipeCount, err := calculatePipeCount(nginxcom.ClientPipePrefix)
	if err != nil {
		return err
	}

	return WritePipeForClient(nginxcom.ClientCertKeyFile, nginxcom.ClientPipeDir, nginxcom.ClientPipePrefix,
		pipeCount, deletePipeAfterUse)
}

// CreateKeyPipes create pipes for cert key files.
func CreateKeyPipes() error {
	if err := PreparePipe(nginxcom.PipePath); err != nil {
		return err
	}
	if err := PreparePipe(nginxcom.AuthPipePath); err != nil {
		return err
	}
	if err := PreparePipe(nginxcom.WebsocketPipePath); err != nil {
		return err
	}

	thirdPipeCount, err := calculatePipeCount(nginxcom.ThirdPipePrefix)
	if err != nil {
		return err
	}
	if thirdPipeCount != 0 {
		if err := PrepareForClient(nginxcom.ClientPipeDir, nginxcom.ThirdPipePrefix, thirdPipeCount); err != nil {
			return err
		}
	}
	pipeCount, err := calculatePipeCount(nginxcom.ClientPipePrefix)
	if err != nil {
		return err
	}
	return PrepareForClient(nginxcom.ClientPipeDir, nginxcom.ClientPipePrefix, pipeCount)
}

// DeleteKeyPipes is used to delete all pipe for nginx
func DeleteKeyPipes() {
	if err := fileutils.DeleteAllFileWithConfusion(nginxcom.ClientPipeDir); err != nil {
		hwlog.RunLog.Errorf("delete pipe dir %s failed: %v", nginxcom.ClientPipeDir, err)
	}
}

// CreateConfItems create some items which used to replace into nginx.conf file
func CreateConfItems() ([]nginxcom.NginxConfItem, error) {
	var ret []nginxcom.NginxConfItem
	template := nginxcom.GetConfigItemTemplate()
	for _, item := range template {
		toVal, err := nginxcom.GetEnvManager().Get(item.Key)
		if err != nil {
			return nil, err
		}
		createdItem := nginxcom.NginxConfItem{
			Key:  item.Key,
			From: item.From,
			To:   toVal,
		}
		ret = append(ret, createdItem)
	}
	return ret, nil
}

// NewNginxManager create NewNginxManager module
func NewNginxManager(enable bool, ctx context.Context) model.Module {
	return &nginxManager{
		enable: enable,
		ctx:    ctx,
	}
}

type nginxManager struct {
	enable bool
	ctx    context.Context
}

// Name module name
func (n *nginxManager) Name() string {
	return nginxcom.NginxManagerName
}

// Enable module enable
func (n *nginxManager) Enable() bool {
	return n.enable
}

func startNginx() {
	if err := initResource(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	for {
		if startNginxCmd() {
			return
		}
		hwlog.RunLog.Error("start nginx failed")
		time.Sleep(retryInterval)
	}
}

func getNorthCrl() (bool, error) {
	reqCertParams := getReqCertParams()
	var crlStr string
	var err error
	for i := 0; i < maxGetNorthCertTimes; i++ {
		crlStr, err = reqCertParams.GetCrl(common.NorthernCertName)
		if err != nil {
			hwlog.RunLog.Infof("reqCertParams.GetCrl err: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		if crlStr != "" {
			break
		}
		hwlog.RunLog.Info("north crl is not imported yet, nginx will no config crl")
		return false, nil
	}

	crlMgr, err := x509.NewCrlMgr([]byte(crlStr))
	if err != nil {
		hwlog.RunLog.Errorf("north crl is invalid, error: %v", err)
		return false, err
	}
	// ignore CRL when it is not signed by the same ca
	if err := crlMgr.CheckCrl(x509.CertData{CertPath: nginxcom.NorthernCertFile}); err != nil {
		hwlog.RunLog.Warnf("north crl is not signed by the ca, error: %v", err)
		return false, nil
	}

	if err = fileutils.WriteData(nginxcom.NorthernCrlFile, []byte(crlStr)); err != nil {
		return false, err
	}
	return true, nil
}

func getReqCertParams() requests.ReqCertParams {
	return requests.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: nginxcom.RootCaPath,
			CertPath:   nginxcom.ClientCertFile,
			KeyPath:    nginxcom.ClientCertKeyFile,
			SvrFlag:    false,
			WithBackup: true,
		},
	}
}

// Start module start
func (n *nginxManager) Start() {
	startNginx()
	registerHandlers()
	msgutil.SendVoidMsg(nginxcom.NginxManagerName, nginxcom.NginxMonitorName,
		nginxcom.ReqActiveMonitor, nginxcom.Monitor)
	for {
		select {
		case _, ok := <-n.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("nginx manager catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("nginx manager has listened stop signal")
			return
		default:
		}
		req, err := modulemgr.ReceiveMessage(nginxcom.NginxManagerName)
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request failed", nginxcom.NginxManagerName)
			continue
		}
		msgutil.Handle(req)
	}
}

func registerHandlers() {
	msgutil.RegisterHandlers(msgutil.Combine(nginxcom.ReqRestartNginx, nginxcom.Nginx), reqRestartNginx)
}

func reqRestartNginx(req *model.Message) {
	startNginx()
	msgutil.SendVoidMsg(nginxcom.NginxManagerName, nginxcom.NginxMonitorName, nginxcom.RespRestartNginx, nginxcom.Nginx)
}

func startNginxCmd() bool {
	var err error
	if err = CreateKeyPipes(); err != nil {
		hwlog.RunLog.Errorf("create key pipes failed: %v", err)
		return false
	}
	defer func() {
		if err != nil {
			DeleteKeyPipes()
		}
	}()
	nginxPath, err := filepath.Abs(startCommand)
	if err != nil {
		hwlog.RunLog.Errorf("get nginx abs path failed: %v", err)
		return false
	}
	if _, err = envutils.RunResidentCmd(nginxPath); err != nil {
		hwlog.RunLog.Errorf("start nginx failed: %v", err)
		return false
	}
	if err = LoadKeysDataToPipes(true); err != nil {
		hwlog.RunLog.Errorf("load keys data to pipes failed: %v", err)
		return false
	}
	hwlog.RunLog.Info("run nginx success")
	go childWaitProcess()
	return true
}

func childWaitProcess() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGCHLD)
	for {
		sig := <-sigCh
		if sig != syscall.SIGCHLD {
			return
		}
		if _, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil); err != nil {
			hwlog.RunLog.Errorf("recycle subprocess resources failed, error:%s", err.Error())
		}
		return
	}
}
