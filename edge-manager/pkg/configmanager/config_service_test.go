// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
// Package configmanager for

package configmanager

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

func TestGenerateToken(t *testing.T) {
	convey.Convey("test GetToken", t, func() {
		token, err := generateToken()
		convey.So(err, convey.ShouldBeNil)
		hwlog.RunLog.Infof("generated token: %v", token)
		getToken, _, err := ConfigRepositoryInstance().GetToken()
		convey.So(err, convey.ShouldBeNil)
		hwlog.RunLog.Infof("get token: %v", getToken)
		convey.Convey("test ifTokenExpired", func() {
			expired, err := ConfigRepositoryInstance().ifTokenExpire()
			convey.So(err, convey.ShouldBeNil)
			convey.So(expired, convey.ShouldBeFalse)
			err = ConfigRepositoryInstance().revokeToken()
			convey.So(err, convey.ShouldBeNil)
			expired, err = ConfigRepositoryInstance().ifTokenExpire()
			convey.So(err, convey.ShouldBeNil)
			convey.So(expired, convey.ShouldBeFalse)

		})
	})
}

func TestCheckAndUpdateToken(t *testing.T) {
	convey.Convey("test checkAndUpdateToken", t, func() {
		//	ensure existing token deleted
		err := ConfigRepositoryInstance().revokeToken()
		convey.So(err, convey.ShouldBeNil)
		checkAndUpdateToken()
		_, _, err = ConfigRepositoryInstance().GetToken()
		convey.So(errors.Is(err, gorm.ErrRecordNotFound), convey.ShouldBeTrue)
	})
}

func TestExportToken(t *testing.T) {
	convey.Convey("test ExportToken", t, func() {
		resp := exportToken(struct{}{})
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
	})
}

func TestUpdateConfig(t *testing.T) {
	convey.Convey("test UpdateConfig", t, func() {
		certInput := certutils.UpdateClientCert{
			CertContent: []byte(base64CertContent),
			CertName:    common.ImageCertName,
		}
		bytes, err := json.Marshal(certInput)
		convey.So(err, convey.ShouldBeNil)
		resp := updateConfig(string(bytes))
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
		convey.So(resp.Data == common.ImageCertName, convey.ShouldBeTrue)
	})
}

func TestDownloadConfig(t *testing.T) {
	convey.Convey("test downloadConfig", t, func() {
		config := ImageConfig{
			Domain:   "domain",
			IP:       "10.10.10.10",
			Port:     443,
			Account:  "ImageRepository",
			Password: []byte("ImageRepository"),
		}
		configStr, err := json.Marshal(config)
		convey.So(err, convey.ShouldBeNil)
		resp := downloadConfig(string(configStr))
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
	})
}
