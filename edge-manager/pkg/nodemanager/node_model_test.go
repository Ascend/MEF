// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
// Package nodemanager for
package nodemanager

import (
	"crypto/rand"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/kubeclient"
)

const (
	idUpperLimit = 1000
	retryTimes   = 3
)

func TestDeleteRelation(t *testing.T) {
	client := kubeclient.GetKubeClient()
	patch := gomonkey.ApplyPrivateMethod(client, "DeleteNodeLabels", func(a string, b []string) (*v1.Node, error) {
		return nil, nil
	})
	defer patch.Reset()
	var nodeId uint64
	for i := 0; i < retryTimes; i++ {
		id, err := randIntn(idUpperLimit)
		if err != nil {
			hwlog.RunLog.Error(err)
			return

		}
		nodeId = uint64(id)
		_, err = NodeServiceInstance().getNodeByID(nodeId)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			break
		}
	}
	convey.Convey("test delete node relation", t, func() {
		newNode := &NodeInfo{
			ID: nodeId, NodeName: "NodeName", UniqueName: "SerialNumber", SerialNumber: "SerialNumber",
			IsManaged: false, SoftwareInfo: "", CreatedAt: time.Now().String(), UpdatedAt: time.Now().String(),
		}
		err := NodeServiceInstance().createNode(newNode)
		convey.So(err, convey.ShouldBeNil)
		gid, err := randIntn(idUpperLimit)
		convey.So(err, convey.ShouldBeNil)
		groupId := uint64(gid)
		newNodeGroup := &NodeGroup{
			ID:        groupId,
			GroupName: "GroupName",
			CreatedAt: time.Now().String(),
			UpdatedAt: time.Now().String(),
		}
		err = NodeServiceInstance().createNodeGroup(newNodeGroup)
		convey.So(err, convey.ShouldBeNil)
		relation := NodeRelation{
			GroupID:   groupId,
			NodeID:    nodeId,
			CreatedAt: time.Now().String(),
		}
		err = test.MockGetDb().Model(NodeRelation{}).Create(relation).Error
		convey.So(err, convey.ShouldBeNil)
		err = deleteRelation(test.MockGetDb(), groupId, nodeId)
		convey.So(err, convey.ShouldBeNil)
		test.MockGetDb().Model(NodeRelation{}).Where(NodeRelation{NodeID: nodeId}).Delete(relation)
		test.MockGetDb().Model(NodeInfo{}).Where(NodeInfo{ID: nodeId}).Delete(&NodeInfo{})
		test.MockGetDb().Model(NodeGroup{}).Where(NodeGroup{ID: groupId}).Delete(&NodeGroup{})
	})
}

func randIntn(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -1, err
	}
	randNum := int((*n).Int64())
	return randNum, nil
}
