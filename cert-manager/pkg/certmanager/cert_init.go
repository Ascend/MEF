// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager cert manager module
package certmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"cert-manager/pkg/config"
)

type handlerFunc func(req interface{}) common.RespMsg

type certManager struct {
	enable bool
	ctx    context.Context
}

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
	res = method(req.GetContent())
	return &res
}

func (cm *certManager) Start() {
	go periodCheck()
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
	resp.FillContent(msg)
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", cm.Name())
		return
	}
}

var (
	certUrlRootPath      = "/certmanager/v1/certificates"
	crlUrlRootPath       = "/certmanager/v1/crl"
	innerCertUrlRootPath = "/inner/v1/certificates"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "import")):      importRootCa,
	common.Combine(http.MethodPost, filepath.Join(certUrlRootPath, "delete-cert")): deleteRootCa,
	common.Combine(http.MethodGet, filepath.Join(certUrlRootPath, "info")):         getCertInfo,
	common.Combine(http.MethodPost, filepath.Join(crlUrlRootPath, "import")):       importCrl,

	common.Combine(http.MethodGet, filepath.Join(innerCertUrlRootPath, "rootca")):   queryRootCa,
	common.Combine(http.MethodGet, filepath.Join(innerCertUrlRootPath, "crl")):      queryCrl,
	common.Combine(http.MethodPost, filepath.Join(innerCertUrlRootPath, "service")): issueServiceCa,
}

func periodCheck() {
	if err := checkCert(); err != nil {
		hwlog.RunLog.Errorf("check cert overdue error:%v", err)
	}
	ticker := time.NewTicker(common.OneDay)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			err := checkCert()
			if err != nil {
				hwlog.RunLog.Errorf("check cert overdue error:%v", err)
				continue
			}
		}
	}
}

var certLock sync.RWMutex

func checkCert() error {
	certData, err := utils.LoadFile(getRootCaPath(common.WsCltName))
	if err != nil {
		return err
	}
	if certData == nil {
		return nil
	}
	crt, err := x509.LoadCertsFromPEM(certData)
	if err != nil {
		return err
	}
	overdueDay, err := x509.GetValidityPeriod(crt)
	overdueData := config.GetCertConfig().CertExpireTime
	if err == nil && (overdueDay > float64(overdueData)) {
		return nil
	}
	if err := updateCert(); err != nil {
		hwlog.RunLog.Error("update cert error, %v", err)
		return err
	}
	hwlog.RunLog.Info("websocket client cert will overdue, update success")
	return nil
}

func updateCert() error {
	certLock.Lock()
	defer lock.Unlock()
	if err := CreateCaIfNotExit(); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	if err := sendCertUpdateMsg(); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	return nil
}

func sendCertUpdateMsg() error {
	tls := certutils.TlsCertInfo{
		RootCaPath: util.RootCaPath,
		CertPath:   util.ServerCertPath,
		KeyPath:    util.ServerKeyPath,
		SvrFlag:    false,
	}

	url := fmt.Sprintf("https://%s:%d/%s", common.EdgeMgrDns, common.EdgeMgrPort, "edgemanager/v1/cert/update")
	httpsReq := httpsmgr.GetHttpsReq(url, tls)
	respByte, err := httpsReq.Get(nil)
	if err != nil {
		return err
	}
	var resp common.RespMsg
	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return err
	}

	status := resp.Status
	if status != common.Success {
		return fmt.Errorf("parse cert response failed: status=%s, msg=%s", status, resp.Msg)
	}
	return nil
}
