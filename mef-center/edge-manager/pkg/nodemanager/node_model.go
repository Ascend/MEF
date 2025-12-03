// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager to init node service
package nodemanager

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/kubeclient"

	"huawei.com/mindxedge/base/common"
)

var (
	nodeServiceSingleton sync.Once
	nodeServiceInstance  NodeService
)

// NodeServiceImpl node service struct
type NodeServiceImpl struct {
	kubeClient *kubeclient.Client
}

// NodeService for node method to operate db
type NodeService interface {
	createNode(*NodeInfo) error
	deleteNode(*NodeInfo) error
	countNodesByName(string, int) (int64, error)
	listManagedNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	listUnManagedNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	countAllNodesByName(string) (int64, error)
	listAllNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	updateNodeInfoBySerialNumber(string, *NodeInfo) error
	getNodeByUniqueName(string) (*NodeInfo, error)
	getNodeInfoBySerialNumber(string) (*NodeInfo, error)
	getNodeByID(uint64) (*NodeInfo, error)
	getNodeBySn(string) (*NodeInfo, error)
	getManagedNodeByID(uint64) (*NodeInfo, error)
	countGroupsByNode(uint64) (int64, error)
	getGroupsByNodeID(uint64) (*[]NodeGroup, error)
	checkNodeManagedStatus(uint64, int) error

	createNodeGroup(*NodeGroup) error
	getNodeGroupsByName(uint64, uint64, string) (*[]NodeGroup, error)
	countNodeGroupsByName(string) (int64, error)
	getNodeGroupByID(uint64) (*NodeGroup, error)
	updateNodeGroupRes(uint64, map[string]interface{}) (int64, error)

	addNodeToGroup(*NodeRelation, string) error
	deleteNodeToGroup(*NodeRelation) (int64, error)
	countNodeByGroup(uint64) (int64, error)

	getRelationsByNodeID(uint64) (*[]NodeRelation, error)
	updateNode(uint64, int, map[string]interface{}) (int64, error)
	updateGroup(uint64, map[string]interface{}) (int64, error)
	listNodeRelationsByGroupId(uint64) (*[]NodeRelation, error)
	deleteNodeGroup(uint64, *[]NodeRelation) error
	listNodes() (*[]NodeInfo, error)
	deleteAllUnManagedNodes() error
	deleteSingleNodeRelation(uint64, uint64) error
	deleteUnmanagedNode(*NodeInfo) error
}

// GetTableCount get table count
func GetTableCount(tb interface{}) (int, error) {
	var total int64
	err := database.GetDb().Model(tb).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

// NodeServiceInstance returns the singleton instance of user service
func NodeServiceInstance() NodeService {
	nodeServiceSingleton.Do(func() {
		nodeServiceInstance = &NodeServiceImpl{
			kubeClient: kubeclient.GetKubeClient(),
		}
	})
	return nodeServiceInstance
}

func (n *NodeServiceImpl) db() *gorm.DB {
	return database.GetDb()
}

// CreateNode Create Node Db
func (n *NodeServiceImpl) createNode(nodeInfo *NodeInfo) error {
	return n.db().Model(NodeInfo{}).Where(NodeInfo{SerialNumber: nodeInfo.SerialNumber}).Assign(
		*nodeInfo).FirstOrCreate(nodeInfo).Error
}

// CreateNodeGroup Create Node Db
func (n *NodeServiceImpl) createNodeGroup(nodeGroup *NodeGroup) error {
	return n.db().Model(NodeGroup{}).Create(nodeGroup).Error
}

// GetNodesByName return SQL result
func (n *NodeServiceImpl) listManagedNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes,
		n.db().Where("is_managed = ?", managed).Scopes(getNodeByLikeName(page, pageSize, nodeName)).
			Find(&nodes).Error
}

// listUnManagedNodesByName return SQL result
func (n *NodeServiceImpl) listUnManagedNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes,
		n.db().Where("is_managed = ?", unmanaged).Scopes(getNodeByLikeName(page, pageSize, nodeName)).
			Find(&nodes).Error
}

