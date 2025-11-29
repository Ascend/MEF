# !/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from functools import wraps
from typing import Dict
from typing import NoReturn

from flask import g, abort
from flask import make_response
from flask import request

from common.checkers import IPChecker

from lib_restful_adapter import LibRESTfulAdapter
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import UserManagerConstants
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.common_check import TokenCheck, get_error_resource
from user_manager.user_manager import UserManager
from user_manager.token_verification import TokenVerification


def get_privilege_auth():
    """
    功能描述： 鉴权
    参数：db数据库对象，Session表对象
    返回值：
    异常描述：NA
    修改记录：2019-3-5创建
    """

    class PrivilegeAuth:

        @staticmethod
        def token_required(func):
            @wraps(func)
            def wrapper(*args, **kwargs):
                # 1、从请求中获取参数并校验
                # 获取用户请求的token信息
                token = request.headers.get("X-Auth-Token")
                # 获取用户请求的URL
                request_url = request.url
                # 判断URL是否为创建会话，创建会话不需要鉴权
                if request.endpoint == "session_service.rf_create_new_session" and request.method == "POST":
                    return func(*args, **kwargs)
                # 获取用户请求的自动刷新信息，是否自动刷新，false：更新session时间。true：不更新session时间
                auto_refresh = request.headers.get("AutoRefresh", "true")
                try:
                    # 获取用户请求的ip地址
                    request_ip = request.headers["X-Real-Ip"]
                except Exception:
                    run_log.error("Get data from request header failed")
                    request_ip = None

                ret_dict = {}
                if not token or "null" == token:
                    ret_dict = {"status": 400, "message": "token is None"}

                if not request_ip or "null" == request_ip:
                    ret_dict = {"status": 400, "message": "request ip is None"}

                res = IPChecker("data").check({"data": request_ip})
                if not res.success:
                    ret_dict = {"status": 400, "message": "request ip invalid"}

                if ret_dict:
                    return PrivilegeAuth.return_message_resource_info(ret_dict, "GeneralError")

                # 2、封装请求数据，并调用token校验接口
                params = {"token": token, "request_ip": request_ip, "auto_refresh": auto_refresh}
                try:
                    ret_data = TokenVerification().get_all_info(params)
                except Exception:
                    run_log.error("Request TokenVerification error.")
                    ret_data = {}
                ret_data = ret_data or {}

                # 3、获取返回数据，处理业务，返回非200，表示token校验通过
                status = ret_data.get("status")
                message = ret_data.get("message")
                if CommonConstants.ERR_CODE_200 == status:
                    if not PrivilegeAuth.check_account_insecure_url(ret_data, request.path, request.method):
                        err_msg = "can not operate because of insecure account"
                        run_log.error(err_msg)
                        operate_log.info(
                            "[%s@%s] %s" % (message.get("user_name"), request.headers.get("X-Real-Ip"), err_msg))
                        ret_dict = {"status": 400, "message": err_msg}
                        return PrivilegeAuth.return_message_resource_info(ret_dict, "GeneralError")

                    message = ret_data.get("message")
                    g.user_id = message.get("user_id")
                    return func(*args, **kwargs)

                # 4、返回非200，表示token校验未通过，进行错误消息封装，然后返回
                return_value = TokenCheck.token_check(message, request_url)
                PrivilegeAuth.add_oper_logger(token)
                return PrivilegeAuth.user_make_response(return_value)

            return wrapper

        @staticmethod
        def return_message_resource_info(ret_dict, input_err_info):
            # 获取error模板资源
            resp_json, _ = get_error_resource(TokenCheck.index_file_path)
            resp_collection_json, _ = get_error_resource(
                TokenCheck.index_file_path_resource)

            return_value = TokenCheck.update_json_of_error_info(
                resp_json, resp_collection_json, ret_dict, input_err_info)
            return PrivilegeAuth.user_make_response(return_value)

        @staticmethod
        def add_oper_logger(token: str) -> NoReturn:
            user_info = PrivilegeAuth.get_user_info(token)
            user_name = user_info.get("user_name")
            if not user_name:
                run_log.error("Not found user name")
            else:
                operate_log.info("[%s@%s] %s" % (user_name, request.headers.get("X-Real-Ip"), "The session expires."))

        @staticmethod
        def get_user_info(token: str) -> Dict[str, str]:
            user_info = {}
            param_dict = {"oper_type": UserManagerConstants.USER_INFO_BY_TOKEN, "token": token}
            try:
                ret_data = UserManager().get_all_info(param_dict)
            except Exception:
                run_log.error("Get user info error.")
                return {}
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_data)
            if ret_status_is_ok:
                user_info = ret_data.get("message", {}).get("result", {}) or {}
            return user_info

        @staticmethod
        def user_make_response(data_list: list):
            if not data_list:
                abort(500)
            return make_response(data_list[1], data_list[0])

        @staticmethod
        def check_account_insecure_url(ret_data, request_uri, request_method):
            message = ret_data.get("message")
            if message and isinstance(message, dict):
                account_insecure = message.get("account_insecure_prompt")
                # 已经修改过默认密码了的token, 允许访问所有接口
                if not account_insecure:
                    return True
                # 默认密码的token，只允许访问修改账号密码的接口
                change_account_info_uri = "/redfish/v1/AccountService/Accounts/1"
                return request_method == "PATCH" and request_uri == change_account_info_uri
            return False

    return PrivilegeAuth
