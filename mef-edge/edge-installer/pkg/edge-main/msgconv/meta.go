// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package msgconv
package msgconv

import (
	"errors"
	"strings"

	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/database"
)

var resourceTypeWhiteList = utils.NewSet(constants.ResourceTypePod, constants.ResourceTypePodPatch,
	constants.ResourceTypeNode, constants.ResourceTypeNodePatch, constants.ResourceTypeConfigMap)

// saveMetadata automatically saves message to sync database between edgecore and edge-main.
// This method needs to intercept messages that could change the database of edgecore.
// Modify resourceTypeWhiteList and msgconvHandlers respectively if a new category of message
// is needed to saved in database.
func (h *messageHandler) saveMetadata(message *model.Message) error {
	resourceType, ok := parseResourceType(message.KubeEdgeRouter.Resource)
	if !ok {
		return nil
	}
	if !resourceTypeWhiteList.Find(resourceType) {
		return nil
	}

	switch message.KubeEdgeRouter.Operation {
	case constants.OptInsert, constants.OptUpdate:
		return insertIntoDb(message)
	case constants.OptDelete:
		return deleteFromDb(message)
	case constants.OptPatch:
		return patchToDb(message)
	default:
		return nil
	}
}

func insertIntoDb(message *model.Message) error {
	resourceType, ok := parseResourceType(message.KubeEdgeRouter.Resource)
	if !ok {
		return errors.New("unknown resource type")
	}

	return database.GetMetaRepository().CreateOrUpdate(database.Meta{
		Type:  resourceType,
		Key:   message.KubeEdgeRouter.Resource,
		Value: string(message.Content),
	})
}

func deleteFromDb(message *model.Message) error {
	return database.GetMetaRepository().DeleteByKey(message.KubeEdgeRouter.Resource)
}

func patchToDb(message *model.Message) error {
	metaKey := message.KubeEdgeRouter.Resource
	metaKey = strings.Replace(metaKey, constants.ActionDefaultNodePatch, constants.ActionDefaultNodeStatus, 1)
	metaKey = strings.Replace(metaKey, constants.ActionPodPatch, constants.ActionPod, 1)
	metaKey = strings.Replace(metaKey, constants.ResMefPodPatchPrefix, constants.ResMefPodPrefix, 1)
	resourceType, ok := parseResourceType(metaKey)
	if !ok {
		return errors.New("unknown resource type")
	}

	meta, err := database.GetMetaRepository().GetByKey(metaKey)
	if err != nil {
		return err
	}
	mergedBytes, err := util.MergePatch([]byte(meta.Value), model.UnformatMsg(message.Content))
	if err != nil {
		return err
	}
	return database.GetMetaRepository().CreateOrUpdate(database.Meta{
		Key:   metaKey,
		Type:  resourceType,
		Value: string(mergedBytes),
	})
}

func parseResourceType(resource string) (string, bool) {
	const resourceTypeIndex = 1

	tokens := strings.Split(resource, "/")
	if len(tokens) <= resourceTypeIndex {
		return "", false
	}
	return tokens[resourceTypeIndex], true
}
