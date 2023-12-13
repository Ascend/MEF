// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager for package main test
package nodemanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"

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
