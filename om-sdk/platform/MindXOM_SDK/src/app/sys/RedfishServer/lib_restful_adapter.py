#!/usr/bin/env python
# -*- coding:utf-8 -*-
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
功    能：插件适配，用于RESTful接口调用，通过其发送 Socket 信息给适配层
"""
import ast
import json

import ibma_client
from common.log.logger import run_log
from common.restfull_socket_model import RestFullSocketModel as RestFullSocketModel
from common.utils.app_common_method import AppCommonMethod


class LibRESTfulAdapter(object):
    """
    功能描述：插件适配
    接口：NA
    """

    @staticmethod
    def lib_restful_interface(model_name, request_type,
                              request_data, need_list=False,
                              item1=None, item2=None, item3=None, item4=None):
        """
        功能描述：底层接口的上层调用接口
        参数：modelName 模块名称，用于加载具体的对象
        requestType 请求类型，包括 POST，PATCH
        request_data 请求的属性
        needList 是否是一个列表，True 或 False
        item1 第一级资源
        item2 第二级资源
        item3 第三级资源
        item4 第四级资源
        返回值：无
        异常描述：NA
        """
        socket_model = RestFullSocketModel("lib_restful_interface", model_name,
                                           request_type, request_data, need_list, item1, item2, item3, item4)
        return LibRESTfulAdapter.send_and_parse_result(socket_model)

    @staticmethod
    def start_timer():
        """
        功能描述：启动 timer 定时器
        参数：无
        返回值：NA
        异常描述：NA
        """
        socket_model = RestFullSocketModel("start_timer", None)
        return LibRESTfulAdapter.send_and_parse_result(socket_model)

    @staticmethod
    def send_and_parse_result(socket_model):
        """
        功能描述：发送 Socket命令并转换结果
        参数：sModel socket模型
        返回值：NA
        异常描述：NA
        """
        # 将参数转成字符串
        msg = socket_model.get_socket_info()
        # 发送 Socket 命令，获取结果
        ret = ibma_client.Client.send_msg(msg)
        if ret[0] != 0:
            # 出现了错误，需要处理下
            return AppCommonMethod.get_json_error_by_array(ret)

        ret = ret[1]
        try:
            # 解析出dict的内容
            return json.loads(ret) if ret not in ["", None] else None
        except Exception:
            # 不是 dict
            try:
                if isinstance(ret, bytes):
                    ret = ret.decode()
                # 通过eval判断是否是list
                return ast.literal_eval(ret)
            except Exception as e:
                run_log.error("error {}".format(e))
                return ret

    @staticmethod
    def check_status_is_ok(ret_dict):
        """
        功能描述：检测status是否为OK
        参数：retDict字典中包含key(status)和key(message)
        返回值：status为OK,则返回True,否则返回False
        异常描述：NA
        """
        return isinstance(ret_dict, dict) and ret_dict.get("status") == AppCommonMethod.OK
