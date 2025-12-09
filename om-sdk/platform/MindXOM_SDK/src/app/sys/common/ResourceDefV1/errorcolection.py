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
