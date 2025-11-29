# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import threading

from flask import request

from common.checkers.param_checker import CreatePatitionChecker, MountPartitionChecker, UnmountPartitionChecker
from common.checkers.param_checker import PartitionIdChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import PartitionCollectionResourceSerializer
from system_service.systems_serializer import PartitionResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query system partitions")
def rf_get_system_partitions():
    """
    功能描述：获取  System 磁盘分区信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2019-4-1 创建
    """
    message = "Query system partitions failed."
    try:
        # 获取资源模板
        partiton_collection_resource = json.loads(PartitionCollectionResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, True)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, partiton_collection_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


CREATE_PARTITION_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Create system partition")
def rf_create_system_partitions():
    """
    功能描述：创建 System 磁盘分区信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2019-4-1 创建
    """
    if CREATE_PARTITION_LOCK.locked():
        message = "Create system partition failed because PartitionView modify is busy."
        run_log.error(f"{message}")
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with CREATE_PARTITION_LOCK:
        message = "Create system partition failed."
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(CreatePatitionChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            # 获取Username和requestIp放入字典request_data_Dict
            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            # 获取 SuccessMessasge 资源模板
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            ret = LibRESTfulAdapter.lib_restful_interface("Partition", "POST", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)

            if ret_status_is_ok:
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": "Create system partition successfully."},
                }
                return ret_dict, success_message_resource

            return ret, CommonConstants.ERR_GENERAL_INFO

        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query partition info")
def rf_get_system_partition_info(partition_id):
    """
    功能描述：获取指定磁盘分区信息
    参数：partition_id
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2019-4-1 创建
    """
    message = "Query partition (patition_id: {}) failed.".format(partition_id)
    try:
        check_ret = RedfishGlobals.check_external_parameter(PartitionIdChecker, {"partition_id": partition_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO

        # 获取资源模板
        partition_resource = json.loads(PartitionResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Partition", "GET", None, False, partition_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)

        if ret_status_is_ok:
            RedfishGlobals.replace_id(partition_resource, partition_id)
            return ret_dict, partition_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


MOUNT_PARTITION_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Mount partition")
def rf_mount_system_partition():
    """
    功能描述：挂载指定磁盘分区
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2023-2-9 创建
    """
    if MOUNT_PARTITION_LOCK.locked():
        message = "Mount partition failed because Partition View mount is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with MOUNT_PARTITION_LOCK:
        try:
            recv_conf_dict = request.get_data()

            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(MountPartitionChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict["_User"] = RedfishGlobals.USER_NAME
            request_data_dict["_Xip"] = request.headers["X-Real-Ip"]
            partition_id = request_data_dict.get("PartitionID")
            del request_data_dict["PartitionID"]

            ret = LibRESTfulAdapter.lib_restful_interface("Partition", "PATCH", request_data_dict, False, partition_id)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)

            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = f"Monut partition [{partition_id}] successfully."
                run_log.info(message)
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message}
                }
                return ret_dict, success_message_resource

            message = f"Monut partition [{partition_id}] failed."
            run_log.error(message)
            return ret, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


UNMOUNT_PARTITION_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Unmount partition")
def rf_unmount_system_partition():
    """
    功能描述：解挂指定磁盘分区
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2023-2-9 创建
    """
    if UNMOUNT_PARTITION_LOCK.locked():
        message = "Unmount partition failed because Partition View unmount is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with UNMOUNT_PARTITION_LOCK:
        try:
            recv_conf_dict = request.get_data()

            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(UnmountPartitionChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict["_User"] = RedfishGlobals.USER_NAME
            request_data_dict["_Xip"] = request.headers["X-Real-Ip"]
            partition_id = request_data_dict.pop("PartitionID")
            # MountPath为空表示解挂
            request_data_dict["MountPath"] = ""
            ret = LibRESTfulAdapter.lib_restful_interface(
                "Partition", "PATCH", request_data_dict, False, partition_id
            )
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)

            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = f"Unmount partition [{partition_id}] successfully."
                run_log.info(message)
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message}
                }
                return ret_dict, success_message_resource

            message = f"Unmount partition [{partition_id}] failed."
            run_log.error(message)
            return ret, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Delete partition")
def rf_delete_system_partition(partition_id):
    """
    功能描述：删除指定磁盘分区
    参数：partition_id
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    记录：2019-4-1 创建
    """
    message = "Delete partition failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(PartitionIdChecker, {"partition_id": partition_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO

        # 获取Username和requestIp放入字典request_data_Dict
        request_data_dict = {"_User": RedfishGlobals.USER_NAME, "_Xip": request.headers["X-Real-Ip"]}
        ret = LibRESTfulAdapter.lib_restful_interface("Partition", "DELETE", request_data_dict, False, partition_id)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret)
        if ret_status_is_ok:
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            message = f"Delete partition [{partition_id}] successfully."
            ret_dict = {
                "status": RedfishGlobals.creation_successfully_code,
                "message": {"Message": message},
            }
            run_log.info(message)
            return ret_dict, success_message_resource

        return ret, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