// listAllNodesByName return SQL result
func (n *NodeServiceImpl) listAllNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes, n.db().Model(&NodeInfo{}).Scopes(getNodeByLikeName(page, pageSize, nodeName)).Find(&nodes).Error
}

// GetNodeGroupsByName return SQL result
func (n *NodeServiceImpl) getNodeGroupsByName(pageNum, pageSize uint64, nodeGroup string) (*[]NodeGroup, error) {
	var nodeGroups []NodeGroup
	return &nodeGroups,
		n.db().Scopes(common.Paginate(pageNum, pageSize), whereGroupNameLike(nodeGroup)).
			Find(&nodeGroups).Error
}

func (n *NodeServiceImpl) countNodeGroupsByName(nodeGroup string) (int64, error) {
	var nodeGroupCount int64
	return nodeGroupCount,
		n.db().Model(&NodeGroup{}).Scopes(whereGroupNameLike(nodeGroup)).
			Count(&nodeGroupCount).Error
}

func whereGroupNameLike(nodeGroupName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("INSTR(group_name, ?)", nodeGroupName)
	}
}

// DeleteNodeToGroup delete Node Db
func (n *NodeServiceImpl) deleteNodeToGroup(relation *NodeRelation) (int64, error) {
	stmt := n.db().Model(NodeRelation{}).Where("group_id = ? and node_id=?",
		relation.GroupID, relation.NodeID).Delete(relation)
	return stmt.RowsAffected, stmt.Error
}

// GetNodeByUniqueName get node info by Serial Number in k8s
func (n *NodeServiceImpl) updateNodeInfoBySerialNumber(sn string, nodeInfo *NodeInfo) error {
	return n.db().Model(NodeInfo{}).Where(NodeInfo{SerialNumber: sn}).Assign(
		*nodeInfo).FirstOrCreate(nodeInfo).Error
}

// GetNodeByUniqueName get node info by unique name in k8s
func (n *NodeServiceImpl) getNodeByUniqueName(name string) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db().Model(NodeInfo{}).Where("unique_name=?", name).First(&node).Error
}

// GetNodeByUniqueName get node info by serial number
func (n *NodeServiceImpl) getNodeInfoBySerialNumber(name string) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db().Model(NodeInfo{}).Where("serial_number=?", name).First(&node).Error
}

func (n *NodeServiceImpl) countNodeByGroup(groupID uint64) (int64, error) {
	var num int64
	return num, n.db().Model(NodeRelation{}).Where("group_id = ?", groupID).Count(&num).Error
}

// GetNodeGroupByID get node group info by group id
func (n *NodeServiceImpl) getNodeGroupByID(groupID uint64) (*NodeGroup, error) {
	var nodeGroup NodeGroup
	return &nodeGroup, n.db().Model(NodeGroup{}).Where("id = ?", groupID).First(&nodeGroup).Error
}

// GetNodeByID return node info by group id
func (n *NodeServiceImpl) getNodeByID(nodeID uint64) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db().Model(NodeInfo{}).Where("id = ?", nodeID).First(&node).Error
}

func (n *NodeServiceImpl) getNodeBySn(sn string) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db().Model(NodeInfo{}).Where("serial_number = ?", sn).First(&node).Error
}

func getNodeByLikeName(page, pageSize uint64, nodeName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(node_name, ?) AND ip != ''", nodeName)
	}
}

// GetNodeRelationByNodeId get nodeRelation
func (n *NodeServiceImpl) getRelationsByNodeID(id uint64) (*[]NodeRelation, error) {
	var nodeRelation []NodeRelation
	return &nodeRelation,
		n.db().Where(&NodeRelation{NodeID: id}).Find(&nodeRelation).Error
}

// UpdateNode update node
func (n *NodeServiceImpl) updateNode(id uint64, isManaged int, columns map[string]interface{}) (int64, error) {
	stmt := n.db().Model(&NodeInfo{}).
		Where("`id` = ? and `is_managed` = ? and `ip` != ''", id, isManaged).
		UpdateColumns(columns)
	return stmt.RowsAffected, stmt.Error
}

