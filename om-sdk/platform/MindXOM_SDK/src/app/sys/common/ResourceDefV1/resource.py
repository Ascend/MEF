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
功    能：Redfish Server 父资源定义
"""

import os
import json

from common.file_utils import FileCheck


class RfResource:
    """
    功能描述：父类资源初始化及方法定义
    接口：NA
    """
    def __init__(self, base_path, rel_path):
        self.response = None
        path = os.path.join(base_path, rel_path)
        index_file_path = os.path.join(path, "index.json")

        try:
            res = FileCheck.check_path_is_exist_and_valid(index_file_path)
            if not res:
                raise ValueError(f"{index_file_path} path invalid : {res.error}")

            with open(index_file_path, "r") as res_file:
                self.resData = json.load(res_file)
        except IOError as io_err:
            raise io_err
        except Exception as err:
            raise err
        self.create_sub_objects(base_path, rel_path)
        self.final_init_processing(base_path, rel_path)

    def create_sub_objects(self, base_path, rel_path):
        pass

    def final_init_processing(self, base_path, rel_path):
        pass

    def get_resource(self):
        self.response = json.dumps(self.resData, indent=4)
        return self.response
