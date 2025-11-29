#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：升级服务资源定义
修改记录：2021-11-7 创建
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfUpgradeService(RfResource):
    """
    功能描述：升级服务配置资源处理
    接口：NA
    修改记录：2021-11-7 创建
    """
    GET_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/UpgradeService")
    GET_UPGRADE_SERVICE_ACTION_DIR = os.path.normpath("redfish/v1/UpgradeService/Actions")

    get_resource_collection: RfResource
    actions: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.get_resource_collection = RfResource(base_path, self.GET_COLLECTION_RESOURCE_DIR)
        self.actions = RfResource(base_path, self.GET_UPGRADE_SERVICE_ACTION_DIR)
