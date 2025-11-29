#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：任务服务资源定义
修改记录：2018-07-24 创建
"""

import os

from common.ResourceDefV1.resource import RfResource


class RfTaskServiceObj(RfResource):
    """
    功能描述：任务配置资源处理
    接口：NA
    修改记录：2018-07-24 创建
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
