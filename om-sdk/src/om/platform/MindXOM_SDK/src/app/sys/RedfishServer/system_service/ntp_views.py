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
import json
import threading

from flask import request

from common.checkers.param_checker import NTPServiceChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import NtpResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query NTP service")
def rf_get_ntp_service_collection():
    """
    功能描述： 查询和配置NTP服务信息
    """
    message = "Query NTP service failed."
    try:
        # 获取资源模板
        ntp_resource = json.loads(NtpResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, ntp_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{message} reason is: {err}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


NTP_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Set NTP service")
def rf_patch_ntp_service_collection():
    if NTP_LOCK.locked():
        message = "Set NTP service failed because lock is locked."
        run_log.error(f"{message}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with NTP_LOCK:
        message = "Set NTP service failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(NTPServiceChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            # 获取资源模板
            ntp_resource = json.loads(NtpResourceSerializer().service.get_resource())
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            ret_dict = LibRESTfulAdapter.lib_restful_interface("NTPService", "PATCH", request_data_dict, False)
            if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                return ret_dict, ntp_resource

            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error(f"{message} reason is: {err}")
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
