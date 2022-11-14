// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/database"
	"sync"

	"gorm.io/gorm"
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
	CreateNode(*NodeInfo) error
	DeleteNodeByName(*NodeInfo) error
	GetNodesByName(uint64, uint64, string) (*[]NodeInfo, error)

	CreateNodeGroup(*NodeGroup) error
	GetNodeGroupsByName(uint64, uint64, string) (*[]NodeGroup, error)

	AddNodeToGroup(*NodeRelation) error
	DeleteNodeToGroup(*NodeRelation) error
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
func (n *NodeServiceImpl) CreateNode(nodeInfo *NodeInfo) error {
	return n.db.Model(NodeInfo{}).Create(nodeInfo).Error
}

// CreateNodeGroup Create Node Db
func (n *NodeServiceImpl) CreateNodeGroup(nodeGroup *NodeGroup) error {
	return n.db.Model(NodeGroup{}).Create(nodeGroup).Error
}

// DeleteNodeByName delete node
func (n *NodeServiceImpl) DeleteNodeByName(nodeInfo *NodeInfo) error {
	return database.GetDb().Model(&NodeInfo{}).Where("node_name = ?",
		nodeInfo.NodeName).Delete(nodeInfo).Error
}

// GetNodesByName return SQL result
func (n *NodeServiceImpl) GetNodesByName(page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
	var nodes []NodeInfo
	return &nodes,
		database.GetDb().Scopes(getNodeByLikeName(page, pageSize, nodeName)).
			Find(&nodes).Error
}

// GetNodeGroupsByName return SQL result
func (n *NodeServiceImpl) GetNodeGroupsByName(page, pageSize uint64, nodeGroup string) (*[]NodeGroup, error) {
	var nodes []NodeGroup
	return &nodes,
		database.GetDb().Scopes(getNodeByLikeName(page, pageSize, nodeGroup)).
			Find(&nodes).Error
}

// AddNodeToGroup add Node Db
func (n *NodeServiceImpl) AddNodeToGroup(relation *NodeRelation) error {
	return n.db.Model(NodeRelation{}).Create(relation).Error
}

// DeleteNodeToGroup delete Node Db
func (n *NodeServiceImpl) DeleteNodeToGroup(relation *NodeRelation) error {
	return n.db.Model(NodeRelation{}).Delete(relation).Error
}

func getNodeByLikeName(page, pageSize uint64, nodeName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(paginate(page, pageSize)).Where("node_name like ?", "%"+nodeName+"%")
	}
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
