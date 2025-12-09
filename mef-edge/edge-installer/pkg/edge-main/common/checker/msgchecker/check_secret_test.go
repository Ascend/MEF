// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package mefmsgchecker
package msgchecker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

func TestSecret(t *testing.T) {
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFunc(configpara.GetPodConfig, MockPodConfig).
		ApplyFuncReturn(configpara.GetNetType, constants.FDWithOM, nil).
		ApplyPrivateMethod(&MsgValidator{}, "checkSystemResources", func() error { return nil })

	defer patches.Reset()

	convey.Convey("test fd secret para", t, func() {
		convey.Convey("test secret para", testSecretParaValidate)
		convey.Convey("test secret data", testSecretDataValidate)
	})

}

var secretData = `
{
    "data":{
        ".dockerconfigjson":"MTIzNA=="
    },
    "kind":"Secret",
    "metadata":{
        "creationTimestamp":"2024-01-29T14:31:19Z",
        "name":"fusion-director-docker-registry-secret",
        "namespace":"websocket",
        "resourceVersion":"0902527",
        "uid":"fusion-director-docker-registry-secret"
    },
    "type":"kubernetes.io/dockerconfigjson"
}`

func getBaseSecretInfo() types.Secret {
	var secret types.Secret
	err := json.Unmarshal([]byte(secretData), &secret)
	if err != nil {
		hwlog.RunLog.Infof("unmarshal config map data failed:%v", err)
	}

	return secret
}

func setSecretMsg(msg *model.Message, secret types.Secret) {
	data, err := json.Marshal(secret)
	if err != nil {
		fmt.Printf("marshal secret failed:%v", err)
		return
	}
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "controller",
		Group:     "resource",
		Operation: "update",
		Resource:  "websocket/secret/fusion-director-docker-registry-secret",
	}

	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1678505303009
	msg.Header.ResourceVersion = "3558793"
	msg.Header.Sync = true

	msg.FillContent(data)
}

type secretTestCase struct {
	description string
	secret      types.Secret
	shouldErr   bool
	assert      convey.Assertion
	expected    interface{}
}

func getSecretTestCase() []secretTestCase {
	return []secretTestCase{
		{
			description: "test secret name contain invalid char",
			secret:      types.Secret{ObjectMeta: types.ObjectMeta{Name: "test-"}},
			shouldErr:   true,
			assert:      convey.ShouldContainSubstring, expected: "Secret.ObjectMeta.Name",
		}, {
			description: "test secret uid is fusion-director-docker-registry-secret",
			secret: types.Secret{ObjectMeta: types.ObjectMeta{Name: "image-pull-secret",
				UID: "fusion-director-docker-registry-secret"}},
			shouldErr: false,
			assert:    convey.ShouldEqual,
			expected:  nil,
		}, {
			description: "test secret uid is uuid format",
			secret: types.Secret{ObjectMeta: types.ObjectMeta{Name: "image-pull-secret",
				UID: "40550ae7-2b7a-4280-9277-df2f2e871ce0"}},
			shouldErr: false,
			assert:    convey.ShouldEqual, expected: nil,
		}, {
			description: "test secret uid invalid",
			secret: types.Secret{ObjectMeta: types.ObjectMeta{Name: "image-pull-secret",
				UID: "image-pull-secret"}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring, expected: "Secret.ObjectMeta.UID",
		}, {
			description: "test secret ResourceVersion invalid",
			secret: types.Secret{ObjectMeta: types.ObjectMeta{Name: "image-pull-secret",
				UID: "image-pull-secret", ResourceVersion: "v1.1"}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring, expected: "Secret.ObjectMeta.ResourceVersion",
		}, {
			description: "test secret ResourceVersion invalid",
			secret: types.Secret{ObjectMeta: types.ObjectMeta{Name: "image-pull-secret",
				UID: "40550ae7-2b7a-4280-9277-df2f2e871ce0", ResourceVersion: "v1.1"}},
			shouldErr: true,
			assert:    convey.ShouldContainSubstring, expected: "Secret.ObjectMeta.ResourceVersion",
		},
	}
}
func testSecretParaValidate() {
	var secret types.Secret
	secret = getBaseSecretInfo()
	var testCase = getSecretTestCase()

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())

	for _, tc := range testCase {
		hwlog.RunLog.Infof("--------------------%s-------------------", tc.description)

		secret.UID = tc.secret.UID
		secret.Name = tc.secret.Name
		secret.ResourceVersion = tc.secret.ResourceVersion

		var msg model.Message
		setSecretMsg(&msg, secret)
		var err error
		if err = msgValidator.Check(&msg); err != nil {
			hwlog.RunLog.Errorf("check msg failed: %v", err)
		}

		if tc.shouldErr {
			convey.So(err.Error(), tc.assert, tc.expected)
		} else {
			convey.So(err, tc.assert, tc.expected)
		}
	}
}

func testSecretDataValidate() {
	var msg model.Message
	var secret types.Secret
	secret = getBaseSecretInfo()
	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())
	var err error

	// 检查 secret Data map 长度超过1
	secret.Data["test"] = []byte("1234")
	setSecretMsg(&msg, secret)
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err.Error(), convey.ShouldContainSubstring, "validation for 'Data' failed on the 'max' tag")
}
