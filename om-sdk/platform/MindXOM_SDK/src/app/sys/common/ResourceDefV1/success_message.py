#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server Success Message资源定义
修改记录：2022-11-24 创建
"""
import os.path

from common.ResourceDefV1.resource import RfResource


class RfSuccessMessage(RfResource):
    """
    功能描述：创建 Success Message 资源对象, 导入配置模板
    接口：NA
    修改记录：2022-11-24 创建
    """
    SUCCESS_MESSAGE_REOURCE_DIR = os.path.normpath("redfish/v1/SuccessMessage")

    success_message_resource: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.success_message_resource = RfResource(base_path, self.SUCCESS_MESSAGE_REOURCE_DIR)