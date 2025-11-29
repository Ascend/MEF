# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

"""
功    能：Redfish Server Systems资源定义
修改记录：2016-10-18 创建
"""
import os

from common.ResourceDefV1.resource import RfResource
from common.ResourceDefV1.systems import RfSystemsCollection


class OMRfSystemsCollection(RfSystemsCollection):
    """
    功能描述：创建Systems对象集合, 导入配置模板
    接口：NA
    修改记录：2016-10-18 创建
    """

    DIGITALWARRANTY_RESOURCE_DIR = os.path.normpath("redfish/v1/Systems/DigitalWarranty")
    digitalwarranty_resource: RfResource

    def create_extend_sub_objects(self, base_path, rel_path):
        self.digitalwarranty_resource = RfResource(base_path, self.DIGITALWARRANTY_RESOURCE_DIR)
