// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	certModificationCheckInterval = 5 * time.Minute
)

type handlerFunc func(message *model.Message) common.RespMsg

type certManager struct {
	enable bool
	ctx    context.Context
}

var certMonitor *common.FileMonitor

// NewCertManager create cert manager
func NewCertManager(enable bool) model.Module {
	cm := &certManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return cm
}

func (cm *certManager) Name() string {
	return common.CertManagerName
}

func (cm *certManager) Enable() bool {
	return cm.enable
}

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, exit := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !exit {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil
	}
	res = method(req)
	return &res
}

func (cm *certManager) Start() {
	go certExpireCheck(cm.ctx)
	go certChangesCheck(cm.ctx)
	for {
		select {
		case _, ok := <-cm.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(cm.Name())
		hwlog.RunLog.Infof("%s receive request from restful service", cm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", cm.Name())
			continue
		}

		go cm.dispatch(req)
	}
}

func (cm *certManager) dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", cm.Name())
		return
	}
	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", cm.Name())
		return
	}
	if err = resp.FillContent(msg); err != nil {
		hwlog.RunLog.Errorf("%s fill content failed: %v", cm.Name(), err)
		return
	}
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", cm.Name())
		return
	}
}

var (
	certUrlRootPath         = "/certmanager/v1/certificates"
	crlUrlRootPath          = "/certmanager/v1/crl"
	innerCertUrlRootPath    = "/inner/v1/certificates"
	getImportedCertsInfoUrl = "/inner/v1/certificates/imported-certs"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "import")):      importRootCa,
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "delete-cert")): deleteRootCa,
	common.Combine(http.MethodGet, filepath.Join(certUrlRootPath, "info")):         getCertInfo,
	common.Combine(http.MethodPost, filepath.Join(crlUrlRootPath, "import")):       importCrl,

	common.Combine(http.MethodGet, filepath.Join(innerCertUrlRootPath, "rootca")):         queryRootCa,
	common.Combine(http.MethodGet, filepath.Join(innerCertUrlRootPath, "crl")):            queryCrl,
	common.Combine(http.MethodPost, filepath.Join(innerCertUrlRootPath, "service")):       issueServiceCa,
	common.Combine(http.MethodPost, filepath.Join(innerCertUrlRootPath, "update-result")): certsUpdateResult,
	common.Combine(http.MethodGet, getImportedCertsInfoUrl):                               getImportedCertsInfo,
}

func certExpireCheck(ctx context.Context) {
	checkCaCerts()

	ticker := time.NewTicker(common.OneDay)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("cert check operation is aborted")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Error("cert check operation is stopped")
				return
			}
			checkCaCerts()
		}
	}
}

func checkCaCerts() {
	go doCheckProcess(NewCertUpdater(CertTypeEdgeSvc))
	go doCheckProcess(NewCertUpdater(CertTypeEdgeCa))
}

func doCheckProcess(updater CertUpdater) {
	if updater == nil {
		hwlog.RunLog.Error("invalid cert updater instance")
		return
	}
	if err := updater.CheckAndSetUpdateFlag(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	go updater.ClearUpdateFlag()
	needUpdate, needForceUpdate, err := updater.IsCertNeedUpdate()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	if !needUpdate {
		return
	}
	if needForceUpdate {
		if err = updater.DoForceUpdate(); err != nil {
			hwlog.RunLog.Error(err)
			return
		}
	}
	if err = updater.PrepareCertUpdate(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	if err = updater.NotifyCertUpdate(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}
	go updater.PostCertUpdate()
	go updater.ForceUpdateCheck()
}

func certChangesCheck(ctx context.Context) {
	var certAndCrlPaths = []string{
		getRootCaPath(common.ImageCertName),
		getRootCaPath(common.SoftwareCertName),
		getCrlPath(common.ImageCertName),
		getCrlPath(common.SoftwareCertName),
	}
	certMonitor = common.NewFileMonitor(certModificationCheckInterval, onCertOrCrlChanged, certAndCrlPaths...)
	certMonitor.Run(ctx)
}

func onCertOrCrlChanged(changedFiles []string) {
	for _, filePath := range changedFiles {
		caName := filepath.Base(filepath.Dir(filePath))
		operation := common.Update
		if strings.HasSuffix(filePath, util.CertSuffix) &&
			!fileutils.IsExist(filePath) && !fileutils.IsExist(caName+backuputils.BackupSuffix) {
			operation = common.Delete
		}
		hwlog.RunLog.Warnf("cert [%s] has been %sd, try to notify clients", caName, operation)
		if err := updateClientCert(caName, operation); err != nil {
			hwlog.RunLog.Errorf("cert [%s]: notify clients failed, %v", caName, err)
			continue
		}
		hwlog.RunLog.Infof("cert [%s]: notify clients successfully", caName)
	}
}
