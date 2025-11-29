# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

from flask import request

from common.checkers.param_checker import StoragesInfoChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import SimpleStorageCollectionResourceSerializer, SimpleStorageResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query simple storages")
def rf_get_system_simple_storages_collection():
    """
    功能描述：获取 Systems Simple Storages 信息
    参数：无
    返回值：http response
    异常描述：NA
    记录：2019-4-1 创建
    """
    input_err_info = "Query simple storages failed."
    try:
        # 获取资源模板
        simple_storage_collection_resource = json.loads(SimpleStorageCollectionResourceSerializer().
                                                        service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Storage", "GET", None, True)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, simple_storage_collection_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query simple storages by storage_id")
def rf_get_system_storage_info(storage_id):
    """
    功能描述：获取  Systems Simple Storages  信息
    参数：无
    返回值：http response
    异常描述：NA
    记录：2019-4-1 创建
    """
    input_err_info = "Query simple storages by storage_id failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(StoragesInfoChecker, {"storage_id": storage_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO

        # 获取资源模板
        simple_storage_resource = json.loads(SimpleStorageResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Storage", "GET", None, False, storage_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            RedfishGlobals.replace_id(simple_storage_resource, storage_id)
            return ret_dict, simple_storage_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
