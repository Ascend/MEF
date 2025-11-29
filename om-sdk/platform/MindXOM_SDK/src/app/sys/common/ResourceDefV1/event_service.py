#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server Event处理模块
修改记录：2016-10-18 创建
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfEventServiceObj(RfResource):
    """
    功能描述：创建EventService资源并导入一级目录模板
    接口：NA
    修改记录：2016-10-18 创建
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
    修改记录：2016-10-18 创建
    """
    def create_sub_objects(self, base_path, rel_path):
        self.subscriptions = \
            RfEventSubscriptionsObj(base_path,
                                    os.path.normpath("redfish/v1/EventService/"
                                                     "Subscriptions/1"))


class RfEventSubscriptionsObj(RfResource):
    pass
