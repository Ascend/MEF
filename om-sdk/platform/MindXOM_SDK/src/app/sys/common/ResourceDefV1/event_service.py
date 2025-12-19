#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功    能：Redfish Server Event处理模块
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfEventServiceObj(RfResource):
    """
    功能描述：创建EventService资源并导入一级目录模板
    接口：NA
    """
    def create_sub_objects(self, base_path, rel_path):
        self.eventColl = \
            RfEventCollection(base_path,
                              os.path.
                              normpath("redfish/v1/EventService/"
                                       "Subscriptions"))

    def patch_resource(self, patch_data):
        pass


class RfEventCollection(RfResource):
    """
    功能描述：创建EventService子对象, 导入实例配置模板
    接口：NA
    """
    def create_sub_objects(self, base_path, rel_path):
        self.subscriptions = \
            RfEventSubscriptionsObj(base_path,
                                    os.path.normpath("redfish/v1/EventService/"
                                                     "Subscriptions/1"))


class RfEventSubscriptionsObj(RfResource):
    pass
