#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功    能：升级服务资源定义
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfUpgradeService(RfResource):
    """
    功能描述：升级服务配置资源处理
    接口：NA
    """
    GET_COLLECTION_RESOURCE_DIR = os.path.normpath("redfish/v1/UpgradeService")
    GET_UPGRADE_SERVICE_ACTION_DIR = os.path.normpath("redfish/v1/UpgradeService/Actions")

    get_resource_collection: RfResource
    actions: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.get_resource_collection = RfResource(base_path, self.GET_COLLECTION_RESOURCE_DIR)
        self.actions = RfResource(base_path, self.GET_UPGRADE_SERVICE_ACTION_DIR)
