// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
// Package configmanager for

package configmanager

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
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

func TestCheckAndUpdateTokenForExpireErr(t *testing.T) {
	convey.Convey("test checkAndUpdateToken for Expire err", t, func() {
		//	ensure existing token deleted
		err := ConfigRepositoryInstance().revokeToken()
		convey.So(err, convey.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(ConfigRepositoryInstance().ifTokenExpire, false, "false")
		defer patch.Reset()
		checkAndUpdateToken()
		_, _, err = ConfigRepositoryInstance().GetToken()
		convey.So(errors.Is(err, gorm.ErrRecordNotFound), convey.ShouldBeTrue)
	})
}

func TestCheckAndUpdateTokenForRevokeTokenErr(t *testing.T) {
	convey.Convey("test checkAndUpdateToken for RevokeToken err", t, func() {
		//	ensure existing token deleted
		err := ConfigRepositoryInstance().revokeToken()
		convey.So(err, convey.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(ConfigRepositoryInstance().ifTokenExpire, true, nil)
		defer patch.Reset()
		checkAndUpdateToken()
		_, _, err = ConfigRepositoryInstance().GetToken()
		convey.So(err, convey.ShouldResemble, errors.New("record not found"))
	})
}

func TestExportToken(t *testing.T) {
	convey.Convey("test ExportToken", t, func() {
		msg := model.Message{}
		err := msg.FillContent(struct{}{})
		convey.So(err, convey.ShouldBeNil)
		resp := exportToken(&msg)
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
		msg := model.Message{}
		err = msg.FillContent(bytes)
		convey.So(err, convey.ShouldBeNil)
		resp := updateConfig(&msg)
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
		msg := model.Message{}
		err = msg.FillContent(configStr)
		convey.So(err, convey.ShouldBeNil)
		resp := downloadConfig(&msg)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
	})
}
