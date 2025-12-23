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
import json

from flask import request

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import MemoryResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query memory summary information")
def rf_system_memory_summary_collection():
    """
    功能描述：查询内存概要信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query memory summary information failed."
    try:
        # 获取  Memory Collection资源模板
        memory_resource = json.loads(MemoryResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给BMC
        info = "MemorySummary"
        ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False, info)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, memory_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
