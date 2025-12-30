// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msgconv
package msgconv

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"gorm.io/gorm"
	"k8s.io/api/core/v1"

	dbcomm "huawei.com/mindx/common/database"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/database"
)

func TestMain(m *testing.M) {
	if err := Init(); err != nil {
		panic(err)
	}
	tcModule := &test.TcBaseWithDb{Tables: []interface{}{database.Meta{}}}
	test.RunWithPatches(tcModule, m, gomonkey.ApplyFunc(dbcomm.GetDb, test.MockGetDb))
}

func ensureNodeExists(nodeName string, node v1.Node) error {
	dataBytes, err := json.Marshal(node)
	if err != nil {
		return err
	}
	stmt := test.MockGetDb().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(database.Meta{})
	if stmt.Error != nil {
		return stmt.Error
	}
	stmt = test.MockGetDb().Create(database.Meta{
		Key:   constants.ActionDefaultNodeStatus + nodeName,
		Type:  constants.ResourceTypeNode,
		Value: string(dataBytes),
	})
	return stmt.Error
}

type messageHeader struct {
	ID       string
	ParentID string
	Sync     bool
}

func createMsg(header messageHeader, route model.MessageRoute, content interface{}) (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return nil, err
	}

	msg.Header.ID = msg.GetId()
	msg.Header.ParentID = header.ParentID
	msg.Header.Sync = header.Sync
	msg.KubeEdgeRouter = route
	if str, ok := content.(string); ok {
		content = string(model.FormatMsg([]byte(str)))
	}
	if content == nil {
		content = "null"
	}
	if err := msg.FillContent(content); err != nil {
		return nil, err
	}
	return msg, nil
}

func mustCreateMsg(header messageHeader, route model.MessageRoute, content interface{}) *model.Message {
	msg, err := createMsg(header, route, content)
	if err != nil {
		panic(err)
	}
	return msg
}

func mustParseTime(layout, timeStr string) time.Time {
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

func convertToEdgeCoreMsg(input *model.Message, outputContent interface{}) error {
	return convertMsg(input, outputContent, Cloud)
}

func convertToFdMsg(input *model.Message, outputContent interface{}) error {
	return convertMsg(input, outputContent, Edge)
}

func convertToCloudcoreMsg(input *model.Message, outputContent interface{}) error {
	return convertMsg(input, outputContent, Edge)
}

func convertMsg(input *model.Message, outputContent interface{}, source Source) error {
	proxy := &Proxy{MessageSource: source, DispatchFunc: dummyDispatchFunc}
	if err := proxy.DispatchMessage(input); err != nil {
		return err
	}

	if outputContent == nil {
		return nil
	}
	return input.ParseContent(outputContent)
}

func dummyDispatchFunc(_ *model.Message) error {
	return nil
}
