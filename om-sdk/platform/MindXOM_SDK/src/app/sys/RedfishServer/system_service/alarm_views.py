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
"""
import json
import threading

from flask import request

from common.checkers.param_checker import AlarmShieldChecker
from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import AlarmInfoResourceSerializer
from system_service.systems_serializer import AlarmResourceSerializer
from system_service.systems_serializer import AlarmShieldResourceSerializer


INCREASE_ALARM_SHIELD = threading.Lock()
DECREASE_ALARM_SHIELD = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Query Alarm service resource")
def rf_get_system_alarm():
    """获取Alarm顶层的信息"""
    message = "Get Alarm service resource failed."
    try:
        # 获取 Alarm 顶层资源模板
        alarm_resource = json.loads(AlarmResourceSerializer().service.get_resource())
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict("ResourceDoesNotExists", CommonConstants.ERR_CODE_404)
        return ret_dict, CommonConstants.ERR_PARTICULAR_INFO
    return {"status": 200}, alarm_resource


@RedfishGlobals.redfish_operate_adapter(request, "Query Alarm info")
def rf_get_system_alarm_info():
    """查询告警资源信息"""
    message = "Get Alarm info failed."
    try:
        alarm_info_resource = json.loads(AlarmInfoResourceSerializer().service.get_resource())
        # 调用接口LibAdapter获取 Softwares 相关信息
        # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Alarm", "GET", None, False)
        RedfishGlobals.update_alarm_time_stamp(ret_dict)
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    return ret_dict, alarm_info_resource


@RedfishGlobals.redfish_operate_adapter(request, "Query Alarm shield")
def rf_get_system_alarm_shield():
    """查询告警屏蔽接口"""
    message = "Query masked alarms failed."
    try:
        alarm_shield_resource = json.loads(AlarmShieldResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("AlarmShield", "GET", None, False)
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    return ret_dict, alarm_shield_resource


@RedfishGlobals.redfish_operate_adapter(request, "Increase Alarm shield")
def rf_increase_system_alarm_shield():
    """
    功能描述：新增告警屏蔽接口
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if INCREASE_ALARM_SHIELD.locked():
        message = "Increase alarm shield failed because increasing alarm shield is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with INCREASE_ALARM_SHIELD:
        message = "Increase alarm shield failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(AlarmShieldChecker, request_data_dict)
            if check_ret is not None:
                run_log.error("%s because of invalid parameter", message)
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            req_alarm_shield_list = request_data_dict.get("AlarmShieldMessages")
            deduplicate_alarm_shield_list = []
            duplicate_alarm_shield_list = []
            for alarm in req_alarm_shield_list:
                if alarm not in deduplicate_alarm_shield_list:
                    deduplicate_alarm_shield_list.append(alarm)
                    continue

                duplicate_alarm_shield_list.append(alarm)

            if duplicate_alarm_shield_list:
                run_log.warning("Duplicate alarm shield rules exist in the request, "
                                "the system will remove the duplicate rules and continue.")

            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            request_data_dict['Type'] = "Increase"
            request_data_dict['AlarmShieldMessages'] = deduplicate_alarm_shield_list
            alarm_shield_resource = json.loads(AlarmShieldResourceSerializer().service.get_resource())
            # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给iBMA
            ret_dict = LibRESTfulAdapter.lib_restful_interface("AlarmShield", "PATCH", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                run_log.error("Increase alarm shield failed.")
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            message = "Increase alarm shield successfully."
            run_log.info(message)
            ret_dict = {
                "status": RedfishGlobals.SuccessfullyCode,
                "message": {"Message": message}
            }
            return ret_dict, success_message_resource

        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Decrease Alarm shield")
def rf_decrease_system_alarm_shield():
    """
    功能描述：减少告警屏蔽接口
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if DECREASE_ALARM_SHIELD.locked():
        message = "Decrease alarm shield failed because increasing alarm shield is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with DECREASE_ALARM_SHIELD:
        message = "Decrease alarm shield failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(AlarmShieldChecker, request_data_dict)
            if check_ret is not None:
                run_log.error("%s because of invalid parameter", message)
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            req_alarm_shield_list = request_data_dict.get("AlarmShieldMessages")
            deduplicate_alarm_shield_list = []
            duplicate_alarm_shield_list = []
            for alarm in req_alarm_shield_list:
                if alarm not in deduplicate_alarm_shield_list:
                    deduplicate_alarm_shield_list.append(alarm)
                    continue

                duplicate_alarm_shield_list.append(alarm)

            if duplicate_alarm_shield_list:
                run_log.warning("Duplicate alarm shield rules exist in the request, "
                                "the system will remove the duplicate rules and continue.")

            request_data_dict['_User'] = RedfishGlobals.USER_NAME
            request_data_dict['_Xip'] = request.headers['X-Real-Ip']
            request_data_dict['Type'] = "Decrease"
            request_data_dict['AlarmShieldMessages'] = deduplicate_alarm_shield_list
            alarm_shield_resource = json.loads(AlarmShieldResourceSerializer().service.get_resource())
            # 根据调用接口LibAdapter返回的retDict字典，生成对应的json,返回给iBMA
            ret_dict = LibRESTfulAdapter.lib_restful_interface("AlarmShield", "PATCH", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                run_log.error("Decrease alarm shield failed.")
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            message = "Decrease alarm shield successfully."
            run_log.info(message)
            ret_dict = {
                "status": RedfishGlobals.SuccessfullyCode,
                "message": {"Message": message}
            }
            return ret_dict, success_message_resource

        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
