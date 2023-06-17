// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller upgrade handler
package edgeinstaller

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

type certHandler struct{}

// Handle certHandler handle entry
func (ch *certHandler) Handle(message *model.Message) error {
	hwlog.RunLog.Info("----------downloading cert content begin----------")
	hwlog.RunLog.Info("edge-installer received message from edge-connector success")
	certName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("convert to cert name failed")
		return errors.New("convert to cert name failed")
	}
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	rootCaRes, err := reqCertParams.GetRootCa(certName)
	if err != nil {
		hwlog.RunLog.Errorf("query cert content from cert-manager failed, error: %v", err)
		return err
	}
	res := certutils.ClientCertResp{
		CertName:    certName,
		CertContent: rootCaRes,
		CertOpt:     common.Update,
	}
	if certName == common.ImageCertName {
		address, err := util.GetImageAddress()
		if err != nil {
			hwlog.RunLog.Errorf("get image registry address failed, error:%v", err)
			return err
		}
		if address == "" {
			hwlog.RunLog.Warn("image registry address should be configured")
			return nil
		}
		res.ImageAddress = address
	}
	data, err := json.Marshal(res)
	if err != nil {
		hwlog.RunLog.Errorf("marshal cert response failed, error: %v", err)
		return errors.New("marshal cert response failed")
	}
	if err = sendMessage(message, string(data)); err != nil {
		hwlog.RunLog.Errorf("edge-installer send message to edge-connector for cert failed, error: %v", err)
		return errors.New("edge-installer send message to edge-connector for cert failed")
	}
	return nil
}
