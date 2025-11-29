# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import os
import shutil

from flask import request

from common.checkers.param_checker import UpdateServiceChecker
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import CommonErrorCodes, FileErrorCodes
from common.constants.upload_constants import UploadConstants
from common.checkers.param_checker import UpdateServiceResetChecker
from common.file_utils import FileUtils
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from upgrade_service.upgrade_serializer import GetUpgradeServiceResourceSerializer, GetUpgradeServiceActionSerializer
from upload_mark_file import UploadMarkFile


@RedfishGlobals.redfish_operate_adapter(request, "Get upgrade service resource")
def get_upgrade_service_resource():
    """
    功能描述：查询升级固件服务资源信息
    参数：无
    :return:http response
    """
    try:
        resp_json = json.loads(GetUpgradeServiceResourceSerializer().service.get_resource())
    except Exception as err:
        err_message = "get upgrade service resource failed."
        run_log.error("%s reason is: %s", err_message, err)
        return {"status": CommonConstants.ERR_CODE_404,
                "message": "ResourceDoesNotExists"}, CommonConstants.ERR_PARTICULAR_INFO
    else:
        run_log.info("get upgrade service resource success.")
        return {"status": CommonConstants.ERR_CODE_200}, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Get upgrade task info")
def get_upgrade_service_actions():
    try:
        # 获取资源模板
        resp_json = json.loads(GetUpgradeServiceActionSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Upgrade_New", "GET", None, False)
        # 根据调用接口LibAdapter返回的retDict字典，生产对应的json
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            run_log.error("get upgrade task info failed.")
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        return ret_dict, resp_json
    except Exception as err:
        err_message = "get upgrade task info failed."
        run_log.error(f"{err_message} reason is: {err}")
        return {"status": CommonConstants.ERR_CODE_400,
                "message": RedfishGlobals.internalErrInfo}, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Start upgrade task")
def rf_upgrade_service_actions():
    if RedfishGlobals.high_risk_exclusive_lock.locked():
        run_log.error("System is busy, operate failed.")
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with RedfishGlobals.high_risk_exclusive_lock:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        if not isinstance(ret_dict.get("message"), dict) or not \
                isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                ret_dict.get("message").get("system_busy"):
            run_log.error("System is busy, operate failed.")
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                return ret[1], CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(UpdateServiceChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO
            # 获取Username和requestIp放入字典request_data_Dict
            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            package_name = request_data_dict.get("ImageURI")
            # 将升级包文件名固定
            old_zipfile_path = os.path.join(UploadConstants.ZIP_UPLOAD_DIR, package_name)
            new_zipfile_path = os.path.join(UploadConstants.ZIP_UPLOAD_DIR, UploadConstants.WEB_DOWNLOAD_PACKAGE_NAME)
            if not os.path.exists(old_zipfile_path):
                ret_dic = {
                    "status": CommonConstants.ERR_CODE_400,
                    "message": [FileErrorCodes.ERROR_FILE_NOT_EXIST_ERROR.code, f"{package_name} not exist"]
                }
                return ret_dic, CommonConstants.ERR_GENERAL_INFO
            shutil.move(old_zipfile_path, new_zipfile_path)
            # 将升级包的完整路径添加到字典中
            request_data_dict["ImageURI"] = new_zipfile_path
            # 获取资源模板
            resp_json = json.loads(GetUpgradeServiceActionSerializer().service.get_resource())
            ret_dict = LibRESTfulAdapter.lib_restful_interface("Upgrade_New", "POST", request_data_dict, False)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            run_log.info("Start upgrade task successfully, the upgrade task is running in the background.")
            return ret_dict, resp_json
        except Exception as err:
            message = "Start upgrade task failed."
            run_log.error(f"{message} reason is: {err}")
            return {"status": CommonConstants.ERR_CODE_400,
                    "message": RedfishGlobals.internalErrInfo}, CommonConstants.ERR_GENERAL_INFO
        finally:
            # 业务完成后清楚文件上传标记（必须）
            UploadMarkFile.clear_upload_mark_file(UploadConstants.MARK_FILES.get("zip"))
            FileUtils.delete_dir_content(UploadConstants.ZIP_UPLOAD_DIR)


@RedfishGlobals.redfish_operate_adapter(request, "Effect firmware")
def rf_upgrade_reset_actions():
    """升级生效接口"""
    if RedfishGlobals.high_risk_exclusive_lock.locked():
        run_log.error("System is busy, operate failed.")
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with RedfishGlobals.high_risk_exclusive_lock:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        if not isinstance(ret_dict.get("message"), dict) or not \
                isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                ret_dict.get("message").get("system_busy"):
            run_log.error("System is busy, operate failed.")
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                run_log.error("request data is not json or request data is none")
                return ret[1], CommonConstants.ERR_PARTICULAR_INFO
            else:
                request_data_dict = ret[1]
                check_ret = RedfishGlobals.check_external_parameter(UpdateServiceResetChecker, request_data_dict)
                if check_ret is not None:
                    run_log.error("check external parameter failed")
                    return check_ret, CommonConstants.ERR_GENERAL_INFO

                resp_json = json.loads(GetUpgradeServiceActionSerializer().service.get_resource())
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Upgrade_effect", "POST", request_data_dict, False)
                if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                    run_log.error("Effect firmware failed.")
                    return ret_dict, CommonConstants.ERR_GENERAL_INFO

                run_log.info("Effect firmware successfully")
                return ret_dict, resp_json
        except Exception as err:
            run_log.error("Effect firmware failed. reason is:%s", err)
            return {"status": CommonConstants.ERR_CODE_400,
                    "message": RedfishGlobals.internalErrInfo}, CommonConstants.ERR_GENERAL_INFO
