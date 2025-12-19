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
功    能：任务服务资源定义
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfTaskServiceObj(RfResource):
    """
    功能描述：任务配置资源处理
    接口：NA
    """
    def create_sub_objects(self, base_path, rel_path):

        self.TasksSet = RfUTasksSetObj(
            base_path, os.path.normpath("redfish/v1/TaskService/1"))
        self.tasks_resource = RfTasksResourceObj(
            base_path, os.path.normpath("redfish/v1/TaskService/1/TasksResource"))


class RfUTasksSetObj(RfResource):
    pass


class RfTasksResourceObj(RfResource):
    pass
