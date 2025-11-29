#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
