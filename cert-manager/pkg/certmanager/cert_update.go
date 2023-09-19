// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package certmanager check root ca cert expiration and do update operations
package certmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"cert-manager/pkg/config"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	// NotUpdating not in updating state
	NotUpdating int64 = 0
	// InUpdating  in updating state
	InUpdating int64 = 1
	// CertTypeEdgeCa cert type for edge service cert
	CertTypeEdgeCa = "EdgeCa"
	// CertTypeEdgeSvc cert type for edge root ca
	CertTypeEdgeSvc = "EdgeSvc"
)

var edgeCaResultChan = make(chan certUpdateResult)
var edgeSvcResultChan = make(chan certUpdateResult)
var edgeCaUpdatingFlag int64
var edgeSvcUpdatingFlag int64

// CertUpdater cert update abstract operations
type CertUpdater interface {
	// CheckAndSetUpdateFlag check and set updating flag to InUpdating
	CheckAndSetUpdateFlag() error
	// ClearUpdateFlag reset updating flag to NotUpdating
	ClearUpdateFlag()
	// IsCertNeedUpdate check if cert need to be updated
	IsCertNeedUpdate() (bool, bool, error)
	// PrepareCertUpdate generate new certs, save them to temp path
	PrepareCertUpdate() error
	// NotifyCertUpdate send certs update notify to all managed edge node by cloudhub
	NotifyCertUpdate() error
	// PostCertUpdate  wait for certs update results, replace old certs with new certs, clean up temporary certs.
	PostCertUpdate()
	// ForceUpdateCheck check certs in background, when certs will be expired(<24h),
	// ignore normal update process, do force update process.
	ForceUpdateCheck()
	// DoForceUpdate do force update cert operation
	DoForceUpdate() error
}

// EdgeSvcCertUpdater cert update operation implementation for edge service cert
type EdgeSvcCertUpdater struct {
	ctx               context.Context
	cancel            context.CancelFunc
	CaCertName        string `json:"caCertName"`
	TempCaCertContent string `json:"tmpCaCertContent"`
}

// EdgeCaCertUpdater cert update operation implementation for edge root ca cert
type EdgeCaCertUpdater struct {
	ctx               context.Context
	cancel            context.CancelFunc
	CaCertName        string `json:"caCertName"`
	TempCaCertContent string `json:"tmpCaCertContent"`
}

// CertUpdatePayload cert update payload data, sent to edge-manager
type CertUpdatePayload struct {
	CertType    string `json:"certType"`
	ForceUpdate bool   `json:"forceUpdate"`
	CaContent   string `json:"caContent"`
}

// NewCertUpdater get a cert update operation instance
func NewCertUpdater(certType string) CertUpdater {
	switch certType {
	case CertTypeEdgeCa:
		var instance EdgeCaCertUpdater
		instance.CaCertName = common.WsSerName
		instance.ctx, instance.cancel = context.WithCancel(context.Background())
		return &instance
	case CertTypeEdgeSvc:
		var instance EdgeSvcCertUpdater
		instance.CaCertName = common.WsCltName
		instance.ctx, instance.cancel = context.WithCancel(context.Background())
		return &instance
	default:
		hwlog.RunLog.Errorf("invalid cert type: %v", certType)
		return nil
	}
}

// CheckAndSetUpdateFlag check and set updating flag for hub_client
func (svc *EdgeSvcCertUpdater) CheckAndSetUpdateFlag() error {
	if !atomic.CompareAndSwapInt64(&edgeSvcUpdatingFlag, NotUpdating, InUpdating) {
		return fmt.Errorf("edge service cert is in updating, try it later")
	}
	return nil
}

// ClearUpdateFlag clear updating flag when hub_client update operation is finished
func (svc *EdgeSvcCertUpdater) ClearUpdateFlag() {
	select {
	case _, _ = <-svc.ctx.Done():
		atomic.StoreInt64(&edgeSvcUpdatingFlag, NotUpdating)
	}
	// reset context
	svc.ctx, svc.cancel = context.WithCancel(context.Background())
	hwlog.RunLog.Info("edge service cert update is finished")
}

// IsCertNeedUpdate check if root ca hub_client need to be updated
func (svc *EdgeSvcCertUpdater) IsCertNeedUpdate() (bool, bool, error) {
	needUpdate, needForceUpdate, err := checkCertValidity(svc.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("check cert validity error: %v", err)
		svc.cancel()
		return false, false, err
	}
	if needForceUpdate {
		hwlog.RunLog.Infof("cert [%v] will be force updated", svc.CaCertName)
		return true, true, nil
	}
	if needUpdate {
		hwlog.RunLog.Infof("cert [%v] will be updated in normal way", svc.CaCertName)
		return true, false, nil
	}
	hwlog.RunLog.Infof("cert [%v] is no need to update. abort update operation", svc.CaCertName)
	svc.cancel()
	return false, false, nil
}

