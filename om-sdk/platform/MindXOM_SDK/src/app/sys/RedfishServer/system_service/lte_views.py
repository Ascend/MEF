# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
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

from common.checkers.param_checker import LTEStatusInfoChecker, LTEConfigInfoChecker
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import CommonErrorCodes
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from net_manager.manager.fd_cfg_manager import FdCfgManager
from system_service.systems_serializer import LteResourceSerializer, LteStatusInfoResourceSerializer, \
    LteConfigInfoResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter

from system_service.systems_common import make_error_dict


@RedfishGlobals.redfish_operate_adapter(request, "Query LTE info")
def rf_get_system_lte():
    """
    功能描述：获取LTE顶层的信息
    参数：无
    返回值 响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query LTE info."
    try:
        # 获取 lte 顶层资源模板
        lte_resource = json.loads(LteResourceSerializer().service.get_resource())
        return {"status": 200}, lte_resource
    except Exception as err:
        run_log.error(f"{message} reason is: {err}")
        ret_dict = make_error_dict("ResourceDoesNotExists", CommonConstants.ERR_CODE_404)
        return ret_dict, CommonConstants.ERR_PARTICULAR_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query LTE status info")
def rf_get_system_lte_status_info():
    """
    功能描述：获取LTE的状态信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query LTE status info failed."
    try:
        lte_statusinfo_resource = json.loads(LteStatusInfoResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取 Shortwaves 相关信息
        ret_dict = LibRESTfulAdapter.lib_restful_interface("LteInfo", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, lte_statusinfo_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


LTE_STATUS_INFO_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify LTE status")
def rf_patch_system_lte_status_info():
    """
    功能描述：获取LTE的状态信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if LTE_STATUS_INFO_LOCK.locked():
        message = "Modify LTE status failed because LteView modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with LTE_STATUS_INFO_LOCK:
        message = "Modify LTE status failed"
        try:
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(LTEStatusInfoChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            request_data_dict['fd_server_ip'] = FdCfgManager.get_cur_fd_ip()
            # 获取资源模板
            lte_statusinfo_resource = json.loads(LteStatusInfoResourceSerializer().service.get_resource())
            # 调用接口LibAdapter获取 Softwares 相关信息
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            ret_dict = LibRESTfulAdapter.lib_restful_interface("LteInfo", "PATCH", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                return ret_dict, lte_statusinfo_resource

            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query LTE config info")
def rf_get_system_lte_config_info():
    """
    功能描述：获取LTE的配置信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query LTE config info failed."
    try:
        # 获取资源模板
        lte_configinfo_resource = json.loads(LteConfigInfoResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取 Shortwaves 相关信息
        ret_dict = LibRESTfulAdapter.lib_restful_interface("LteConfigInfo", "GET", None, False)

        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, lte_configinfo_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


LTE_CONFIG_INFO_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Set LTE config info")
def rf_patch_system_lte_config_info():
    """
    功能描述：获取LTE的配置信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if LTE_CONFIG_INFO_LOCK.locked():
        message = "Set LTE config info failed because LteView modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with LTE_CONFIG_INFO_LOCK:
        message = "Config LTE APN failed."
        try:
            # 取出请求数据
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(LTEConfigInfoChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO
            # 获取Success Message资源模板
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            # 调用接口LibAdapter获取 Softwares 相关信息
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            ret_dict = LibRESTfulAdapter.lib_restful_interface("LteConfigInfo", "PATCH", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": "Config LTE APN successfully."},
                }
                return ret_dict, success_message_resource
            elif ret_dict.get("message") == CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey:
                message = CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey
            ret_code = CommonConstants.ERR_CODE_400
            ret_dict = make_error_dict(message, ret_code)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
