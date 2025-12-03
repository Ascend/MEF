// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager for node_informer test
package nodemanager

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

func TestGetMEFNodeStatusForOffline(t *testing.T) {
	convey.Convey("test GetMEFNodeStatus For Get Lable Err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		str, err := service.GetMEFNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetMEFNodeStatusForGetNodeErr(t *testing.T) {
	convey.Convey("test GetMEFNodeStatus For Get Node Err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, "err")
		defer patch.Reset()
		str, err := service.GetMEFNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetK8sNodeStatus(t *testing.T) {
	convey.Convey("test GetK8sNodeStatus For offline", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		str, err := service.GetK8sNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
		convey.So(str, convey.ShouldEqual, "offline")
	})
}

func TestGetAllocatableResource(t *testing.T) {
	convey.Convey("test GetAllocatableResource For err", t, func() {
		hostname := "local"
		service := &nodeSyncImpl{}
		patch := gomonkey.ApplyFuncReturn(service.getNode, nil, nil)
		defer patch.Reset()
		_, err := service.GetK8sNodeStatus(hostname)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestGetAvailableResource test get available resource
func TestGetAvailableResource(t *testing.T) {
	convey.Convey("test get available resource success", t, func() {
		convey.Convey("When calling GetAvailableResource with valid nodeID and hostname", func() {
			s := &nodeSyncImpl{}

			patches := gomonkey.ApplyPrivateMethod(&NodeServiceImpl{}, "getGroupsByNodeID",
				func(nodeID uint64) (*[]NodeGroup, error) {
					groups := []NodeGroup{
						{
							ResourcesRequest: `{"Cpu":"2", "Memory":"4Gi", "Npu":"1"}`,
						},
					}
					return &groups, nil
				},
			)

			patches.ApplyMethodFunc(
				&nodeSyncImpl{}, "GetAllocatableResource",
				func(hostname string) (*NodeResource, error) {
					return &NodeResource{
						Cpu:    resource.MustParse("10"),
						Memory: resource.MustParse("20Gi"),
						Npu:    resource.MustParse("5"),
					}, nil
				},
			)
			defer patches.Reset()
			nodeID := uint64(1)
			hostname := "test-host"
			result, err := s.GetAvailableResource(nodeID, hostname)

			convey.Convey("Then it should return the correct available resources", func() {
				convey.So(err, convey.ShouldBeNil)
				convey.So(result, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestGetEdgeConnStatus test get EdgeConnStatus
func TestGetEdgeConnStatus(t *testing.T) {
	convey.Convey("test GetEdgeConnStatus success", t, testGetEdgeConnStatusSuccess)
	convey.Convey("test GetEdgeConnStatus failed", t, testGetEdgeConnStatusFailed)
}

func testGetEdgeConnStatusFailed() {
	snList := []string{"sn1", "sn2", "sn3"}

	convey.Convey("When creating a new message fails", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return nil, errors.New("failed to create message")
		})
		defer patches.Reset()

		result := getEdgeConnStatus(snList...)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result["sn1"], convey.ShouldBeFalse)
	})

	convey.Convey("When sending a sync message fails", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendSyncMessage, func(msg *model.Message, timeout time.Duration) (*model.Message, error) {
			return nil, errors.New("failed to send sync message")
		})
		defer patches.Reset()

		result := getEdgeConnStatus(snList...)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result["sn2"], convey.ShouldBeFalse)
	})

	convey.Convey("When parsing the response content fails", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendSyncMessage, func(msg *model.Message, timeout time.Duration) (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyMethod(reflect.TypeOf(&model.Message{}), "ParseContent", func(_ *model.Message, v interface{}) error {
			return errors.New("parse content failed")
		})
		defer patches.Reset()

		result := getEdgeConnStatus(snList...)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result["sn3"], convey.ShouldBeFalse)
	})
}

func testGetEdgeConnStatusSuccess() {
	snList := []string{"sn1", "sn2", "sn3"}

	convey.Convey("When all operations succeed", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendSyncMessage, func(msg *model.Message, timeout time.Duration) (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyMethod(reflect.TypeOf(&model.Message{}), "ParseContent", func(_ *model.Message, v interface{}) error {
			peers := []websocketmgr.WebsocketPeerInfo{
				{Sn: "sn1"},
				{Sn: "sn2"},
			}
			reflect.ValueOf(v).Elem().Set(reflect.ValueOf(peers))
			return nil
		})
		defer patches.Reset()

		result := getEdgeConnStatus(snList...)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result["sn1"], convey.ShouldBeTrue)
		convey.So(result["sn2"], convey.ShouldBeTrue)
		convey.So(result["sn3"], convey.ShouldBeFalse)
	})
}

// TestReportChangedNodeInfo test report changed node add and delete cases
func TestReportChangedNodeInfo(t *testing.T) {
	convey.Convey("Given a node with a valid serial number", t, func() {
		node := mockNode()
		testRptChangeNodeActionAdd(node)
		testRptChangeNodeActionDel(node)
		testRptChangeNodeActionInvalid(node)
	})

	convey.Convey("Given a node without a serial number", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				Labels: map[string]string{
					"other-label": "value",
				},
			},
		}

		convey.Convey("When the action is 'add'", func() {
			reportChangedNodeInfo(nodeActionAdd, node)
		})
	})
}

func testRptChangeNodeActionAdd(node *v1.Node) {
	convey.Convey("When the action is 'add'", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendMessage, func(msg *model.Message) error {
			return nil
		})
		defer patches.Reset()

		reportChangedNodeInfo(nodeActionAdd, node)
	})

	convey.Convey("When creating a new message fails", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return nil, errors.New("failed to create message")
		})
		defer patches.Reset()

		reportChangedNodeInfo(nodeActionAdd, node)
	})
}

