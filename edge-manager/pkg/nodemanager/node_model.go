// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/database"
)

var (
	nodeServiceSingleton sync.Once
	nodeServiceInstance  NodeService
)

// NodeServiceImpl node service struct
type NodeServiceImpl struct {
	db *gorm.DB
}

// NodeService for node method to operate db
type NodeService interface {
	createNode(*NodeInfo) error
	deleteNodeByName(*NodeInfo) (int64, error)
	listManagedNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	listUnManagedNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	listAllNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	getNodeByUniqueName(string) (*NodeInfo, error)
	getNodeByID(int64) (*NodeInfo, error)
	getManagedNodeByID(int64) (*NodeInfo, error)
	countGroupsByNode(int64) (int64, error)

	createNodeGroup(*NodeGroup) error
	getNodeGroupsByName(uint64, uint64, string) (*[]NodeGroup, error)
	countNodeGroupsByName(string) (int64, error)
	getNodeGroupByID(int64) (*NodeGroup, error)

	deleteNodeToGroup(*NodeRelation) (int64, error)
	countNodeByGroup(groupID int64) (int64, error)

	getRelationsByNodeID(int64) (*[]NodeRelation, error)
	updateNode(int64, int, map[string]interface{}) (int64, error)
	updateGroup(int64, map[string]interface{}) (int64, error)
	deleteRelationsToNode(int64) error
	listNodeRelationsByGroupId(int64) (*[]NodeRelation, error)
	addNodeToGroup(*[]NodeRelation) error
	deleteNodeGroup(groupID int64) (int64, error)
	listNodes() (*[]NodeInfo, error)
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
		nodeServiceInstance = &NodeServiceImpl{db: database.GetDb()}
	})
	return nodeServiceInstance
}

// CreateNode Create Node Db
func (n *NodeServiceImpl) createNode(nodeInfo *NodeInfo) error {
	return n.db.Model(NodeInfo{}).Create(nodeInfo).Error
}

// CreateNodeGroup Create Node Db
func (n *NodeServiceImpl) createNodeGroup(nodeGroup *NodeGroup) error {
	return n.db.Model(NodeGroup{}).Create(nodeGroup).Error
}

// DeleteNodeByName delete node
func (n *NodeServiceImpl) deleteNodeByName(nodeInfo *NodeInfo) (int64, error) {
	stmt := n.db.Model(&NodeInfo{}).Where("node_name = ?",
		nodeInfo.NodeName).Delete(nodeInfo)
	return stmt.RowsAffected, stmt.Error
}

// GetNodesByName return SQL result
func (n *NodeServiceImpl) listManagedNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes,
		n.db.Where("is_managed = ?", managed).Scopes(getNodeByLikeName(page, pageSize, nodeName)).
			Find(&nodes).Error
}

// listUnManagedNodesByName return SQL result
func (n *NodeServiceImpl) listUnManagedNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes,
		n.db.Where("is_managed = ?", unmanaged).Scopes(getNodeByLikeName(page, pageSize, nodeName)).
			Find(&nodes).Error
}

// listAllNodesByName return SQL result
func (n *NodeServiceImpl) listAllNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes, n.db.Model(&NodeInfo{}).Scopes(getNodeByLikeName(page, pageSize, nodeName)).Find(&nodes).Error
}

// GetNodeGroupsByName return SQL result
func (n *NodeServiceImpl) getNodeGroupsByName(pageNum, pageSize uint64, nodeGroup string) (*[]NodeGroup, error) {
	var nodeGroups []NodeGroup
	return &nodeGroups,
		n.db.Scopes(common.Paginate(pageNum, pageSize), whereGroupNameLike(nodeGroup)).
			Find(&nodeGroups).Error
}

func (n *NodeServiceImpl) countNodeGroupsByName(nodeGroup string) (int64, error) {
	var nodeGroupCount int64
	return nodeGroupCount,
		n.db.Model(&NodeGroup{}).Scopes(whereGroupNameLike(nodeGroup)).
			Count(&nodeGroupCount).Error
}

func whereGroupNameLike(nodeGroupName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("INSTR(group_name, ?)", nodeGroupName)
	}
}

