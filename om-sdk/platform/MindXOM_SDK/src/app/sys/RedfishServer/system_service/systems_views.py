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
功    能：系统资源 URL处理模块
修改记录：2022-10-29 创建
"""
import threading
import json
import time

from flask import request

from common.checkers.param_checker import SystemChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict, lib_rest_check_status
from system_service.systems_serializer import EthernetEthIpListResourceSerializer
from system_service.systems_serializer import SystemServiceResourceSerializer
from system_service.systems_serializer import SystemTimeResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query system resources")
def rf_get_system_collection():
    """
    功能描述：查询系统资源信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-3-18 创建
    """
    message = "Query system resources failed."
    try:
        # 获取资源模板
        system_resource = json.loads(SystemServiceResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给iBMC
        ret_dictionary = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dictionary):
            return ret_dictionary, system_resource
        else:
            return ret_dictionary, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{message} reason is: {err}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


SYSTEM_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify system resources")
def rf_patch_system_collection():
    """
    功能描述：修改系统资源信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-3-18 创建
    """
    if SYSTEM_LOCK.locked():
        message = "Modify system failed because lock is locked."
        run_log.error(f"{message}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with SYSTEM_LOCK:
        message = "Modify system resources failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(SystemChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO
            # 获取Username和requestIp放入字典request_data_Dict
            request_data_dict["_User"] = RedfishGlobals.USER_NAME
            request_data_dict["_Xip"] = request.headers["X-Real-Ip"]
            # 获取资源模板
            system_resource = json.loads(SystemServiceResourceSerializer().service.get_resource())
            # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给iBMC
            ret_dictionary = LibRESTfulAdapter.lib_restful_interface("System", "PATCH", request_data_dict, False)
            ret = lib_rest_check_status(request_data_dict, ret_dictionary)
            if not ret:
                input_info = ret_dictionary.get("message", None)
                ret_dict = make_error_dict(input_info, CommonConstants.ERR_CODE_400)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给BMC
            return ret_dictionary, system_resource
        except Exception as err:
            run_log.error(f"{message} reason is: {err}")
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query system time")
def rf_get_system_time():
    """
    功能描述：返回系统时间
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2022-12-2 创建
    """
    date = time.ctime()
    ret = {"status": RedfishGlobals.SuccessfullyCode, "message": {"Datetime": date}}
    # 获取资源模板
    system_time_resource = json.loads(SystemTimeResourceSerializer().service.get_resource())
    return ret, system_time_resource


@RedfishGlobals.redfish_operate_adapter(request, "Query eth ip list")
def rf_eth_ip_list():
    """
    获取网口与IP地址列表
    """
    try:
        eth_ip_list_resource = json.loads(EthernetEthIpListResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("RestoreDefaultsAction", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, eth_ip_list_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("Query eth ip list failed: %s", err)
        ret_dict = make_error_dict("Query eth ip list failed.", CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