func testRptChangeNodeActionDel(node *v1.Node) {
	convey.Convey("When the action is 'delete'", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendMessage, func(msg *model.Message) error {
			return nil
		})
		defer patches.Reset()

		reportChangedNodeInfo(nodeActionDelete, node)
	})

	convey.Convey("When sending the message fails", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendMessage, func(msg *model.Message) error {
			return errors.New("failed to send message")
		})
		defer patches.Reset()

		reportChangedNodeInfo(nodeActionDelete, node)
	})
}

func testRptChangeNodeActionInvalid(node *v1.Node) {
	convey.Convey("When the action is invalid", func() {
		patches := gomonkey.ApplyFunc(model.NewMessage, func() (*model.Message, error) {
			return &model.Message{}, nil
		})
		patches.ApplyFunc(modulemgr.SendMessage, func(msg *model.Message) error {
			return nil
		})
		defer patches.Reset()

		reportChangedNodeInfo("invalid-action", node)
	})
}

// TestListMEFNodeStatus test ListMEFNodeStatus with nodes and empty cases
func TestListMEFNodeStatus(t *testing.T) {
	convey.Convey("Given a nodeSyncImpl instance with a list of nodes", t, testGetEdgeConnStatusNormal)
	convey.Convey("Given a nodeSyncImpl instance with empty nodes", t, testGetEdgeConnStatusEmpty)
}

func testGetEdgeConnStatusNormal() {
	var s nodeSyncImpl
	s.informer = informers.NewSharedInformerFactory(&kubernetes.Clientset{}, time.Second).Core().V1().Pods().Informer()
	mockObjects := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				snNodeLabelKey: "sn1",
			},
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	convey.Convey("When the informer returns a list of nodes", func() {
		patches := gomonkey.ApplyMethodReturn(s.informer.GetStore(), "List", []interface{}{mockObjects}).
			ApplyFunc(getEdgeConnStatus, func(serialNumbers ...string) map[string]bool {
				return map[string]bool{
					"sn1": true,
				}
			})
		defer patches.Reset()

		result := s.ListMEFNodeStatus()
		convey.So(result["node1"], convey.ShouldBeEmpty)
	})
}

