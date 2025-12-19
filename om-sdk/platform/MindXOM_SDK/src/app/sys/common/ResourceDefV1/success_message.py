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
功    能：Redfish Server Success Message资源定义
"""
import os.path

from common.ResourceDefV1.resource import RfResource


class RfSuccessMessage(RfResource):
    """
    功能描述：创建 Success Message 资源对象, 导入配置模板
    接口：NA
    """
    SUCCESS_MESSAGE_REOURCE_DIR = os.path.normpath("redfish/v1/SuccessMessage")

    success_message_resource: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.success_message_resource = RfResource(base_path, self.SUCCESS_MESSAGE_REOURCE_DIR)