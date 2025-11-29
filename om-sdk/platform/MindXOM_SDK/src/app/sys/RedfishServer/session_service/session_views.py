# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import functools
import json
import threading

from flask import current_app
from flask import request, g

from common.checkers.param_checker import CreateSessionServiceChecker
from common.checkers.param_checker import DeleteServiceChecker
from common.checkers.param_checker import SessionServiceInfoChecker
from common.constants.base_constants import UserManagerConstants, CommonConstants
from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import BizException
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals, USER_LOCK_STATE
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from session_service.session_serializer import SessionServiceResourceSerializer
from session_service.session_serializer import SessionsMembersResourceSerializer
from user_manager.user_manager import SessionManager, UserManager


@RedfishGlobals.redfish_operate_adapter(request, "Query timeout period failed")
def rf_get_session_service_collection():
    """
    功能描述：查询会话服务信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-28 创建
    """
    message = "Get timeout period failed."
    try:
        # 获取session service资源模板
        session_service_resource = json.loads(SessionServiceResourceSerializer().service.get_resource())
        message = "Failed to query session information."
        # 调用插件层接口获取session time out的时间
        recv_conf_dict = {"oper_type": UserManagerConstants.OPER_TYPE_GET_SESSION_TIMEOUT}
        ret_dict = SessionManager().get_all_info(recv_conf_dict)
        # 判断返回的字典是否正确，如果正确，填充模板
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        if ret_status_is_ok:
            return ret_dict, session_service_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": message,
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


SESSION_SERVICE_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Modify session info")
def rf_patch_session_service_collection():
    """
    功能描述：修改会话服务信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-28 创建
    """
    if SESSION_SERVICE_LOCK.locked():
        message = "Modify session info failed because SessionViews modify is busy."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code, message],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with SESSION_SERVICE_LOCK:
        message = "Modify session info failed."
        try:
            # 取出数据
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO
            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(SessionServiceInfoChecker,
                                                                request_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            request_data_dict["oper_type"] = UserManagerConstants.OPER_TYPE_MODIFY_SESSION_TIMEOUT
            request_data_dict["user_id"] = g.user_id
            # 获取 SessionService资源模板
            session_service_resource = json.loads(SessionServiceResourceSerializer().service.get_resource())
            ret_dict = SessionManager().patch_request(request_data_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                return ret_dict, session_service_resource

            if RedfishGlobals.get_user_locked_state(ret_dict):
                message = f"{message} {USER_LOCK_STATE}"
                RedfishGlobals.set_operational_log(request=request, username=RedfishGlobals.USER_NAME,
                                                   message=message)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": message,
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


SESSION_CREATE_LOCK = threading.Lock()


def head_add_token(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        res = func(*args, **kwargs)
        if res.status_code == CommonConstants.ERR_CODE_201:
            res.headers["X-Auth-Token"] = g.token
        return res
    return wrapper


@head_add_token
def rf_create_new_session():
    """
    功能描述：创建会话
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    @return:
    """
    message = "Login failed."
    print_operation_log = True
    tmp_user = None
    if SESSION_CREATE_LOCK.locked():
        message = "Create new session failed because SessionViews post is busy."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code, message],
        }
        return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

    with SESSION_CREATE_LOCK:
        run_log.info("Create new session called.")
        try:
            # 取出数据,并校验其合法性
            recv_conf_dict = request.get_data()
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                print_operation_log = False
                ret_dict = ret[1]
                return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_PARTICULAR_INFO)

            # 获取用户请求，并创建一个新的session
            recv_conf_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(CreateSessionServiceChecker, recv_conf_dict)
            if check_ret is not None:
                print_operation_log = False
                ret = {"status": 400, "message": [110207, "The user name or password error."]}
                return RedfishGlobals.return_error_info_message(ret, CommonConstants.ERR_GENERAL_INFO)

            run_log.info("Create new session begin.")
            recv_conf_dict["oper_type"] = UserManagerConstants.OPER_TYPE_GET_USER_TOKEN
            recv_conf_dict["real_ip"] = request.headers["X-Real-Ip"]

            # 判断用户名是否为有效用户名
            try:
                UserManager.find_user_by_username(recv_conf_dict.get("UserName"))
            except BizException:
                run_log.error("Get user by user name failed")
                print_operation_log = False
                ret_dict = {"status": 400, "message": [110207, "The user name or password error."]}
                return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            ret_dict = SessionManager().get_all_info(recv_conf_dict)
            # 判断返回的字典是否正确，如果正确，填充模板
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                # 向全局变量中设置UserName
                RedfishGlobals.USER_NAME = recv_conf_dict.get("UserName", None)
                current_app.user_name = RedfishGlobals.USER_NAME
                message = "Login successfully."
                sessions_members_resource = json.loads(SessionsMembersResourceSerializer().service.get_resource())
                result_value = ret_dict.get("message").get("result")
                tmp_id = result_value.get("Id", None)
                g.token = result_value.get("Token", None)
                RedfishGlobals.replace_id(sessions_members_resource, tmp_id)
                # 将成功创建会话的响应状态码设置为201
                ret_dict["status"] = CommonConstants.ERR_CODE_201
                return RedfishGlobals.update_json_of_list(sessions_members_resource, ret_dict)

            if RedfishGlobals.get_user_locked_state(ret_dict):
                message = f"{message}{USER_LOCK_STATE}"
            # 增加锁定后操作日志记录
            status = ret_dict.get("message")
            # 用户不存在错误码
            if status[0] == 110210:
                print_operation_log = False
                ret_dict = {"status": 400, "message": [110207, "The user name or password error."]}
            tmp_user = recv_conf_dict.get("UserName", None)
            if not RedfishGlobals.check_input_parm(tmp_user):
                # 进入此分支表示前端校验已被绕过，不记录非法输入
                tmp_user = "----"
            current_app.user_name = tmp_user
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": message,
            }
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
        finally:
            if print_operation_log:
                RedfishGlobals.set_operational_log(request, current_app.user_name, message)


SESSION_DELETE_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Logout")
def rf_session_id(member_id):
    """
    功能描述：删除指定会话
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    修改记录：2019-1-28 创建
    """
    message = "Logout failed."
    if SESSION_DELETE_LOCK.locked():
        message = "Logout failed because SessionViews delete is busy."
        run_log.error(message)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED.code, message],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with SESSION_DELETE_LOCK:
        try:
            check_ret = RedfishGlobals.check_external_parameter(DeleteServiceChecker, {"index": member_id})
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            # 调用数据库删除相关信息
            recv_conf_dict = {"dialog_id": member_id, "oper_type": UserManagerConstants.OPER_TYPE_DELETE_USER_TOKEN}
            ret_dict = SessionManager().get_all_info(recv_conf_dict)
            # 判断返回的字典是否正确，如果正确，填充模板
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                message = "Logout successfully."
                ret_dict = {"status": ret_dict.get("status"), "message": {"Message": "Logout Success"}}
                return ret_dict, success_message_resource
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("Delete session service error: %s", err)
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": message,
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