// PrepareCertUpdate create new root ca cert hub_client when update operation starts
func (svc *EdgeSvcCertUpdater) PrepareCertUpdate() error {
	hwlog.RunLog.Infof("cert [%v] update operation starts", svc.CaCertName)
	var err error
	svc.TempCaCertContent, err = CreateTempCaCert(svc.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("create temp new cert for [%v] failed: %v", svc.CaCertName, err)
		svc.cancel()
	}
	hwlog.RunLog.Infof("create temp new cert for [%v] success", svc.CaCertName)
	return err
}

// NotifyCertUpdate send hub_client update notify to edge-manager
func (svc *EdgeSvcCertUpdater) NotifyCertUpdate() error {
	payload := CertUpdatePayload{
		CertType:  CertTypeEdgeSvc,
		CaContent: svc.TempCaCertContent,
	}
	if err := sendCertUpdateNotify(payload); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] update notify error:%v", svc.CaCertName, err)
		svc.cancel()
		return err
	}
	hwlog.RunLog.Infof("send cert [%v] update notify success", svc.CaCertName)
	return nil
}

// PostCertUpdate post process when hub_client update operation is finished
func (svc *EdgeSvcCertUpdater) PostCertUpdate() {
	select {
	case _, _ = <-svc.ctx.Done():
		hwlog.RunLog.Warnf("cert [%v] update post process is cancelled", svc.CaCertName)
	case result := <-edgeSvcResultChan:
		if result.ResultCode != updateSuccessCode {
			hwlog.RunLog.Errorf("get cert [%v] failed update result, reason: %v", svc.CaCertName, result.Desc)
			svc.cancel()
			return
		}
		if err := svc.DoForceUpdate(); err != nil {
			hwlog.RunLog.Errorf("do cert [%v] final force update operation failed: %v", svc.CaCertName, err)
		}
		hwlog.RunLog.Infof("cert [%v] update operation post process success", svc.CaCertName)
	}
}

// ForceUpdateCheck check if the root ca cert hub_client remaining valid time is <= 24h
func (svc *EdgeSvcCertUpdater) ForceUpdateCheck() {
	_, forceUpdate, err := checkCertValidity(svc.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("check cert [%v] validity failed: %v", svc.CaCertName, err)
		svc.cancel()
		return
	}
	if forceUpdate {
		if err = svc.DoForceUpdate(); err != nil {
			hwlog.RunLog.Errorf("do cert [%v] force update operation failed: %v", svc.CaCertName, err)
		}
		hwlog.RunLog.Infof("cert [%v] force update operation success", svc.CaCertName)
		return
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-svc.ctx.Done():
			hwlog.RunLog.Warnf("cert [%v] force-update check operation is cancelled", svc.CaCertName)
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Warnf("period cert [%v] force-update check operation is stopped", svc.CaCertName)
				svc.cancel()
				return
			}
			_, forceUpdate, err = checkCertValidity(svc.CaCertName)
			if err != nil {
				hwlog.RunLog.Errorf("check cert [%v] validity failed: %v", svc.CaCertName, err)
				continue
			}
			if !forceUpdate {
				continue
			}
			if err = svc.DoForceUpdate(); err != nil {
				hwlog.RunLog.Errorf("do cert [%v] force update operation failed: %v", svc.CaCertName, err)
				return
			}
			hwlog.RunLog.Infof("cert [%v] force update operation success", svc.CaCertName)
			return
		}
	}
}

// CheckAndSetUpdateFlag check and set updating flag for hub_svr
func (ca *EdgeCaCertUpdater) CheckAndSetUpdateFlag() error {
	if !atomic.CompareAndSwapInt64(&edgeCaUpdatingFlag, NotUpdating, InUpdating) {
		return fmt.Errorf("edge root ca cert is in updating, try it later")
	}
	return nil
}

// ClearUpdateFlag clear updating flag when hub_svr update operation is finished
func (ca *EdgeCaCertUpdater) ClearUpdateFlag() {
	select {
	case _, _ = <-ca.ctx.Done():
		atomic.StoreInt64(&edgeCaUpdatingFlag, NotUpdating)
	}
	// reset context
	ca.ctx, ca.cancel = context.WithCancel(context.Background())
	hwlog.RunLog.Info("edge root ca cert update is finished")
}

