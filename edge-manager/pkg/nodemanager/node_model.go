// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/database"
	"sync"

	"gorm.io/gorm"
	"huawei.com/mindxedge/base/common"
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
	deleteNodeByName(*NodeInfo) error
	listNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	listUnManagedNodesByName(uint64, uint64, string) (*[]NodeInfo, error)
	getNodeByUniqueName(string) (*NodeInfo, error)
	getNodeByID(int64) (*NodeInfo, error)

	createNodeGroup(*NodeGroup) error
	getNodeGroupsByName(uint64, uint64, string) (*[]NodeGroup, error)
	getNodeGroupByID(int64) (*NodeGroup, error)

	deleteNodeToGroup(*NodeRelation) error
	countNodeByGroup(groupID int64) (int64, error)

	getRelationsByNodeID(int64) (*[]NodeRelation, error)
	updateNode(int64, map[string]interface{}) error
	deleteRelationsToNode(int64) error
	deleteRelation(*NodeRelation) (int64, error)
	countNodesByStatus(string) (int64, error)
	listNodeGroup(uint64, uint64, string) (*[]NodeGroup, error)
	listNodeRelationsByGroupId(int64) (*[]NodeRelation, error)
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
func (n *NodeServiceImpl) deleteNodeByName(nodeInfo *NodeInfo) error {
	return n.db.Model(&NodeInfo{}).Where("node_name = ?",
		nodeInfo.NodeName).Delete(nodeInfo).Error
}

// GetNodesByName return SQL result
func (n *NodeServiceImpl) listNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
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

// GetNodeGroupsByName return SQL result
func (n *NodeServiceImpl) getNodeGroupsByName(page, pageSize uint64, nodeGroup string) (*[]NodeGroup, error) {
	var nodes []NodeGroup
	return &nodes,
		n.db.Scopes(getNodeByLikeName(page, pageSize, nodeGroup)).
			Find(&nodes).Error
}

// DeleteNodeToGroup delete Node Db
func (n *NodeServiceImpl) deleteNodeToGroup(relation *NodeRelation) error {
	return n.db.Model(NodeRelation{}).Delete(relation).Error
}

func (n *NodeServiceImpl) getNodeByUniqueName(name string) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db.Model(NodeInfo{}).Where("unique_name=?", name).First(&node).Error
}

func (n *NodeServiceImpl) countNodeByGroup(groupID int64) (int64, error) {
	var num int64
	return num, n.db.Model(NodeRelation{}).Where("group_id = ?", groupID).Count(&num).Error
}

func (n *NodeServiceImpl) getNodeGroupByID(groupID int64) (*NodeGroup, error) {
	var nodeGroup NodeGroup
	return &nodeGroup, n.db.Model(NodeGroup{}).Where("id = ?", groupID).First(&nodeGroup).Error
}

func (n *NodeServiceImpl) getNodeByID(nodeID int64) (*NodeInfo, error) {
	var node NodeInfo
	return &node, n.db.Model(NodeInfo{}).Where("id = ?", nodeID).First(&node).Error
}

func getNodeByLikeName(page, pageSize uint64, nodeName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(paginate(page, pageSize)).Where("node_name like ?", "%"+nodeName+"%")
	}
}

// GetNodeRelationByNodeId get nodeRelation
func (n *NodeServiceImpl) getRelationsByNodeID(id int64) (*[]NodeRelation, error) {
	var nodeRelation []NodeRelation
	return &nodeRelation,
		n.db.Where(&NodeRelation{NodeID: id}).Find(&nodeRelation).Error
}

// UpdateNode update node
func (n *NodeServiceImpl) updateNode(id int64, columns map[string]interface{}) error {
	return n.db.Model(&NodeInfo{}).Where("`id` = ?", id).UpdateColumns(columns).Error
}

func (n *NodeServiceImpl) deleteRelationsToNode(id int64) error {
	return n.db.Model(&NodeRelation{}).Where(&NodeRelation{NodeID: id}).Delete(&NodeRelation{}).Error
}

func (n *NodeServiceImpl) deleteRelation(relation *NodeRelation) (int64, error) {
	stmt := n.db.Model(&NodeRelation{}).
		Where(&NodeRelation{NodeID: relation.NodeID, GroupID: relation.GroupID}).
		Delete(&NodeRelation{})
	return stmt.RowsAffected, stmt.Error
}

// getNodeStatistics get node count by status
func (n *NodeServiceImpl) countNodesByStatus(status string) (int64, error) {
	var nodeCount int64
	return nodeCount,
		n.db.Model(&NodeInfo{}).Where(&NodeInfo{Status: status}).Count(&nodeCount).Error
}

func (n *NodeServiceImpl) listNodeGroup(pageNum uint64, pageSize uint64, name string) (*[]NodeGroup, error) {
	var nodeGroups []NodeGroup
	return &nodeGroups,
		n.db.Scopes(getNodeGroupByLikeName(pageNum, pageSize, name)).Find(&nodeGroups).Error
}

func getNodeGroupByLikeName(page, pageSize uint64, nodeGroupName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(paginate(page, pageSize)).Where("group_name like ?", "%"+nodeGroupName+"%")
	}
}

func (n *NodeServiceImpl) listNodeRelationsByGroupId(groupId int64) (*[]NodeRelation, error) {
	var relations []NodeRelation
	return &relations,
		n.db.Model(&NodeRelation{}).Where(&NodeRelation{GroupID: groupId}).Find(&relations).Error
}

func paginate(page, pageSize uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = common.DefaultPage
		}
		if pageSize > common.DefaultMaxPageSize {
			pageSize = common.DefaultMaxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
