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
功    能：Redfish Server 根服务资源定义
修改记录：2016-10-18 创建
"""
import os

from common.ResourceDefV1.service_root import RfServiceRoot
from om_common.ResourceDefV1.om_systems import OMRfSystemsCollection
from om_common.ResourceDefV1.event_subscription_service import RfEventSubscription


class OMRfServiceRoot(RfServiceRoot):
    """
    功能描述：创建该根服务下的所有资源
    接口：NA
    修改记录：2016-10-18 创建
    """

    def create_extend_sub_objects(self, base_path, rel_path):
        self.systems = OMRfSystemsCollection(base_path, os.path.normpath("redfish/v1/Systems"))
        self.subscription_service = RfEventSubscription(base_path, os.path.normpath("redfish/v1/EventService"))
