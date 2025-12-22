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
功    能：Redfish Server Systems资源定义
"""
import os

from common.ResourceDefV1.resource import RfResource
from common.ResourceDefV1.systems import RfSystemsCollection


class OMRfSystemsCollection(RfSystemsCollection):
    """
    功能描述：创建Systems对象集合, 导入配置模板
    接口：NA
    """

    DIGITALWARRANTY_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/DigitalWarranty")
    digitalwarranty_resource: RfResource

    def create_extend_sub_objects(self, base_path, rel_path):
        self.digitalwarranty_resource = RfResource(base_path, self.DIGITALWARRANTY_RESOURCE_DIR)
