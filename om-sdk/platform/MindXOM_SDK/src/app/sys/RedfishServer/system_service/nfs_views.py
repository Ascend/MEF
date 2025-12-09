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

from common.checkers.param_checker import NfsManagerChecker
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import NfsServiceErrorCodes
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import NfsManageServiceResourceSerializer

MOUNT_NFS_LOCK = threading.Lock()
UNMOUNT_NFS_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Query NFS manage info")
def rf_get_system_nfs_manage():
    """
    功能描述：处理查询 Nfs 分区信息请求
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2019-8-17 创建
    """
    message = "Query NFS info failed."
    try:
        # 获取资源模板
        nfs_manage_resource = json.loads(NfsManageServiceResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("NfsManage", "GET", None, False)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            return ret_dict, nfs_manage_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error(f"{message} reason is: {err}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Mount NFS request")
def rf_mount_system_nfs_manage():
    """
    功能描述：挂载NFS分区
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2023-2-9 创建
    """
    if MOUNT_NFS_LOCK.locked():
        message = "Mount NFS request failed because lock is locked."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [NfsServiceErrorCodes.OPERATE_IS_BUSY.code, message]
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with MOUNT_NFS_LOCK:
        message = "Mount NFS request failed."
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(NfsManagerChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            request_data_dict['Type'] = "mount"
            # 获取 Success Message 的资源模板
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            ret = LibRESTfulAdapter.lib_restful_interface("NfsManage", "POST", request_data_dict, False)
            if LibRESTfulAdapter.check_status_is_ok(ret):
                status_code = RedfishGlobals.SuccessfullyCode
                message = "Mount NFS successfully."
                ret_dict = {"status": status_code, "message": {"Message": message}}
                return ret_dict, success_message_resource

            return ret, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error(f"{message} reason is: {err}")
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Unmount NFS request")
def rf_unmount_system_nfs_manage():
    """
    功能描述：解挂NFS分区
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2023-2-9 创建
    """
    if UNMOUNT_NFS_LOCK.locked():
        message = "Unmount NFS request failed because lock is locked."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [NfsServiceErrorCodes.OPERATE_IS_BUSY.code, message]
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with UNMOUNT_NFS_LOCK:
        message = "Unmount NFS request failed."
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(NfsManagerChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            request_data_dict['Type'] = "umount"
            # 获取 Success Message 的资源模板
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            ret = LibRESTfulAdapter.lib_restful_interface("NfsManage", "POST", request_data_dict, False)
            if LibRESTfulAdapter.check_status_is_ok(ret):
                status_code = RedfishGlobals.SuccessfullyCode
                message = "Unmount NFS successfully."
                ret_dict = {"status": status_code, "message": {"Message": message}}
                return ret_dict, success_message_resource

            return ret, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error(f"{message} reason is: {err}")
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