func testGetEdgeConnStatusEmpty() {
	var s nodeSyncImpl
	s.informer = informers.NewSharedInformerFactory(&kubernetes.Clientset{}, time.Second).Core().V1().Pods().Informer()
	convey.Convey("When the informer returns an empty list", func() {
		patches := gomonkey.ApplyMethodReturn(s.informer.GetStore(), "List", []interface{}{})
		defer patches.Reset()

		result := s.ListMEFNodeStatus()
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(len(result), convey.ShouldEqual, 0)
	})
}

// TestHandleAddNode test add node in success and fail cases
func TestHandleAddNode(t *testing.T) {
	convey.Convey("test handle add node success", t, testHandleAddNodeSuccess)
	convey.Convey("test handle add node failed", t, testHandleAddNodeFailed)
}

func testHandleAddNodeSuccess() {
	s := &nodeSyncImpl{}
	var c *checker.AndChecker

	convey.Convey("When the node info check passes and createNode succeeds", func() {
		patches := gomonkey.ApplyMethodReturn(c, "Check", checker.CheckResult{
			Result: true,
		})
		patches.ApplyFuncReturn(GetTableCount, 0, nil)
		patches.ApplyFunc(NodeServiceInstance().createNode, func(nodeInfo *NodeInfo) error {
			return nil
		})
		defer patches.Reset()

		node := mockNode()

		s.handleAddNode(node)
	})
}

func testHandleAddNodeFailed() {
	s := &nodeSyncImpl{}
	var c *checker.AndChecker

	convey.Convey("When the node info check fails", func() {
		patches := gomonkey.ApplyMethodReturn(c, "Check", checker.CheckResult{
			Result: false,
			Reason: "invalid node info",
		})
		defer patches.Reset()

		s.handleAddNode(&v1.Node{})
	})

	convey.Convey("When the node count exceeds the maximum", func() {
		patches := gomonkey.ApplyMethodReturn(c, "Check", checker.CheckResult{
			Result: true,
		})
		patches.ApplyFuncReturn(GetTableCount, maxNodeSize, nil)
		defer patches.Reset()

		s.handleAddNode(&v1.Node{})
	})

	convey.Convey("When the node info check passes and createNode fails", func() {
		patches := gomonkey.ApplyMethodReturn(c, "Check", checker.CheckResult{
			Result: true,
		})
		patches.ApplyFuncReturn(GetTableCount, 0, nil)
		patches.ApplyFunc(NodeServiceInstance().createNode, func(nodeInfo *NodeInfo) error {
			return errors.New("failed to create node")
		})
		defer patches.Reset()

		s.handleAddNode(&v1.Node{})
	})
}

// TestGetMEFNodeStatus get MEF node status for success and fail cases
func TestGetMEFNodeStatus(t *testing.T) {
	convey.Convey("get MEF node status success", t, testGetMEFNodeStatusSuccess)
	convey.Convey("get MEF node status failed", t, testGetMEFNodeStatusFailed)
}

func testGetMEFNodeStatusSuccess() {
	convey.Convey("When calling GetMEFNodeStatus with a valid hostname", func() {
		s := &nodeSyncImpl{}
		patches := gomonkey.ApplyPrivateMethod(&nodeSyncImpl{}, "getNode",
			func(hostname string) (*v1.Node, error) {
				return &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							snNodeLabelKey: "12345",
						},
					},
				}, nil
			},
		)
		defer patches.Reset()
		hostname := "test-host"
		_, err := s.GetMEFNodeStatus(hostname)

		convey.So(err, convey.ShouldBeNil)
	})
}

