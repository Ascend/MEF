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

"""
功    能：Processors资源 URL处理模块
"""
import json

from flask import request

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import AiProcessorResourceSerializer
from system_service.systems_serializer import CPUProcessorResourceSerializer
from system_service.systems_serializer import ProcessorResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query Processor collection information")
def rf_system_processor_collection():
    """
    功能描述：查询处理器接口集合信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query Processor collection information failed."
    try:
        # 获取 Processor Collection资源模板
        processor_resource = json.loads(ProcessorResourceSerializer().service.get_resource())
        return {"status": 200}, processor_resource
    except Exception as err:
        run_log.error(f"{input_err_info} reason is: {err}")
        ret_dict = make_error_dict("ResourceDoesNotExists", CommonConstants.ERR_CODE_404)
        return ret_dict, CommonConstants.ERR_PARTICULAR_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query CPU summary information")
def rf_system_processor_cpu():
    """
    功能描述：查询CPU概要信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query CPU summary information failed."
    try:
        # 获取 System CPU Collection 资源模板
        cpu_processor_resource = json.loads(CPUProcessorResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给BMC
        info = "ProcessorsSummary"
        ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False, info)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, cpu_processor_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{input_err_info} reason is: {err}")
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query AiProcessor summary information")
def rf_system_processor_ai():
    """
    功能描述：查询AI处理器资源信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query AiProcessor summary information failed."
    try:
        # 获取 System AiProcessor Collection资源模板
        ai_processor_resource = json.loads(AiProcessorResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给BMC
        ret_dict = LibRESTfulAdapter.lib_restful_interface("AiProcessor", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, ai_processor_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{input_err_info} reason is: {err}")
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