// IsCertNeedUpdate check if root ca hub_svr need to be updated
func (ca *EdgeCaCertUpdater) IsCertNeedUpdate() (bool, bool, error) {
	needUpdate, needForceUpdate, err := checkCertValidity(ca.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("check cert validity error: %v", err)
		ca.cancel()
		return false, false, err
	}
	if needForceUpdate {
		hwlog.RunLog.Infof("cert [%v] will be force updated", ca.CaCertName)
		return true, true, nil
	}
	if needUpdate {
		hwlog.RunLog.Infof("cert [%v] will be updated in normal way", ca.CaCertName)
		return true, false, nil
	}
	hwlog.RunLog.Infof("cert [%v] is no need to update. abort update operation", ca.CaCertName)
	ca.cancel()
	return false, false, nil
}

// PrepareCertUpdate create new root ca hub_svr when update operation starts
func (ca *EdgeCaCertUpdater) PrepareCertUpdate() error {
	hwlog.RunLog.Infof("cert [%v] update operation starts", ca.CaCertName)
	var err error
	ca.TempCaCertContent, err = CreateTempCaCert(ca.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("create temp new cert for [%v] failed: %v", ca.CaCertName, err)
		ca.cancel()
	}
	hwlog.RunLog.Infof("create temp new cert for [%v] success", ca.CaCertName)
	return err
}

// NotifyCertUpdate send hub_svr update notify to edge-manager
func (ca *EdgeCaCertUpdater) NotifyCertUpdate() error {
	payload := CertUpdatePayload{
		CertType:  CertTypeEdgeCa,
		CaContent: ca.TempCaCertContent,
	}
	if err := sendCertUpdateNotify(payload); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] update notify error:%v", ca.CaCertName, err)
		ca.cancel()
		return err
	}
	hwlog.RunLog.Infof("send cert [%v] update notify success", ca.CaCertName)
	return nil
}

// PostCertUpdate post process when hub_svr update operation is finished
func (ca *EdgeCaCertUpdater) PostCertUpdate() {
	select {
	case _, _ = <-ca.ctx.Done():
		hwlog.RunLog.Warnf("cert [%v] update post process is cancelled", ca.CaCertName)
	case result := <-edgeCaResultChan:
		if result.ResultCode != updateSuccessCode {
			hwlog.RunLog.Errorf("get cert [%v] failed update result, reason: %v", ca.CaCertName, result.Desc)
			ca.cancel()
			return
		}
		if err := ca.DoForceUpdate(); err != nil {
			hwlog.RunLog.Errorf("do cert [%v] final force update operation failed: %v", ca.CaCertName, err)
		}
		hwlog.RunLog.Infof("cert [%v] update operation post process success", ca.CaCertName)
	}
}

// ForceUpdateCheck check if the root ca cert hub_svr remaining valid time is <= 24h
func (ca *EdgeCaCertUpdater) ForceUpdateCheck() {
	_, forceUpdate, err := checkCertValidity(ca.CaCertName)
	if err != nil {
		hwlog.RunLog.Errorf("check cert [%v] validity failed: %v", ca.CaCertName, err)
		ca.cancel()
		return
	}
	if forceUpdate {
		if err = ca.DoForceUpdate(); err != nil {
			hwlog.RunLog.Errorf("do cert [%v] force update operation failed: %v", ca.CaCertName, err)
		}
		hwlog.RunLog.Infof("cert [%v] force update operation success", ca.CaCertName)
		return
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ca.ctx.Done():
			hwlog.RunLog.Warnf("cert [%v] force-update check operation is cancelled", ca.CaCertName)
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Warnf("period cert [%v] force-update check operation is stopped", ca.CaCertName)
				ca.cancel()
				return
			}
			_, forceUpdate, err = checkCertValidity(ca.CaCertName)
			if err != nil {
				hwlog.RunLog.Errorf("check cert [%v] validity failed: %v", ca.CaCertName, err)
				continue
			}
			if !forceUpdate {
				continue
			}
			if err = ca.DoForceUpdate(); err != nil {
				hwlog.RunLog.Errorf("do cert [%v] force update operation failed: %v", ca.CaCertName, err)
			}
			hwlog.RunLog.Infof("cert [%v] force update operation success", ca.CaCertName)
			return
		}
	}
}

// DoForceUpdate update root ca hub_svr when it's remaining valid time is <= 24h, ignore edge nodes update operation.
func (ca *EdgeCaCertUpdater) DoForceUpdate() error {
	defer ca.cancel()
	newCaCertContent, err := execForceUpdate(ca.CaCertName, CertTypeEdgeCa)
	if err != nil {
		hwlog.RunLog.Errorf("do force update process for ca cert [%v] failed: %v", ca.CaCertName, err)
		return fmt.Errorf("do force update process for ca cert [%v] failed: %v", ca.CaCertName, err)
	}
	ca.TempCaCertContent = newCaCertContent
	hwlog.RunLog.Infof("do force update for ca cert [%v] success", ca.CaCertName)
	return nil
}