func (n *NodeServiceImpl) checkNodeManagedStatus(nodeId uint64, expected int) error {
	expectedStatus := expected == managed
	var node NodeInfo
	if err := n.db().Model(NodeInfo{}).Where("id = ?", nodeId).First(&node).Error; err != nil {
		return fmt.Errorf("failed to get node info from db")
	}
	statusMap := map[int]string{managed: "managed", unmanaged: "unmanaged"}
	if node.IsManaged != expectedStatus {
		return fmt.Errorf("node is not %s", statusMap[expected])
	}
	return nil
}

// UpdateGroup update group
func (n *NodeServiceImpl) updateGroup(id uint64, columns map[string]interface{}) (int64, error) {
	stmt := n.db().Model(&NodeGroup{}).Where("`id` = ?", id).UpdateColumns(columns)
	return stmt.RowsAffected, stmt.Error
}

func (n *NodeServiceImpl) listNodeRelationsByGroupId(groupId uint64) (*[]NodeRelation, error) {
	var relations []NodeRelation
	return &relations,
		n.db().Model(&NodeRelation{}).Where(&NodeRelation{GroupID: groupId}).Find(&relations).Error
}

func (n *NodeServiceImpl) getManagedNodeByID(nodeID uint64) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db().Model(NodeInfo{}).Where("id = ? and is_managed = ?", nodeID, managed).First(&node).Error
}

func (n *NodeServiceImpl) countGroupsByNode(nodeID uint64) (int64, error) {
	var num int64
	return num, n.db().Model(NodeRelation{}).Where("node_id = ?", nodeID).Count(&num).Error
}

func (n *NodeServiceImpl) getGroupsByNodeID(nodeID uint64) (*[]NodeGroup, error) {
	var nodeGroups []NodeGroup
	return &nodeGroups, n.db().Model(NodeGroup{}).
		Select("*").
		Joins("LEFT JOIN node_relations ON node_groups.id = node_relations.group_id").
		Where("node_relations.node_id = ?", nodeID).
		Scan(&nodeGroups).Error
}

func (n *NodeServiceImpl) listNodes() (*[]NodeInfo, error) {
	var nodes []NodeInfo
	err := n.db().Model(NodeInfo{}).Limit(maxNodeInfos).Find(&nodes).Error
	return &nodes, err
}

// countNodesByName count nodes by name
func (n *NodeServiceImpl) countNodesByName(name string, isManaged int) (int64, error) {
	var count int64
	return count,
		n.db().Model(&NodeInfo{}).Where("INSTR(node_name, ?) and is_managed = ? and ip != ''", name, isManaged).
			Count(&count).Error
}

// countNodesByName count all nodes by name
func (n *NodeServiceImpl) countAllNodesByName(name string) (int64, error) {
	var count int64
	return count,
		n.db().Model(&NodeInfo{}).Where("INSTR(node_name, ?) and ip != ''", name).
			Count(&count).Error
}

func (n *NodeServiceImpl) updateNodeGroupRes(groupId uint64, columns map[string]interface{}) (int64, error) {
	stmt := n.db().Model(NodeGroup{}).Where("id = ?", groupId).UpdateColumns(columns)
	return stmt.RowsAffected, stmt.Error
}

func (n *NodeServiceImpl) deleteAllUnManagedNodes() error {
	return n.db().Model(NodeInfo{}).Where("`is_managed` = ?", unmanaged).Delete(&NodeInfo{}).Error
}

func (n *NodeServiceImpl) addNodeToGroup(relation *NodeRelation, uniqueName string) error {
	return database.Transaction(n.db(), func(tx *gorm.DB) error {
		if err := tx.Model(NodeRelation{}).Create(relation).Error; err != nil {
			return fmt.Errorf("db create node relation error, groupID %d, nodeID %d: %v",
				relation.GroupID, relation.NodeID, err)
		}
		label := map[string]string{fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, relation.GroupID): ""}
		if _, err := kubeclient.GetKubeClient().AddNodeLabels(uniqueName, label); err != nil {
			hwlog.RunLog.Errorf("k8s add label err %v", err)
			return err
		}
		return nil
	})
}

