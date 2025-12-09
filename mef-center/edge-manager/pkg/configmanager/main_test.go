// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configmanager for package main test
package configmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"gorm.io/gorm"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/config"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

const (
	tokenDbName     = "./test-token.db"
	tokenExpireDays = 7
	// an expired cert in base64
	base64CertContent = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVzekNDQXh1Z0F3SUJBZ0lVQndBVFlEb0ZTdE9GRW1VM2I1RmVDUl" +
		"FCMTRZd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1V6RUxNQWtHQTFVRUJoTUNRMDR4RURBT0JnTlZCQWdNQjBOb1pXNW5aSFV4RURBT0Jn" +
		"TlZCQWNNQjBObwpaVzVuWkhVeER6QU5CZ05WQkFvTUJraDFZWGRsYVRFUE1BMEdBMVVFQ3d3R1NIVmhkMlZwTUI0WERUSXpNRFF3Ck" +
		"16QTROREUxT0ZvWERUTXpNRE16TVRBNE5ERTFPRm93VXpFTE1Ba0dBMVVFQmhNQ1EwNHhFREFPQmdOVkJBZ00KQjBOb1pXNW5aSFV4" +
		"RURBT0JnTlZCQWNNQjBOb1pXNW5aSFV4RHpBTkJnTlZCQW9NQmtoMVlYZGxhVEVQTUEwRwpBMVVFQ3d3R1NIVmhkMlZwTUlJQm9qQU" +
		"5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FZOEFNSUlCaWdLQ0FZRUE3bmdXCm16WFJCOFBFVkM5R0Rmb1JuOVM5OWtBWkdxVUFuWjdodDJ1" +
		"M21VVDZjQnFkWGZ4V1FGOUFocERTZ0M4aGNxbWoKelhyYk1rUUxNV2pWNFBLYzNWTktlbVBiSTIwRlBCOFp5TXc4MUdGTGJGYzFQRH" +
		"pYcmdrWGhvaFo4NTVMS2kwQgptZ3NzYWZjcno5cU00ckxzVWVPTGQ4SkpFTVltVDRRZ2hqeHUxSjk0TVpDUExVbHVJNnUxejNDUVYw" +
		"dnF6UGRpClUvVU9aWm5keG4wVmd3ajdJQ3JFbjg2eUZ2bFFjRmFaMEFFenU3MWp4RFRmUno1OWQwWDZTZDNCV1BWQ2I4ZHYKQndlcj" +
		"VFZHNjOTZXbnVHd2RNZDFhZmZJOEQ3bFN0elJ0Tk5HRUk2NU5UbXBSREk4QUhuVEQ1RTMvRHBOSkVCTgpjZ0dlWFVqckYvVXdtTTV5" +
		"M08yUzlQSWZkZmQ5MGZtY040ajEzV0hqOTBWb3l5OFZPYVRnc0lzVGM5d1EzcndTCmR1SVVXRWlnVFlYYVBkMCsyeWRNY3VzUi9iYT" +
		"RpOEFCd3U2RmVwZ3dUYzg3ZTRqSFJsVkJSckJwZzlOWjZZQ2wKOGhHRGlGNngyVUtZcDkxSjZEdWZsMGU4K01yK09RczVxWGhya1VM" +
		"aGRCMnlJR1doWlZlRHBYREM1MjVwQWdNQgpBQUdqZnpCOU1CMEdBMVVkRGdRV0JCU0hkZ081RlBwUlBSS3lFQngyQTU5RXRPbms2ek" +
		"FmQmdOVkhTTUVHREFXCmdCU0hkZ081RlBwUlBSS3lFQngyQTU5RXRPbms2ekFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQXNHQTFVZER3" +
		"UUUKQXdJQkJqQWRCZ05WSFNVRUZqQVVCZ2dyQmdFRkJRY0RBUVlJS3dZQkJRVUhBd0l3RFFZSktvWklodmNOQVFFTApCUUFEZ2dHQk" +
		"FGT1BvL3FlVTkvbmZNSTRha0RFcFgraE1aWXNYNTg4VU5DVnB5MnlzRXdBV0hqc3o3TUxpVGRuCjMyb2lpWllUWmZMVlNERHhRR2hO" +
		"Wmg1RXZXTXF5MkR5aXZPdnhCYmZBUjBnRWRFSElWV1BOeGRGeVNMWjNNUDIKVzFNZzNBdE0zMUJJcTVUZkV1VWtxSk5TZjkrbS9GOD" +
		"d2VGJZQnptblY1ak9wWmV0V0UwSHFMc0lObEpUM1NEVwpsVDV5WUFDQ2lvNXUxZGRIQ3RHb2RIK0sxMm8rM2I1WGtERkxhRGJzMjYz" +
		"SitHS1k5K3dRVm1QaVBsU3VibDBZCktjTGZRNjAzd3hWNWpiZ0MvNlREak15b1BMZStLSzhDNjRJWFNOaU1KdDRhWWsyZXg4MndWK1" +
		"R1VXgzOU05MWQKcDU5dlBHNmlpdlNsZmd6cEgxamlwRnowWFhTejNQRlU4cU4xakNmL1VRN01SZFhBS3ozeHgzSzJwMmtkUnY5Kwpv" +
		"eHpsSXZuMDM2c0tQVUU5Nis2UzRtY3grUkpBZDFXL3IvNkNDTDJKWGtWUVNPUFdUQWJ0Y3dvT3lkMDgySXFoCjkvbVZpdVI5bUMzOW" +
		"ZxMUVkdkF4bEMzSGd2U281UmxuNklmYWZYTTNsMlNsVm5TSHhUMWZUaWpnT2xTNE9aNDEKdit3OWlLWkFZdz09Ci0tLS0tRU5EIENF" +
		"UlRJRklDQVRFLS0tLS0="
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcBase := &test.TcBaseWithDb{
		DbPath: tokenDbName,
		Tables: append(tables, &TokenInfo{}),
	}
	patches := gomonkey.ApplyPrivateMethod(ConfigRepositoryInstance(), "db", func() *gorm.DB {
		return test.MockGetDb()
	}).
		ApplyFuncReturn(config.GetAuthConfig, config.AuthInfo{TokenExpireTime: tokenExpireDays}).
		ApplyFuncReturn(common.SendSyncMessageByRestful, common.RespMsg{
			Status: common.Success, Msg: "",
			Data: []nodemanager.NodeInfo{{
				ID: 1,
			}},
		}).
		ApplyFuncReturn(modulemgr.SendMessage, nil).
		ApplyFuncReturn(util.GetImageAddress, "xxxx", nil).
		ApplyFuncReturn(config.GetCertCrlPairCache, config.CertCrlPair{CertPEM: base64CertContent}, nil).
		ApplyMethodReturn(&kubeclient.Client{}, "CreateOrUpdateSecret", nil, nil)

	test.RunWithPatches(tcBase, m, patches)
}
