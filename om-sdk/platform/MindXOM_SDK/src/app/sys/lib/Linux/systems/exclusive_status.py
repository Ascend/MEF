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
from bin.global_exclusive_control import GlobalExclusiveController


class ExclusiveStatus:

    def __init__(self):
        """
        功能描述：初始化函数
        参数：
        返回值：无
        异常描述：NA
        """
        self.system_busy = False

    def get_all_info(self):
        self.system_busy = GlobalExclusiveController.locked()
