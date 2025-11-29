# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Ethernet资源 URL处理模块
修改记录：2022-10-29 创建
"""
import json
import threading
import time

from flask import request

from common.checkers.param_checker import EthernetInterfaceIdChecker, EthernetInterfaceChecker
from common.constants.error_codes import CommonErrorCodes
from common.log.logger import run_log
from common.constants.base_constants import CommonConstants
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_common import modify_null_tag_to_default
from system_service.systems_serializer import EthernetCollectionResourceSerializer
from system_service.systems_serializer import EthernetGetMembersResourceSerializer
from system_service.systems_serializer import EthernetPatchMembersResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query ethernet interface resources")
def rf_get_system_ethernet_collection():
    """
    功能描述：获取以太网接口集合信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2016-11-22 创建
    修改记录：2016-12-07 增加TeamAndBridge对应处理
    """
    input_err_info = "Query ethernet interface resources failed."
    try:
        # 获取EthernetInterfaces资源模板
        ethernet_collection_resource = json.loads(EthernetCollectionResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取EthernetInterfaces相关信息
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Nic", "GET", None, True)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, ethernet_collection_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Get ethernet interface info")
def rf_get_system_ethernet_eth_x(eth_id):
    """
    功能描述：查询以太网接口资源信息
    参数：eth_id
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2016-11-22 创建
    """
    message = "Get ethernet interface info failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(EthernetInterfaceIdChecker, {"eth_id": eth_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        # 获取get ethernet member资源模板
        ethernet_getmembers_resource = json.loads(EthernetGetMembersResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取ethId相关信息
        ret = LibRESTfulAdapter.lib_restful_interface("Nic", "GET", None, False, eth_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)
        if ret_status_is_ok:
            RedfishGlobals.replace_id(ethernet_getmembers_resource, eth_id)
            return ret, ethernet_getmembers_resource

        return ret, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


MODIFY_ETHERNET_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify ethernet interface")
def rf_modify_system_ethernet_eth_x(eth_id):
    """
    功能描述：修改以太网接口资源信息
    参数：eth_id
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2016-11-22 创建
            2019-3-10 添加 patch 请求
    """
    if MODIFY_ETHERNET_LOCK.locked():
        message = "Modify ethernet failed because EthernetViews modify is busy."
        run_log.error(message)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, message]}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with MODIFY_ETHERNET_LOCK:
        message = f"Modify ethernet interface (ethId:[{eth_id}]) failed."
        try:
            check_ret = RedfishGlobals.check_external_parameter(EthernetInterfaceIdChecker, {"eth_id": eth_id})
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            check_ret = RedfishGlobals.check_external_parameter(EthernetInterfaceChecker, ret[1])
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict = modify_null_tag_to_default(ret[1])
            starttime = time.strftime('%Y-%m-%dT%H:%M:%S%z', time.localtime(time.time()))
            # 获取patch ethernet member资源模板
            ethernet_patchmembers_resource = json.loads(EthernetPatchMembersResourceSerializer().service.get_resource())
            ret = LibRESTfulAdapter.lib_restful_interface("Nic", "PATCH", request_data_dict, False, eth_id)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)
            if not ret_status_is_ok:
                return ret, CommonConstants.ERR_GENERAL_INFO

            task_percentage = 'ok'
            if isinstance(ret.get("message"), dict) and ret.get("message").get("task_percentage"):
                task_percentage = ret.get("message").get("task_percentage")
            RedfishGlobals.replaceinfo(ethernet_patchmembers_resource, eth_id, starttime,
                                       request_data_dict, task_percentage)
            return ret, ethernet_patchmembers_resource
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
