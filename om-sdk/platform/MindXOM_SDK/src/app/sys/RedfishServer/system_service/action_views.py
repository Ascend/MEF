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
功    能：系统资源 URL处理模块
修改记录：2022-11-26 创建
"""
import json
import sys

from flask import request

from common.checkers.param_checker import ComputerSystemResetChecker
from common.checkers.param_checker import ResetOpChecker
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import CommonErrorCodes
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict


@RedfishGlobals.redfish_operate_adapter(request, "Restart system")
def rf_system_reset():
    """
    功能描述：复位系统操作
    return: 响应字典 资源模板或错误消息
    """
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
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO
            else:
                request_data_dict = ret[1]
                check_ret = RedfishGlobals.check_external_parameter(ComputerSystemResetChecker, request_data_dict)
                if check_ret is not None:
                    return check_ret, CommonConstants.ERR_GENERAL_INFO

                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                action = request_data_dict.get("ResetType", None)
                # 注意：此接口内部用定时器延迟重启，所以总能收到返回值
                ret_dict = LibRESTfulAdapter.lib_restful_interface("Actions", "POST", request_data_dict, False)
                ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
                # 适配smartkit
                # smartkit只接收500错误，使用sys.exit()可保证一定返回500错误
                # 若不用sys.exit()，则内部定时器走到系统重启指令时，可能nginx先于ibma-edge被杀掉，那么就会返回502错误，
                # 不符合smartkit要求
                if ret_status_is_ok and action in ("GracefulRestart",):
                    # 执行完sys.exit()不会进到装饰器最后打印日志的部分，因此先记录日志
                    operation_log = "Restart system (GracefulRestart) successfully."
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, operation_log)
                    sys.exit()
                if ret_status_is_ok:
                    ret_dict = {
                        "status": RedfishGlobals.SuccessfullyCode,
                        "message": {"Message": f"Restart system ({action}) successfully."}
                    }
                    return ret_dict, success_message_resource
                return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("Restart system failed, error: %s", err)
            ret_dict = make_error_dict("Restart system failed", CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Restore defaults system")
def rf_restore_defaults():
    """
    功能描述：恢复出厂设置
    return: 响应字典 资源模板或错误消息
    """
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
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO
            else:
                request_data_dict = ret[1]
                check_ret = RedfishGlobals.check_external_parameter(ResetOpChecker, request_data_dict)
                if check_ret is not None:
                    return check_ret, CommonConstants.ERR_GENERAL_INFO

                request_data_dict['_User'] = RedfishGlobals.USER_NAME
                request_data_dict['_Xip'] = request.headers['X-Real-Ip']
                # 获取资源模板
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                ret_dict = LibRESTfulAdapter.lib_restful_interface(
                    "RestoreDefaultsAction", "POST", request_data_dict, False)
                ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
                if ret_status_is_ok:
                    ret_dict = {
                        "status": RedfishGlobals.SuccessfullyCode,
                        "message": {"Message": f"Restore defaults system successfully."}
                    }
                    return ret_dict, success_message_resource
                return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("Restore defaults system failed. reason is %s", err)
            ret_dict = make_error_dict("Restore defaults system failed.", CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
