#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server Systems资源定义
修改记录：2019-1-18 创建
"""

import os
from common.ResourceDefV1.resource import RfResource


class RfErrorObj(RfResource):

    def create_sub_objects(self, base_path, rel_path):
        self.errorcolection = Rferrorcolection(
            base_path, os.path.normpath("redfish/v1/ErrorCollection/1"))


class Rferrorcolection(RfResource):
    pass