// DoForceUpdate update root ca hub_client when it's remaining valid time is <= 24h, ignore edge nodes update operation.
func (svc *EdgeSvcCertUpdater) DoForceUpdate() error {
	defer svc.cancel()
	newCaCertContent, err := execForceUpdate(svc.CaCertName, CertTypeEdgeSvc)
	if err != nil {
		hwlog.RunLog.Errorf("do force update process for ca cert [%v] failed: %v", svc.CaCertName, err)
		return fmt.Errorf("do force update process for ca cert [%v] failed: %v", svc.CaCertName, err)
	}
	svc.TempCaCertContent = newCaCertContent
	hwlog.RunLog.Infof("do force update for ca cert [%v] success", svc.CaCertName)
	return nil
}

// process for update operation: create new ca, send update notify, update old ca cert.
func execForceUpdate(caCertName, caCertType string) (string, error) {
	tempCaCertContent, err := CreateTempCaCert(caCertName)
	if err != nil {
		hwlog.RunLog.Errorf("create or load temp ca cert [%v] failed: %v", caCertName, err)
		return "", fmt.Errorf("create or load temp ca cert [%v] failed: %v", caCertName, err)
	}
	payload := CertUpdatePayload{
		CertType:    caCertType,
		ForceUpdate: true,
		CaContent:   tempCaCertContent,
	}
	if err = sendCertUpdateNotify(payload); err != nil {
		hwlog.RunLog.Errorf("send cert [%v] force update notify failed: %v", caCertName, err)
		return "", fmt.Errorf("send cert [%v] force update notify failed: %v", caCertName, err)
	}
	if err = UpdateCaCertWithTemp(caCertName); err != nil {
		hwlog.RunLog.Errorf("update local ca cert failed: %v", err)
		return "", fmt.Errorf("update local ca cert failed: %v", err)
	}
	if err = RemoveTempCaCert(caCertName); err != nil {
		hwlog.RunLog.Warnf("remove temporary root cert failed: %v", err)
	}
	return tempCaCertContent, nil
}

// check if the root ca cert need to be updated
// 1st return true means it need to be updated in normal way.
// 2nd return true means it need to be force updated as soon as possible
func checkCertValidity(certName string) (bool, bool, error) {
	certData, err := utils.LoadFile(getRootCaPath(certName))
	if err != nil {
		hwlog.RunLog.Errorf("load cert file [%v] error: %v", getRootCaPath(certName), err)
		return false, false, err
	}
	// skip update operation when no cert file found
	if certData == nil {
		hwlog.RunLog.Warnf("cert file [%v] not exist or content is empty, skip cert update", getRootCaPath(certName))
		return false, false, nil
	}
	crt, err := x509.LoadCertsFromPEM(certData)
	if err != nil {
		hwlog.RunLog.Errorf("parse pem cert file error: %v", err)
		return false, false, err
	}
	remainingValidDays, err := x509.GetValidityPeriod(crt)
	if err != nil {
		hwlog.RunLog.Errorf("check cert validity period error: %v", err)
		return false, false, err
	}
	preExpiredDays := config.GetCertConfig().CertExpireTime
	if remainingValidDays > float64(preExpiredDays) {
		return false, false, nil
	}
	// the cert validity period is <= 24h or already expired.
	if remainingValidDays <= 1 {
		return true, true, nil
	}
	return true, false, nil
}

func sendCertUpdateNotify(payload CertUpdatePayload) error {
	tls := certutils.TlsCertInfo{
		RootCaPath: util.RootCaPath,
		CertPath:   util.ServerCertPath,
		KeyPath:    util.ServerKeyPath,
		SvrFlag:    false,
	}

	url := fmt.Sprintf("https://%s:%d%s", common.EdgeMgrDns, common.EdgeMgrPort, common.ResCertUpdate)
	httpsReq := httpsmgr.GetHttpsReq(url, tls)
	postJsonData, err := json.Marshal(payload)
	if err != nil {
		hwlog.RunLog.Errorf("serialize cert update payload data error: %v", err)
		return err
	}
	respBytes, err := httpsReq.PostJson(postJsonData)
	if err != nil {
		hwlog.RunLog.Errorf("do http post request error: %v", err)
		return err
	}
	var resp common.RespMsg
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		hwlog.RunLog.Errorf("deserialize http response content error: %v", err)
		return err
	}
	status := resp.Status
	if status != common.Success {
		err = fmt.Errorf("cert update operation failed, result status:%s, msg:%s", status, resp.Msg)
		hwlog.RunLog.Error(err)
		return err
	}
	return nil
}
