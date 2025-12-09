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

from flask import request

from common.checkers.param_checker import ExtendedDeviceChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import ExtendedDevicesCollectionResourceSerializer, \
    ExtendedDevicesMemberResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query system extended device resources")
def rf_system_extended_devices_collection():
    """
    功能描述：获取外设的集合信息
    参数：无
    返回值：http response
    异常描述：NA
    记录：2019-3-26 创建
    """
    input_err_info = "Query system extended device resources failed."
    try:
        # 获取资源模板
        extended_devices_collection_resource = json.loads(ExtendedDevicesCollectionResourceSerializer(). \
                                                          service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Extend", "GET", None, True)

        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, extended_devices_collection_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{input_err_info} reason is: {err}")
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query extended device by extend id")
def rf_system_extended_device_info(extend_id):
    """
    功能描述：获取外设信息
    参数：无
    返回值：http response
    异常描述：NA
    记录：2019-3-26 创建
    """
    input_err_info = "Query extended device by extend id failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(ExtendedDeviceChecker, {"extend_id": extend_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO

        # 获取资源模板
        extended_devices_member_resource = json.loads(ExtendedDevicesMemberResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取ethId相关信息
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Extend", "GET", None, False, extend_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)

        if ret_status_is_ok:
            RedfishGlobals.replace_id(extended_devices_member_resource, extend_id)
            return ret_dict, extended_devices_member_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{input_err_info} reason is: {err}")
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
