#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import time
from typing import Optional

from common.constants import error_codes
from common.exception.biz_exception import Exceptions
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.exception_utils import ExceptionUtils
from user_manager.user_manager import UserManager, SessionManager, UserUtils, EdgeConfigManage


class TokenVerification(object):
    """
    功能描述：Token合法性校验
    接口：NA
    修改记录：2020-04-11 创建
    """
    def __init__(self):
        self.user_name: Optional[str] = None
        self.user_id: Optional[int] = None
        self.account_insecure_prompt: bool = True

    @staticmethod
    def judge_token(request_ip: str, token: str, auto_refresh: str):
        user_session = TokenVerification.check_and_get_session(auto_refresh, request_ip, token)
        # 数据库读取token过期时间
        edge_config = EdgeConfigManage.find_edge_config()
        token_timeout = edge_config.token_timeout

        user_one = UserManager.find_user_by_id(user_session.user_id)
        time_variance = time.perf_counter() - float(user_session.reset_time)
        if time_variance > int(token_timeout):
            # 会话超时后删除Session
            SessionManager.delete_session_by_user_id(user_one.id)
            run_log.warning("The session has expired and been deleted.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_TIME_LIMIT, time_variance)

        # 更新session, 如果传入true, 则不更新session, 如果传入false需要更新session
        if auto_refresh != "true":
            try:
                SessionManager.update_session(time.perf_counter())
            except Exception:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_DB_OPERATION_FAILED)

        account_insecure_prompt = user_one.account_insecure_prompt  # 账户安全提示
        password_modify_time = user_one.pword_modify_time
        password_valid_days = UserUtils.get_password_valid_days(account_insecure_prompt, password_modify_time)
        column_map = {}
        if password_valid_days != "--" and int(password_valid_days) <= 0:
            column_map["account_insecure_prompt"] = True
            UserManager.update_user_specify_column(user_one.id, column_map)
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_VALID_DAY,
                                           password_valid_days)

        return user_one.username_db, user_one.id, user_one.account_insecure_prompt

    @staticmethod
    def check_and_get_session(auto_refresh, request_ip, token):
        # 从header中获取到的auto_refresh字段只能是字符串类型的true或者false
        if auto_refresh not in ("true", "false"):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG, "auto_refresh")

        # 判断token是否匹配
        if not token:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
        user_session = SessionManager.find_session_by_user_id(1)
        if not user_session:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
        if not UserUtils.check_hash_password(user_session.token, token):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)

        # ip地址错误
        if request_ip != user_session.request_ip:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_REQUEST_IP_ADDR, request_ip)

        return user_session

    def get_all_info(self, param_dict: dict) -> dict:
        token = param_dict.get("token")
        auto_refresh = param_dict.get("auto_refresh")
        request_ip = param_dict.get("request_ip")
        try:
            # 判断token是否有效
            user_name, user_id, account_insecure_prompt = TokenVerification.judge_token(request_ip, token, auto_refresh)
        except Exception as ex:
            run_log.error("Token check failed %s", ex)
            return {"status": AppCommonMethod.ERROR, "message": ExceptionUtils.exception_process(ex)}

        self.user_name = user_name
        self.user_id = user_id
        self.account_insecure_prompt = account_insecure_prompt
        return {"status": AppCommonMethod.OK, "message": AppCommonMethod.get_json_info(self)}
