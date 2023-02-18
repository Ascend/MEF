// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller upgrade handler
package edgeinstaller

import (
	"encoding/json"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/modulemanager/model"
)

const (
	retryTime = 30
	waitTime  = 5 * time.Second
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
			RootCaPath:    util.RootCaPath,
			CertPath:      util.ServerCertPath,
			KeyPath:       util.ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	var rootCaRes string
	var err error
	for i := 0; i < retryTime; i++ {
		rootCaRes, err = reqCertParams.GetRootCa(certName)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	address, err := util.GetImageAddress()
	if err != nil {
		hwlog.RunLog.Errorf("get image registry address failed, %v", err)
		return err
	}
	res := certutils.QueryCertRes{
		CertName: certName,
		Cert:     rootCaRes,
		Address:  address,
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
