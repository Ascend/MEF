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
import os
import shutil
import threading

from flask import request
from flask import send_from_directory

from common.constants.error_codes import CommonErrorCodes
from common.constants.product_constants import LOG_COLLECT_LIST
from common.file_utils import FileCreate, FileUtils
from common.log.logger import run_log
from common.checkers.param_checker import LogServiceDownloadChecker
from common.constants.error_codes import LogErrorCodes
from common.constants.base_constants import LoggerConstants, CommonConstants
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_serializer import LogServiceResourceSerializer, LogCollectProgressResourceSerializer


@RedfishGlobals.redfish_operate_adapter(request, "Query log services")
def rf_system_logservices():
    """
    功能描述：获取日志信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query log services failed."
    try:
        # 获取log_rervice模板
        log_service_resource = json.loads(LogServiceResourceSerializer().service.get_resource())
        log_servers_list = LOG_COLLECT_LIST
        ret_dict = {'status': 200, 'message': log_servers_list}
        return ret_dict, log_service_resource
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": input_err_info}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


LOG_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Query the progress of log collecting")
def rf_system_log_collect_progress():
    """
    功能描述：获取日志收集进度
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    input_err_info = "Query the progress of log collecting failed."
    try:
        log_collect_progress_resource = json.loads(LogCollectProgressResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "GET", None, False)

        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, log_collect_progress_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": input_err_info}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Collect log")
def rf_system_log_download():
    """
    功能描述：日志收集
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if LOG_LOCK.locked():
        message = "Collect log failed because lock is locked."
        run_log.error(f"{message}")
        ret_dict = LoggerCollectUtils.make_log_error_dict(LogErrorCodes.ERROR_LOG_COLLECT, "Collect log failed.")
        return RedfishGlobals.update_json_of_error_info(ret_dict, CommonConstants.ERR_GENERAL_INFO)
    with LOG_LOCK:
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return RedfishGlobals.update_json_of_error_info(ret_dict, CommonConstants.ERR_PARTICULAR_INFO)

            # 获取用户的入参并判断是否在指定的范围内
            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(LogServiceDownloadChecker, request_data_dict)
            if check_ret is not None:
                return RedfishGlobals.update_json_of_error_info(check_ret, CommonConstants.ERR_GENERAL_INFO)

            if os.path.exists(LoggerConstants.WEB_LOG_COLLECT_PATH):
                os.remove(LoggerConstants.WEB_LOG_COLLECT_PATH)

            if not FileCreate.create_dir(LoggerConstants.REDFISH_LOG_COLLECT_DIR, 0o1700):
                run_log.error("Collect log failed because create collect_log_dir failed.")
                ret_dict = LoggerCollectUtils.make_log_error_dict(LogErrorCodes.ERROR_LOG_COLLECT, "Collect log failed")
                return RedfishGlobals.update_json_of_error_info(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            ret_dict = LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "POST", request_data_dict, False)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                run_log.error("Collect log failed, om core operate failed")
                ret_dict = LoggerCollectUtils.make_log_error_dict(LogErrorCodes.ERROR_LOG_COLLECT, "Collect log failed")
                return RedfishGlobals.update_json_of_error_info(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            # 将文件拷贝到MindXOM用户目录下，提供下载，并通知monitor进程删除临时日志文件
            shutil.copyfile(LoggerConstants.MONITOR_TMP_COLLECT_LOG, LoggerConstants.WEB_LOG_COLLECT_PATH)
            LibRESTfulAdapter.lib_restful_interface("LoggerCollect", "DELETE", None)
            ret = send_from_directory(LoggerConstants.REDFISH_LOG_COLLECT_DIR,
                                      os.path.basename(LoggerConstants.WEB_LOG_COLLECT_PATH), as_attachment=True)
            run_log.info("Collect log successfully.")
            return ret

        except Exception as err:
            run_log.error("Collect log failed reason, %s", err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "Collect log failed"}
            return RedfishGlobals.update_json_of_error_info(ret_dict, CommonConstants.ERR_GENERAL_INFO)

        finally:
            FileUtils.delete_file_or_link(LoggerConstants.WEB_LOG_COLLECT_PATH)


class LoggerCollectUtils:

    @staticmethod
    def make_log_error_dict(error_code: CommonErrorCodes, input_err_info):
        if not input_err_info:
            input_err_info = error_code.messageKey
        # 记录操作日志
        run_log.info(input_err_info)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [error_code.code, error_code.messageKey]
        }
        return ret_dict
