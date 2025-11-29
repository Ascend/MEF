# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import threading

from flask import request, g

from account_service.account_serializer import AccountServiceResourceSerializer
from account_service.account_serializer import AccountsMembersResourceSerializer
from account_service.account_serializer import AccountsResourceSerializer
from common.checkers.param_checker import AccountServiceInfoChecker
from common.checkers.param_checker import ChangeNameAndPasswordChecker
from common.checkers.param_checker import MemberIdChecker
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import UserManagerConstants
from common.constants.error_codes import CommonErrorCodes
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from user_manager.user_manager import UserManager


@RedfishGlobals.redfish_operate_adapter(request, "Query password expiration days")
def rf_get_account_password_expiration_days():
    """
    功能描述：查询用户服务信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-28 创建
    """
    message = "Query password validity failed."
    try:
        # 获取 AccountService资源模板
        account_service_resource = json.loads(AccountServiceResourceSerializer().service.get_resource())
        recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_ACCOUNT_EXPIRATION_DAY}
        ret_dict = UserManager().get_all_info(recv_conf_dict)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, account_service_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": message,
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


ACCOUNT_SERVICE_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify user service info")
def rf_modify_account_password_expiration_days():
    """
    功能描述：修改用户服务信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-28 创建
    """
    if ACCOUNT_SERVICE_LOCK.locked():
        message = "Modify user account info failed because AccountViews modify is busy."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED, message],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with ACCOUNT_SERVICE_LOCK:
        message = "Modify user account info failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                input_err_info = "Request data is not json."
                ret_dict = ret[1]
                run_log.error(input_err_info)
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(AccountServiceInfoChecker, request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict["oper_type"] = UserManagerConstants.OPER_TYPE_MODIFY_ACCOUNT_EXPIRATION_DAY
            request_data_dict["user_id"] = g.user_id
            # 获取 AccountService资源模板
            account_service_resource = json.loads(AccountServiceResourceSerializer().service.get_resource())
            ret_dict = UserManager().patch_request(request_data_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                return ret_dict, account_service_resource

            if RedfishGlobals.get_user_locked_state(ret_dict):
                message = CommonConstants.USER_LOCK_STATE
                RedfishGlobals.set_operational_log(request=request, username=RedfishGlobals.USER_NAME, message=message)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": message,
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query user account info collection")
def rf_get_account_info_collection():
    """
    功能描述:查询用户集合信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2022-11-14创建
    """
    message = "Query user account collection failed."
    try:
        # 获取Accounts模板
        accounts_resource = json.loads(AccountsResourceSerializer().service.get_resource())
        # 调用插件层获取返回信息
        recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_ID_LIST}
        ret_dict = UserManager().get_all_info(recv_conf_dict)
        # 判断返回的字典是否正确
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            ret_dict = {
                "status": ret_dict.get("status"),
                "message": ret_dict.get("message").get("result"),
            }
            return ret_dict, accounts_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": message,
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query specified account info")
def rf_get_specified_account_info(member_id):
    """
    功能描述：查询指定用户资源信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-22 创建
    """
    input_err_info = "Query specified account info failed."
    try:
        check_ret = RedfishGlobals.check_external_parameter(MemberIdChecker, {"member_id": member_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        # 获取 AccountsMembers 资源模板
        accounts_members_resource = json.loads(AccountsMembersResourceSerializer().service.get_resource())
        # 调用插件层获取返回信息
        recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_USER_INFO, "user_id": member_id}
        ret_dict = UserManager().get_all_info(recv_conf_dict)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, accounts_members_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": input_err_info,
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


ACCOUNTS_MEMBERS_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify specified account info")
def rf_modify_specified_account_info(member_id):
    """
    功能描述：修改指定用户信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-24 创建
    """
    if ACCOUNTS_MEMBERS_LOCK.locked():
        message = "Modify specified account info failed because AccountViews modify is busy."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code, message],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with ACCOUNTS_MEMBERS_LOCK:
        message = "Modify account info failed."
        try:
            # member_id校验
            check_ret = RedfishGlobals.check_external_parameter(MemberIdChecker, {"member_id": member_id})
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO
            # 取出数据
            recv_conf_dict = request.get_data()
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(ChangeNameAndPasswordChecker, request_data_dict)
            if check_ret is not None:
                ret = {"status": 400, "message": [110207, "The user name or password error."]}
                return ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict["oper_type"] = UserManagerConstants.OPER_TYPE_MODIFY_USER_INFO
            request_data_dict["user_id"] = member_id
            update_name = request_data_dict.get("UserName", None)
            if update_name != RedfishGlobals.USER_NAME:
                ret = {"status": 400, "message": [110207, "The user name or password error."]}
                return ret, CommonConstants.ERR_GENERAL_INFO

            # 获取 AccountsMembers 资源模板
            accounts_members_resource = json.loads(AccountsMembersResourceSerializer().service.get_resource())
            ret_dict = UserManager().patch_request(request_data_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                return ret_dict, accounts_members_resource

            if RedfishGlobals.get_user_locked_state(ret_dict):
                message = CommonConstants.USER_LOCK_STATE
                RedfishGlobals.set_operational_log(request=request, username=RedfishGlobals.USER_NAME, message=message)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": message,
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