// DeleteNodeByName delete node
func (n *NodeServiceImpl) deleteNode(nodeInfo *NodeInfo) error {
	return database.Transaction(n.db(), func(tx *gorm.DB) error {
		if err := tx.Model(&NodeRelation{}).Where(&NodeRelation{NodeID: nodeInfo.ID}).
			Delete(&NodeRelation{}).Error; err != nil {
			return fmt.Errorf("db delete node(%d) relation error", nodeInfo.ID)
		}
		if err := tx.Model(&NodeInfo{}).Where("node_name = ?", nodeInfo.NodeName).
			Delete(nodeInfo).Error; err != nil {
			return fmt.Errorf("db delete node(%d) error", nodeInfo.ID)
		}
		if err := n.kubeClient.DeleteNode(nodeInfo.UniqueName); err != nil && isNodeNotFound(err) {
			hwlog.RunLog.Warnf("k8s dont have this node(%s), err=%v", nodeInfo.UniqueName, err)
		} else if err != nil {
			return fmt.Errorf("k8s delete node(%s) failed", nodeInfo.UniqueName)

		}
		return nil
	})
}

func (n *NodeServiceImpl) deleteUnmanagedNode(nodeInfo *NodeInfo) error {
	return database.Transaction(n.db(), func(tx *gorm.DB) error {
		if err := tx.Model(&NodeInfo{}).Where("node_name = ?", nodeInfo.NodeName).
			Delete(nodeInfo).Error; err != nil {
			return fmt.Errorf("db delete node(%d) error", nodeInfo.ID)
		}
		if err := n.kubeClient.DeleteNode(nodeInfo.UniqueName); err != nil && isNodeNotFound(err) {
			hwlog.RunLog.Warnf("k8s dont have this node(%s), err=%v", nodeInfo.UniqueName, err)
		} else if err != nil {
			return fmt.Errorf("k8s delete node(%s) failed", nodeInfo.UniqueName)

		}
		return nil
	})
}

func (n *NodeServiceImpl) deleteSingleNodeRelation(groupID, nodeID uint64) error {
	return database.Transaction(n.db(), func(tx *gorm.DB) error {
		return deleteRelation(tx, groupID, nodeID)
	})
}

func deleteRelation(tx *gorm.DB, groupID, nodeID uint64) error {
	var nodeInfo NodeInfo
	if err := tx.Model(NodeInfo{}).Where("id = ?", nodeID).First(&nodeInfo).Error; err != nil {
		return fmt.Errorf("db get node %d failed", nodeID)
	}
	stmt := tx.Model(NodeRelation{}).Where("group_id = ? and node_id=?", groupID, nodeID).Delete(&NodeRelation{})
	if stmt.Error != nil {
		return fmt.Errorf("db delete node %d to group %d failed", nodeID, groupID)
	}
	if stmt.RowsAffected < 1 {
		return fmt.Errorf("no such relation(node:%d, group:%d)", nodeID, groupID)
	}
	nodeLabel := fmt.Sprintf("%s%d", common.NodeGroupLabelPrefix, groupID)
	_, err := kubeclient.GetKubeClient().DeleteNodeLabels(nodeInfo.UniqueName, []string{nodeLabel})
	if err != nil && isNodeNotFound(err) {
		hwlog.RunLog.Warnf("k8s delete label failed, err=%v", err)
	} else if err != nil {
		hwlog.RunLog.Errorf("k8s delete label(group %d) failed: %v", groupID, err)
		return fmt.Errorf("k8s delete label(group %d) failed", groupID)
	}
	return nil
}

func (n *NodeServiceImpl) deleteNodeGroup(groupID uint64, relations *[]NodeRelation) error {
	return database.Transaction(n.db(), func(tx *gorm.DB) error {
		for _, relation := range *relations {
			if err := deleteRelation(tx, groupID, relation.NodeID); err != nil {
				return fmt.Errorf("delete node relation failed, when delete node group:%s", err.Error())
			}
		}
		if stmt := tx.Model(NodeGroup{}).Where("id = ?", groupID).Delete(&NodeGroup{}); stmt.Error != nil ||
			stmt.RowsAffected != 1 {
			return fmt.Errorf("delete node group by group id %d failed", groupID)
		}
		return nil
	})
}
