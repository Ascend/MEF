#!/usr/bin/env python
# -*- coding:utf-8 -*-
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
功    能：MidWareSocket模型，用于socket通讯使用
"""
import json
from common.utils.app_common_method import AppCommonMethod


class MidWareSocketModel(object):
    """
    功能描述：MidWareSocket模型
    接口：NA
    """

    def __init__(self, header, route, content=None):
        """
        功能描述：初始化函数
        参数：header 资源名称
              route 请求类型
              content 内容信息
        返回值：无
        异常描述：NA
        """
        self.header = header
        self.route = route
        self.content = content

    @staticmethod
    def check_socket_model(json_data):
        """
        功能描述：校验socket模型
        参数：json_data json数据
        返回值：Socket对象
        异常描述：NA
        """
        header = json_data.get("header", None)
        route = json_data.get("route", None)

        if header and route:
            return True
        else:
            return False

    @staticmethod
    def get_socket_model(json_data):
        """
        功能描述：将json文本转成对象
        参数：json_data json数据
        返回值：Socket对象
        异常描述：NA
        """

        header = json_data.get("header", None)
        route = json_data.get("route", None)
        content = json_data.get("content", None)

        return MidWareSocketModel(header, route, content)

    def get_socket_info(self):
        """
        功能描述：获取socket通讯使用的内容
        参数：NA
        返回值：字符串
        异常描述：NA
        """
        ret_dict = AppCommonMethod.get_json_info(self)

        return json.dumps(ret_dict)