func testGetMEFNodeStatusFailed() {
	s := &nodeSyncImpl{}
	convey.Convey("When the node does not have a serial number label", func() {
		patches := gomonkey.ApplyPrivateMethod(&nodeSyncImpl{}, "getNode",
			func(hostname string) (*v1.Node, error) {
				return &v1.Node{}, nil
			},
		)
		defer patches.Reset()
		status, err := s.GetMEFNodeStatus("test-host")

		convey.So(err, convey.ShouldBeNil)
		convey.So(status, convey.ShouldEqual, statusOffline)
	})
}

// TestTransferNodeType test transfer node type for success and fail cases
func TestTransferNodeType(t *testing.T) {
	convey.Convey("test transferNodeType with right input", t, testTransferNodeTypeSuccess)
	convey.Convey("test transferNodeType with wrong input", t, testTransferNodeTypeFailed)
}

func testTransferNodeTypeSuccess() {
	convey.Convey("When the input is a *v1.Node without master label", func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"other-label": "value",
				},
			},
		}
		result := transferNodeType(node)

		convey.Convey("Then it should return the node", func() {
			convey.So(result, convey.ShouldEqual, node)
		})
	})

	convey.Convey("When the input is a *v1.Node with master label", func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					masterNodeLabelKey: "true",
				},
			},
		}
		result := transferNodeType(node)

		convey.So(result, convey.ShouldBeNil)
	})
}

func testTransferNodeTypeFailed() {
	convey.Convey("When the input is nil", func() {
		result := transferNodeType(nil)

		convey.So(result, convey.ShouldBeNil)
	})

	convey.Convey("When the input is not a *v1.Node", func() {
		result := transferNodeType("not a node")

		convey.So(result, convey.ShouldBeNil)
	})
}

// TestEvalNodeStatus test eval node status for success and fail cases
func TestEvalNodeStatus(t *testing.T) {
	convey.Convey("test eval node status ready", t, testEvalNodeStatusReady)
	convey.Convey("test eval node status not ready", t, testEvalNodeStatusNotReady)
	convey.Convey("test eval node status unknown", t, testEvalNodeStatusUnknown)
	convey.Convey("test eval node with wrong condition", t, testEvalNodeWrongCon)
	convey.Convey("test eval node with no condition", t, testEvalNodeNoCon)
}

func testEvalNodeStatusReady() {
	convey.Convey("When the node is ready", func() {
		node := &v1.Node{
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		status := evalNodeStatus(node)
		convey.So(status, convey.ShouldEqual, statusReady)
	})
}

func testEvalNodeStatusNotReady() {
	convey.Convey("When the node is not ready", func() {
		node := &v1.Node{
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionFalse,
					},
				},
			},
		}

		status := evalNodeStatus(node)
		convey.So(status, convey.ShouldEqual, statusNotReady)
	})
}

func testEvalNodeStatusUnknown() {
	convey.Convey("When the node status is unknown", func() {
		node := &v1.Node{
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionUnknown,
					},
				},
			},
		}

		status := evalNodeStatus(node)
		convey.So(status, convey.ShouldEqual, statusUnknown)
	})
}

func testEvalNodeWrongCon() {
	convey.Convey("When the node has no ready condition", func() {
		node := &v1.Node{
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeMemoryPressure,
						Status: v1.ConditionTrue,
					},
				},
			},
		}

		status := evalNodeStatus(node)
		convey.So(status, convey.ShouldEqual, statusOffline)
	})
}

func testEvalNodeNoCon() {
	convey.Convey("When the node has no conditions", func() {
		node := &v1.Node{
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{},
			},
		}

		status := evalNodeStatus(node)
		convey.So(status, convey.ShouldEqual, statusOffline)
	})
}

func mockNode() *v1.Node {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				snNodeLabelKey: "test-sn",
			},
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{
					Type:    v1.NodeInternalIP,
					Address: "192.168.1.1",
				},
			},
		},
	}
	return node
}