// DeleteNodeToGroup delete Node Db
func (n *NodeServiceImpl) deleteNodeToGroup(relation *NodeRelation) (int64, error) {
	stmt := n.db.Model(NodeRelation{}).Where("group_id = ? and node_id=?",
		relation.GroupID, relation.NodeID).Delete(relation)
	return stmt.RowsAffected, stmt.Error
}

// GetNodeByUniqueName get node info by unique name in k8s
func (n *NodeServiceImpl) getNodeByUniqueName(name string) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db.Model(NodeInfo{}).Where("unique_name=?", name).First(&node).Error
}

func (n *NodeServiceImpl) countNodeByGroup(groupID int64) (int64, error) {
	var num int64
	return num, n.db.Model(NodeRelation{}).Where("group_id = ?", groupID).Count(&num).Error
}

// GetNodeGroupByID get node group info by group id
func (n *NodeServiceImpl) getNodeGroupByID(groupID int64) (*NodeGroup, error) {
	var nodeGroup NodeGroup
	return &nodeGroup, n.db.Model(NodeGroup{}).Where("id = ?", groupID).First(&nodeGroup).Error
}

// GetNodeByID return node info by group id
func (n *NodeServiceImpl) getNodeByID(nodeID int64) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db.Model(NodeInfo{}).Where("id = ?", nodeID).First(&node).Error
}

func getNodeByLikeName(page, pageSize uint64, nodeName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(common.Paginate(page, pageSize)).Where("INSTR(node_name, ?)", nodeName)
	}
}

// GetNodeRelationByNodeId get nodeRelation
func (n *NodeServiceImpl) getRelationsByNodeID(id int64) (*[]NodeRelation, error) {
	var nodeRelation []NodeRelation
	return &nodeRelation,
		n.db.Where(&NodeRelation{NodeID: id}).Find(&nodeRelation).Error
}

// UpdateNode update node
func (n *NodeServiceImpl) updateNode(id int64, isManaged int, columns map[string]interface{}) (int64, error) {
	stmt := n.db.Model(&NodeInfo{}).Where("`id` = ? and `is_managed` = ?", id, isManaged).UpdateColumns(columns)
	return stmt.RowsAffected, stmt.Error
}

// UpdateGroup update group
func (n *NodeServiceImpl) updateGroup(id int64, columns map[string]interface{}) (int64, error) {
	stmt := n.db.Model(&NodeGroup{}).Where("`id` = ?", id).UpdateColumns(columns)
	return stmt.RowsAffected, stmt.Error
}

func (n *NodeServiceImpl) deleteRelationsToNode(id int64) error {
	return n.db.Model(&NodeRelation{}).Where(&NodeRelation{NodeID: id}).Delete(&NodeRelation{}).Error
}

func (n *NodeServiceImpl) listNodeRelationsByGroupId(groupId int64) (*[]NodeRelation, error) {
	var relations []NodeRelation
	return &relations,
		n.db.Model(&NodeRelation{}).Where(&NodeRelation{GroupID: groupId}).Find(&relations).Error
}

func (n *NodeServiceImpl) getManagedNodeByID(nodeID int64) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db.Model(NodeInfo{}).Where("id = ? and is_managed = ?", nodeID, managed).First(&node).Error
}

// AddNodeToGroup add Node Db
func (n *NodeServiceImpl) addNodeToGroup(relation *[]NodeRelation) error {
	return n.db.Model(NodeRelation{}).Create(relation).Error
}

func (n *NodeServiceImpl) countGroupsByNode(nodeID int64) (int64, error) {
	var num int64
	return num, n.db.Model(NodeRelation{}).Where("node_id = ?", nodeID).Count(&num).Error
}

func (n *NodeServiceImpl) deleteNodeGroup(groupID int64) (int64, error) {
	stmt := n.db.Model(NodeGroup{}).Where("`id` = ?", groupID).Delete(&NodeGroup{})
	return stmt.RowsAffected, stmt.Error
}

func (n *NodeServiceImpl) listNodes() (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes, n.db.Model(NodeInfo{}).Find(&nodes).Error
}
