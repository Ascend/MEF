// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager for package main test
package nodemanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/kubeclient"
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcBaseWithDb := &test.TcBaseWithDb{
		DbPath: ":memory:?cache=shared",
		Tables: append(tables, &NodeInfo{}, &NodeRelation{}, &NodeGroup{}),
	}

	env = environment{}
	service := &nodeSyncImpl{}
	client := &kubeclient.Client{}
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFuncReturn(NodeSyncInstance, service).
		ApplyMethodReturn(service, "ListMEFNodeStatus", map[string]string{}).
		ApplyMethodReturn(service, "GetMEFNodeStatus", statusOffline, nil).
		ApplyMethodReturn(service, "GetK8sNodeStatus", statusOffline, nil).
		ApplyMethodReturn(service, "GetAllocatableResource", &NodeResource{}, nil).
		ApplyMethodReturn(service, "GetAvailableResource", &NodeResource{}, nil).
		ApplyFuncReturn(kubeclient.GetKubeClient, client).
		ApplyPrivateMethod(client, "patchNode", func(_ *kubeclient.Client) (*v1.Node, error) { return &v1.Node{}, nil }).
		ApplyMethodReturn(client, "ListNode", &v1.NodeList{}, nil).
		ApplyMethodReturn(client, "DeleteNode", nil)

	test.RunWithPatches(tcBaseWithDb, m, patches)
}
