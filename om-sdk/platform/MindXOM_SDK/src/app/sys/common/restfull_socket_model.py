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
功    能：Socket 模型，用于socket通讯使用
"""
import json
from common.utils.app_common_method import AppCommonMethod


class RestFullSocketModel(object):
    """
    功能描述：socket 模型
    接口：NA
    """

    def __init__(self, method, model_name, request_type=None, request_data=None,
                 need_list=False, item1=None, item2=None, item3=None,
                 item4=None):
        """
                     功能描述：初始化函数
                     参数：method 函数名称
             modelName 模块名称，用于加载具体的对象
             needList 是否是一个列表，True 或 False
             item1 第一级资源
             item2 第二级资源
             item3 第三级资源
             item4 第四级资源
                     返回值：无
                     异常描述：NA
        """
        self.method = method
        self.model_name = model_name
        self.request_type = request_type
        self.request_data = request_data
        self.need_list = need_list
        self.item1 = item1
        self.item2 = item2
        self.item3 = item3
        self.item4 = item4

    @staticmethod
    def check_socket_model(json_data):
        """
        功能描述：将json文本转成对象
        参数：json_data json数据
        返回值：Socket对象
        异常描述：NA
        """
        method = json_data.get("method", None)
        if method:
            return True
        else:
            return False

    @staticmethod
    def get_socket_model(json_data):
        """
        功能描述：将json文本转成对象
        参数：jsonData json数据
        返回值：Socket对象
        异常描述：NA
        """
        method = json_data.get("method", None)
        model_name = json_data.get("model_name", None)
        request_type = json_data.get("request_type", None)
        request_data = json_data.get("request_data", None)
        need_list = json_data.get("need_list", False)
        item1 = json_data.get("item1", None)
        item2 = json_data.get("item2", None)
        item3 = json_data.get("item3", None)
        item4 = json_data.get("item4", None)

        return RestFullSocketModel(method, model_name, request_type, request_data,
                                   need_list, item1, item2, item3, item4)

    def get_socket_info(self):
        """
        功能描述：获取 socket通讯使用的内容
        参数：NA
        返回值：字符串
        异常描述：NA
        """
        r_dict = AppCommonMethod.get_json_info(self)
        return json.dumps(r_dict)
