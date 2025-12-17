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
功    能：系统资源模组管理 URL处理模块
"""
import json

from flask import request

from common.checkers.param_checker import DeviceChecker
from common.checkers.param_checker import ModuleChecker
from common.constants.error_codes import CommonErrorCodes
from lib_restful_adapter import LibRESTfulAdapter
from ibma_redfish_globals import RedfishGlobals
from common.log.logger import run_log
from common.constants.base_constants import CommonConstants
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import ModuleCollectionResourceSerializer, ModuleInfoResourceSerializer, \
    DeviceInfoResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query system module collection")
def rf_get_system_module_collection():
    """
    功能描述：查询拓展模组资源集合信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query system module collection failed."
    try:
        modules_resource = json.loads(ModuleCollectionResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Module", "GET", None, True)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            return ret_dict, modules_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is:%s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query system module info")
def rf_get_system_module_info(module_id):
    """
    功能描述：查询模组详细信息
    参数：module_id 为模组名称
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query system module info failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(ModuleChecker, {"module_id": module_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        module_info_resource = json.loads(ModuleInfoResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Module", "GET", None, False, module_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            RedfishGlobals.replace_id(module_info_resource, module_id)
            ret_dict_tmp = ret_dict["message"]["devices"]
            module_info_resource = RedfishGlobals.replace_mem_count_kv_list(module_info_resource,
                                                                            ret_dict_tmp, module_id)
            return ret_dict, module_info_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is:%s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query system device info")
def rf_get_system_device_info(module_id, device_id):
    """
    功能描述：查询设备属性
    参数：module_id 为模组名称，device_id为设备名称
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query system device info failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(DeviceChecker, {"module_id": module_id,
                                                                            "device_id": device_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        device_resource = json.loads(DeviceInfoResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Device", "GET", None, False, module_id, device_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            module_replace_odata_id(device_resource, module_id, device_id)
            return ret_dict, device_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is:%s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Modify system device resources")
def rf_patch_system_device_info(module_id, device_id):
    """
    功能描述：修改设备属性
    参数：module_id 为模组名称，device_id为设备名称
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Modify system device info failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(DeviceChecker, {"module_id": module_id,
                                                                            "device_id": device_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        # 取出数据
        recv_conf_dict = request.get_data()
        # 校验数据是否为json
        ret = RedfishGlobals.check_json_request(recv_conf_dict)
        if ret[0] != 0:
            ret_dict = ret[1]
            return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

        request_data_dict = ret[1]
        if not isinstance(request_data_dict.get("Attributes"), dict):
            run_log.error("Attributes param type is wrong")
            ret_dict = {
                "status": 400,
                "message": [
                    CommonErrorCodes.ERROR_PARAMETER_INVALID.code,
                    CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey
                ]
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        # 获取Username和requestIp放入字典request_data_Dict
        request_data_dict["_User"] = RedfishGlobals.USER_NAME
        request_data_dict["_Xip"] = request.headers["X-Real-Ip"]
        # 获取 System Collection资源模板
        device_resource = json.loads(DeviceInfoResourceSerializer().service.get_resource())
        # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Device", "PATCH", request_data_dict,
                                                           False, module_id, device_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            module_replace_odata_id(device_resource, module_id, device_id)
            return ret_dict, device_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is:%s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


def module_replace_odata_id(resp_json, module_id, device_id):
    """
    功能描述：替换模板中的ID
    参数：resp_json JSON 模板
         module_id and device_id 替换的ID
    异常描述：无
    """
    resp_json['Id'] = device_id
    resp_json['@odata.id'] = resp_json['@odata.id'].replace('oDataID1', module_id)
    resp_json['@odata.id'] = resp_json['@odata.id'].replace('oDataID2', device_id)
    resp_json['@odata.context'] = resp_json['@odata.context'].replace('oDataID1', module_id)
