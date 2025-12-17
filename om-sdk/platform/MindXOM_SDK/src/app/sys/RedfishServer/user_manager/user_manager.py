# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import base64
import hashlib
import hmac
import json
import math
import os
import re
import threading
import time
from copy import deepcopy
from datetime import datetime
from typing import Dict, Any, List, AnyStr, NoReturn, Optional

from sqlalchemy import Float
from sqlalchemy import cast

from common.checkers import PasswordComplexityChecker
from common.checkers.param_checker import DeleteServiceChecker
from common.constants import error_codes
from common.constants.base_constants import CommonConstants, UserManagerConstants
from common.constants.error_codes import UserManageErrorCodes
from common.exception.biz_exception import Exceptions
from common.file_utils import FileUtils
from common.log.logger import run_log, operate_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.common_check import CommonCheck
from common.utils.date_utils import DateUtils
from common.utils.exception_utils import ExceptionUtils
from common.utils.number_utils import NumberUtils
from common.utils.system_utils import SystemUtils
from common.checkers import IPChecker, IntegerChecker
from lib_restful_adapter import LibRESTfulAdapter
from redfish_db.session import session_maker, simple_session_maker
from user_manager.hash_generator import HashGenerator
from user_manager.models import User, Session, HisPwd, EdgeConfig, LastLoginInfo


def generate_token(payload: dict, secret_key: bytes, salt: bytes) -> str:
    """
        通过hashlib.sha512算法和HMAC，生成token
        步骤1：通过hashlib.sha512算法得到salt和secret_key拼接后的hash值sha512_key（密钥）
        步骤2：通过HMAC得到hash对象mac
        步骤3：拼接返回结果
        :param payload: 作为生成HMAC hash对象入参（sha512_key + header + payload）一部分
        以及最后用于拼token
        :param secret_key: 密钥
        :param salt: 盐值 盐值+密钥 作为sha512算法入参，生成sha512key
        :return: token.
    """
    header = {'alg': 'SHA512', 'typ': 'JWT'}
    header_json = json.dumps(header)
    header_bs = base64.urlsafe_b64encode(header_json.encode("utf-8")).rstrip(b"=")

    payload_json = json.dumps(payload)
    payload_bs = base64.urlsafe_b64encode(payload_json.encode("utf-8")).rstrip(b"=")
    sha512_key = hashlib.sha512(salt + secret_key).digest()

    hm = hmac.new(sha512_key, header_bs + b'.' + payload_bs, digestmod='SHA512')
    hm_bs = base64.urlsafe_b64encode(hm.digest()).rstrip(b"=")

    return (header_bs + b'.' + payload_bs + b'.' + hm_bs).decode("utf-8")


class UserManager(object):
    USER_MANAGER_LOCK = threading.Lock()
    # 登录错误次数
    Login_Failed_times = 0

    def __init__(self):
        self.PasswordExpirationDays = None
        self.result = None

    @staticmethod
    def find_user_by_token(token: str) -> User:
        """
        根据token获取当前的用户信息
        """
        if not token:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, "token")
        user_session = SessionManager.find_session_by_token(token)
        user_info = UserManager.find_user_by_id(user_session.user_id)
        return user_info

    @staticmethod
    def find_users() -> List[AnyStr]:
        """
        根据用户id，获取当前的用户信息
        """
        with session_maker() as session:
            # 获取当前用户所有的历史密码记录
            session_username_list = session.query(User.username_db).order_by(User.id).all()
            username_list = []
            if session_username_list:
                for username in session_username_list:
                    username_list.append(username[0])
        return username_list

    @staticmethod
    def find_user_by_id(user_id: int) -> User:
        """
        根据用户id，获取当前的用户信息
        """
        with session_maker() as session:
            user_info = session.query(User).filter_by(id=user_id).first()
            if not user_info:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_FOUND)
            session.expunge(user_info)
        return user_info

    @staticmethod
    def find_user_list() -> List[User]:
        """
        获取当前数据库中所有的用户信息
        """
        user_list: List = []
        with session_maker() as session:
            user_info_list = session.query(User).all()
            if not user_info_list:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
            for one_user in user_info_list:
                user_list.append(deepcopy(one_user))
        return user_list

    @staticmethod
    def find_user_id_list() -> List[AnyStr]:
        """
        查询用户表中id的列表
        @return: 字符串类型的id集合
        """
        user_info_list = UserManager.find_user_list()
        id_list = []
        for one_user in user_info_list:
            id_list.append(str(one_user.id))
        return id_list

    @staticmethod
    def find_user_by_username(username: AnyStr, fd_modify_flag: bool = False) -> User:
        """
        根据用户名，获取当前的用户信息
        @param username: 用户名
        @param fd_modify_flag: fd修改标志位
        @return:
        """
        with session_maker() as session:
            user_info = session.query(User).filter_by(username_db=username).first()
            if not user_info:
                if fd_modify_flag:
                    # uds接口提示信息
                    raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
                else:
                    raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
            session.expunge(user_info)
        return user_info

    @staticmethod
    def update_user_specify_column(user_id: int, columns_map: Dict[AnyStr, Any]):
        if columns_map and user_id:
            with session_maker() as session:
                session.query(User).filter_by(id=user_id).update(columns_map)

    @staticmethod
    def get_account_service() -> int:
        """
        功能描述：查询会话服务信息
        实现sessiontimeout的查询
        参数：无
        返回值：sessiontimeout
        """
        edge_config = EdgeConfigManage.find_edge_config()
        default_expiration_days = edge_config.default_expiration_days
        return default_expiration_days

    @staticmethod
    def modify_account_service(expiration_day: int, user_id, password: AnyStr) -> int:
        """
         功能描述：修改用户服务信息
         参数：无
         返回值：
         异常描述：NA
        """

        run_log.info("Start to modify account service.")
        # 判断密码超时入参类型
        if expiration_day is None:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST,
                                           "expiration_day")
        if not isinstance(expiration_day, int):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_TYPE, "expiration_day")
        if not NumberUtils.in_range(expiration_day, UserManagerConstants.MIN_PASSWORD_EXPIRATION_DAY,
                                    UserManagerConstants.MAX_PASSWORD_EXPIRATION_DAY, "expiration_day"):
            run_log.error("Parameter expiration_day range error.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE, "expiration_day")

        # 校验密码
        if not user_id or not password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        if not UserManager.check_user_password(user_id, password):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        return UserManager.modify_expiration_days(expiration_day)

    @staticmethod
    def modify_expiration_days(expiration_day: int) -> int:
        expiration_day = NumberUtils.is_int_maxsize(expiration_day, "expiration_day")
        # 更新数据库
        EdgeConfigManage.update_expiration_days(expiration_day)
        option_value = EdgeConfigManage.find_edge_config().default_expiration_days
        if not isinstance(option_value, int):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_FORMAT, "PasswordExpirationDays")
        # check 修改是否成功
        return option_value

    # 根据用户id查询
    @staticmethod
    def get_user_information(user_id: int):
        """
        功能描述：查询指定用户信息
        参数：无
        返回值：
        异常描述：NA
        """
        user_id = NumberUtils.is_int_maxsize(user_id, "user_id")
        user_one = UserManager.find_user_by_id(NumberUtils.is_int_maxsize(user_id, "user_id"))
        password_valid_days = UserUtils.get_password_valid_days(
            user_one.account_insecure_prompt, user_one.pword_modify_time)

        # 新增位置信息反馈
        wrong_times = UserManager.Login_Failed_times
        UserManager.Login_Failed_times = 0
        last_login_ip = SessionManager.get_login_ip_from_db()
        # 组拚返回的数据字典
        return {'Id': str(user_one.id),
                'LastLoginSuccessTime': user_one.last_login_success_time,
                'LastLoginFailureTime': user_one.last_login_failure_time,
                'AccountInsecurePrompt': user_one.account_insecure_prompt,
                'ConfigNavigatorPrompt': user_one.config_navigator_prompt,
                'PwordWrongTimes': wrong_times,
                'PasswordValidDays': password_valid_days,
                'LastLoginIP': last_login_ip}

    @staticmethod
    def modify_username_password(user_id: int, new_username: AnyStr,
                                 old_password: AnyStr, new_password: AnyStr,
                                 new_password_second: AnyStr):
        """
        修改用户名与密码
        @param user_id:用户的id
        @param new_username:用户输入的用户名
        @param old_password: 用户输入的原密码
        @param new_password:用户输入的新密码
        @param new_password_second: 用户输入的二次确认密码
        """
        # 修改密码和用户名 锁定状态和使能
        run_log.info("Modify user info start")

        # 对参数进行校验
        user_id = UserManager.check_modify_user_password(new_username, old_password, user_id)
        user_info = UserManager.find_user_by_id(user_id)

        if (new_password or new_password_second) and new_password != new_password_second:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_TWO_INCONSISTENT)

        # 判断根目录对应分区的空间是否被占满，如果已经占满，则抛出异常信息
        if SystemUtils.get_available_size("/") == 0:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_DIR_SPACE_LOW, "/")

        if new_password and new_password_second and user_info.username_db == new_username:
            # 仅修改密码
            run_log.info("Modify user password info start")
            UserUtils.check_password(new_password, "Password")
            UserUtils.judge_password(user_id, new_username,
                                     user_info.username_db, new_password,
                                     new_password_second)
            UserUtils.modify_password(user_id, new_password, False)
        else:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)

    @staticmethod
    def check_modify_user_password(new_username, old_password, user_id):
        """
        修改前对参数进行校验
        @param new_username:
        @param old_password:
        @param user_id:
        @return:
        """
        # 校验user_id、new_username是否存在
        user_id = NumberUtils.is_int_maxsize(user_id, "user_id")
        if not new_username:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        if not old_password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        # 校验用户名、密码的合法性
        UserUtils.check_username(new_username)
        # 判断当前用户否锁定，如果已经锁定，则抛出异常，如果锁定时间到则解锁
        UserUtils.unlock_user_locked(user_id)
        # 判断原密码是否正确
        UserUtils.verify_pword_and_locked(user_id, old_password)
        return user_id

    @staticmethod
    def check_user_password(user_id: int, password: AnyStr) -> bool:
        """
        检验密码是否合法
        :param user_id:用户id
        :param password: 用户的密码
        :return:
        """
        user_id = NumberUtils.is_int_maxsize(user_id, "user_id")
        # 密码为空提示用户名密码错误
        if not password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)

        # 判断当前用户否锁定，如果已经锁定，则抛出异常，如果锁定时间到则解锁
        UserUtils.unlock_user_locked(user_id)
        # 判断密码是否正确
        UserUtils.verify_pword_and_locked(user_id, password)
        return True

    def fd_modify_passwd(self, param_dict: Dict[AnyStr, Any]):
        # FD修改密码
        default_list = ["Password", "UserName"]
        if not CommonCheck.check_sub_list(default_list, list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)

        username = param_dict.get("UserName", None)
        password = param_dict.get("Password")
        UserUtils.check_username(username, True)
        UserUtils.check_password(password, "Password")
        user_one = UserManager.find_user_by_username(username, True)
        user_id = user_one.id

        UserUtils.judge_password(user_id, username, username, password, password)
        hash_pword = UserUtils.hash_pword(password)
        modify_time = datetime.utcnow().strftime(CommonConstants.STR_DATE_FORMAT)
        user_value = {
            'pword_hash': hash_pword,
            'account_insecure_prompt': True,
            'pword_modify_time': modify_time,
        }
        with session_maker() as session:
            # 修改os用户密码
            request_data_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_RESET_OS_PASSWORD,
                "Username": username, "Password": password
            }
            ret_dict = LibRESTfulAdapter.lib_restful_interface("UserOperate", "POST", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            # 判断脚本是否正确执行
            if not ret_status_is_ok:
                run_log.error("modify os user info failed.")
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_OS_MODIFY)

            # 修改密码
            session.query(User).filter_by(id=user_id).update(user_value)
            # 向历史表中添加数据
            HisPwdManage.save_his_pwd(user_id, hash_pword, modify_time, session)
            # FD修改完成后删除会话
            session.query(Session).filter_by(user_id=user_id).delete()
            run_log.warning("The session has been deleted.")

        run_log.info("modify user info successfully.")
        # 返回信息
        self.result = {'AccountInsecurePrompt': True}
        return user_id

    def get_all_info(self, param_dict: Dict[str, Any]) -> Dict[str, Any]:
        try:
            oper_type = param_dict.get("oper_type")
            del param_dict['oper_type']
            if UserManagerConstants.USER_INFO_BY_TOKEN == oper_type:
                token = param_dict.get("token")
                user_info = UserManager.find_user_by_token(token)
                if user_info:
                    self.result = {'user_id': user_info.id, 'user_name': user_info.username_db}
            elif UserManagerConstants.OPER_TYPE_GET_ACCOUNT_EXPIRATION_DAY == oper_type:
                # 获取密码有效期
                expiration_day = UserManager.get_account_service()
                self.PasswordExpirationDays = expiration_day
            elif UserManagerConstants.OPER_TYPE_GET_USER_ID_LIST == oper_type:
                # 获取用户id列表
                self.result = UserManager.find_user_id_list()
            elif UserManagerConstants.OPER_TYPE_GET_USER_INFO == oper_type:
                # 获取用户信息
                user_id: int = param_dict.get("user_id")
                user_id = NumberUtils.is_int_maxsize(user_id, "user_id")
                self.result = UserManager.get_user_information(user_id)
            elif UserManagerConstants.OPER_TYPE_GET_USER_LIST == oper_type:
                # 获取当前的用户名列表
                self.result = UserManager.find_users()
            elif UserManagerConstants.OPER_TYPE_CHECK_PASSWORD == oper_type:
                # 密码校验
                if not CommonCheck.check_sub_list(["password", "user_id"], list(param_dict.keys())):
                    raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
                user_id: int = param_dict.get("user_id")
                password: AnyStr = param_dict.get("password")
                ret: bool = UserManager.check_user_password(user_id, password)
                if not ret:
                    raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
                # 获取当前的用户名列表
                self.result = ret
        except Exception as ex:
            run_log.error("Token check failed.")
            return {"status": AppCommonMethod.ERROR, "message": ExceptionUtils.exception_process(ex)}
        return {"status": AppCommonMethod.OK, "message": AppCommonMethod.get_json_info(self)}

    def modify_password_expiration_day_from_fd(self, param_dict):
        default_list = ["PasswordExpirationDays"]
        if not CommonCheck.check_sub_list(default_list, list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        expiration_day = param_dict.get("PasswordExpirationDays", None)
        if not NumberUtils.in_range(
                expiration_day, UserManagerConstants.MIN_PASSWORD_EXPIRATION_DAY,
                UserManagerConstants.MAX_PASSWORD_EXPIRATION_DAY, "expiration_day"):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE, "expiration_day")
        UserManager.modify_expiration_days(expiration_day)

    def modify_user_name_password(self, param_dict):
        default_list = ["old_password", "Password", "new_password_second", "UserName", "user_id"]
        if not CommonCheck.check_sub_list(default_list, list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        user_id = param_dict.get("user_id", None)
        new_username = param_dict.get("UserName", None)
        old_password = param_dict.get("old_password")
        new_password = param_dict.get("Password")
        new_password_second = param_dict.get("new_password_second")
        UserManager.modify_username_password(
            user_id, new_username, old_password, new_password, new_password_second)
        # 返回信息
        self.result = self.get_user_information(user_id)
        # web修改完成后删除会话
        SessionManager.delete_session_by_user_id(user_id)
        run_log.warning("The session has been deleted.")

    def modify_password_expiration_day(self, param_dict):
        if not CommonCheck.check_sub_list(
                ["PasswordExpirationDays", "Password", "user_id"], list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        expiration_day = param_dict.get("PasswordExpirationDays", None)
        user_id = param_dict.get("user_id", None)
        password = param_dict.get("Password")
        ret_expiration_day = UserManager.modify_account_service(expiration_day, user_id, password)
        self.PasswordExpirationDays = ret_expiration_day

    def patch_request(self, param_dict: Dict[AnyStr, Any]) -> Any:
        try:
            if UserManager.USER_MANAGER_LOCK.locked():
                run_log.warning("UserManager modify is busy")
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED)
            with UserManager.USER_MANAGER_LOCK:
                oper_type = param_dict.get("oper_type")
                if not oper_type:
                    raise Exceptions.biz_exception(
                        error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, "oper_type")
                del param_dict['oper_type']
                if UserManagerConstants.OPER_TYPE_MODIFY_ACCOUNT_EXPIRATION_DAY == oper_type:
                    # 修改密码有效期
                    self.modify_password_expiration_day(param_dict)
                if UserManagerConstants.OPER_TYPE_MODIFY_USER_INFO == oper_type:
                    # 修改用户名与密码
                    self.modify_user_name_password(param_dict)
                if UserManagerConstants.OPER_TYPE_MODIFY_PASSWORD == oper_type:
                    self.fd_modify_passwd(param_dict)
                if UserManagerConstants.OPER_TYPE_FD_MODIFY_ACCOUNT_EXPIRATION_DAY == oper_type:
                    # FD修改密码有效期天数
                    self.modify_password_expiration_day_from_fd(param_dict)
        except Exception as ex:
            run_log.error("UserManager patch failed.")
            return {"status": AppCommonMethod.ERROR, "message": ExceptionUtils.exception_process(ex)}
        return {"status": AppCommonMethod.OK, "message": AppCommonMethod.get_json_info(self)}


class SessionManager(object):
    USER_IP = None
    libConfigMutex = threading.Lock()

    def __init__(self):
        self.SessionTimeout = None
        self.result = None

    @staticmethod
    def get_session_service() -> int:
        """
        功能描述：查询会话服务信息
        实现sessiontimeout的查询
        参数：无
        返回值：sessiontimeout
        """
        edge_config = EdgeConfigManage.find_edge_config()
        session_timeout = edge_config.token_timeout
        # 秒 转为分钟
        return int(session_timeout / 60)

    @staticmethod
    def delete_dialog_id(dialog_id: AnyStr) -> NoReturn:
        """
        功能描述：根据index删除指定会话
        参数：index
        返回值：
        异常描述：NA
        """
        # 校验用户表中是否存在该用户

        session_one = SessionManager.find_session_by_dialog_id(dialog_id)
        if session_one:
            SessionManager.delete_session_by_user_id(session_one.user_id)
            run_log.warning("The session has been deleted.")

    @staticmethod
    def find_session_by_token(token: str) -> Session:
        """
        根据token获取当前的session信息
        """
        with session_maker() as session:
            user_session = session.query(Session).filter_by(user_id=1).first()
            if not user_session:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
            if not UserUtils.check_hash_password(user_session.token, token):
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
            session.expunge(user_session)
        return user_session

    @staticmethod
    def find_session_by_dialog_id(dialog_id: str) -> Session:
        """
        根据token获取当前的session信息
        """
        if not dialog_id:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
        if not NumberUtils.in_range(len(dialog_id), UserManagerConstants.MIN_SESSION_LENGTH,
                                    UserManagerConstants.MAX_SESSION_LENGTH, "dialog_id"):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE, "dialog_id")
        with session_maker() as session:
            user_session = session.query(Session).filter_by(dialog_id=dialog_id).first()
            if not user_session:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
            session.expunge(user_session)
        return user_session

    @staticmethod
    def find_session_by_user_id(user_id: int) -> Session:
        """
        根据用户id 获取当前的的session
        """
        with session_maker() as session:
            user_session = session.query(Session).filter_by(user_id=user_id).first()
            if not user_session:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND)
            session.expunge(user_session)
        return user_session

    @staticmethod
    def delete_timeout_session() -> int:
        """
        删除超时Session
        """
        timeout = time.perf_counter() - EdgeConfigManage.find_edge_config().token_timeout
        with session_maker() as session:
            return session.query(Session).filter(
                cast(Session.reset_time, Float) < timeout
            ).delete(synchronize_session=False)

    @staticmethod
    def update_session(reset_time) -> int:
        with session_maker() as session:
            ret = session.query(Session).filter_by(user_id=1).update({"reset_time": reset_time})
        return ret

    @staticmethod
    def delete_session_by_user_id(user_id: int) -> int:
        """
        根据session的user_id删除当前用户下的session信息
        """
        with session_maker() as session:
            ret = session.query(Session).filter_by(user_id=user_id).delete()
        return ret

    @staticmethod
    def save_session(session_info: Session) -> NoReturn:
        """
        根据session的id信息删除当前的session
        """
        with session_maker() as session:
            session.add(session_info)

    @staticmethod
    def check_default_passwd(user_one):
        """
        检验传入用户名是否正确
        @param user_one: 用户名
        """
        # 1、校验用户名是否为字符串
        # 5、初始密码禁止登陆
        if user_one.account_insecure_prompt:
            raise Exceptions.biz_exception(
                error_codes.UserManageErrorCodes.ERROR_PASSWORD_IS_INITIAL,
                str(user_one.id) + ":" + user_one.username_db)

    @staticmethod
    def create_redfish_session(username: AnyStr, password: AnyStr, real_ip: AnyStr) -> AnyStr:
        """
         功能描述： 创建会话
         参数：request_data　request_ip
         返回值：
         异常描述：NA
        """
        column_map = {}
        user_id: Optional[int] = None

        try:
            # 1、检查参数
            if not username:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
            try:
                UserUtils.check_username(username)
            except Exception as ex:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD) from ex
            # 2、根据用户名获取当前用户, 用户不存在，主动抛"110210, User not found."异常
            try:
                user_one: User = UserManager.find_user_by_username(username)
            except Exception as ex:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD) from ex
            user_id = user_one.id
            # 3、校验用户是否为锁定状态
            UserUtils.unlock_user_locked(user_id)

            res = IPChecker("data").check({"data": real_ip})
            if not res.success:
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG, "ip")

            # 4、校验用户的密码是否与输入密码匹配
            UserUtils.verify_pword_and_locked(user_id, password)
            # 5、创建会话前删除原来的session会话
            if SessionManager.delete_session_by_user_id(user_one.id):
                run_log.warning("The session has been deleted.")
                message = f"The user[{SessionManager.get_login_ip_from_db()}] is forced to logout."
                operate_log.info("[%s@%s] %s" % (user_one.username_db, SessionManager.USER_IP, message))
            # 6、保存最新的会话信息
            current_time = DateUtils.get_format_time(time.time())
            # 会话超时判断改用系统运行时差判断
            session_info = Session()
            session_info.dialog_id = UserUtils.get_security_random().hex()
            token = UserUtils.get_token()
            session_info.token = UserUtils.hash_pword(token)
            session_info.user_id = user_one.id
            session_info.request_ip = real_ip
            session_info.visit_times = 0
            session_info.create_time = current_time
            session_info.reset_time = time.perf_counter()
            SessionManager.save_session(session_info)
            # 记录错误次数
            UserManager.Login_Failed_times = user_one.pword_wrong_times
            # 7、将用户成功创建会话的时间记录到user表中
            if user_one.last_login_success_time or user_one.log_in_time:
                column_map['last_login_success_time'] = user_one.log_in_time
            column_map['log_in_time'] = current_time
            # 将用户登录错误次数清零
            column_map['pword_wrong_times'] = 0
            column_map['start_lock_time'] = 0
            return token
        finally:
            # 每次登录尝试都更新登录IP信息
            SessionManager.set_login_ip(real_ip)
            if column_map and user_id:
                UserManager.update_user_specify_column(user_id, column_map)

    @staticmethod
    def set_login_ip(real_ip):
        try:
            with session_maker() as session:
                # 至多一条数据，先删后建
                session.query(LastLoginInfo).delete()
                session.add(LastLoginInfo(ip=real_ip))
        except Exception as err:
            run_log.error("set login ip to db failed, catch %s", err.__class__.__name__)

    @staticmethod
    def get_login_ip_from_db() -> Optional[str]:
        login_ip = None
        try:
            with session_maker() as session:
                info: LastLoginInfo = session.query(LastLoginInfo).first()
                login_ip = info.ip if info else None
        except Exception as err:
            run_log.error("get login ip from db failed, catch %s", err.__class__.__name__)
        return login_ip

    @staticmethod
    def get_login_ip():
        return SessionManager.USER_IP or SessionManager.get_login_ip_from_db()

    def get_all_info(self, param_dict: Dict[AnyStr, Any]) -> Dict[AnyStr, Any]:
        try:
            oper_type = param_dict.get("oper_type")
            del param_dict['oper_type']
            if UserManagerConstants.OPER_TYPE_GET_USER_TOKEN == oper_type:
                if SessionManager.libConfigMutex.locked():
                    run_log.warning("SessionManager modify is busy")
                    raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED)
                with SessionManager.libConfigMutex:
                    self.create_session(param_dict)
            # 删除session
            if UserManagerConstants.OPER_TYPE_DELETE_USER_TOKEN == oper_type:
                if SessionManager.libConfigMutex.locked():
                    run_log.warning("SessionManager modify is busy")
                    raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED)
                with SessionManager.libConfigMutex:
                    self.del_session(param_dict)
            if UserManagerConstants.OPER_TYPE_GET_SESSION_TIMEOUT == oper_type:
                # 获取session超时时间
                session_timeout = SessionManager.get_session_service()
                self.SessionTimeout = session_timeout
        except Exception as ex:
            run_log.error("Operation failed, %s", ex)
            return {"status": AppCommonMethod.ERROR, "message": ExceptionUtils.exception_process(ex)}
        return {"status": AppCommonMethod.OK, "message": AppCommonMethod.get_json_info(self)}

    def create_session(self, param_dict):
        # 校验参数是否为预期的参数
        def_param_list = ["UserName", "Password", "real_ip"]
        if not CommonCheck.check_sub_list(def_param_list, list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        # 获取请求的参数列表
        username = param_dict.get("UserName", None)
        password = param_dict.get("Password")
        real_ip = param_dict.get("real_ip", None)

        res = IPChecker("data").check({"data": real_ip})
        if not res.success:
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG, "ip")

        # 校验是否符合登录规则
        request_data_dict = {"oper_type": UserManagerConstants.OPER_TYPE_CHECK_SECURITY_CONFIG, "real_ip": real_ip}
        ret_dict = LibRESTfulAdapter.lib_restful_interface("UserOperate", "POST", request_data_dict, False)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        # 判断脚本是否正确执行
        if not ret_status_is_ok:
            # 内部通讯失败，返回内部错误 send_msg_err_list为send_msg的错误返回值
            send_msg_err_list = ["Send message failed.", "Connect server failed."]
            error_message = ret_dict.get("message")
            if error_message == "Socket path is not exist.":
                run_log.error("Send msg failed, reason: %s.", error_message)
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_SERVICE_IS_STARTING)
            if error_message in send_msg_err_list or "client_kmc.psd invalid" in error_message:
                run_log.error("Send msg failed, reason: %s.", error_message)
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_SERVICE_STARTUP_FAILED)
            raise Exceptions.biz_exception(UserManageErrorCodes.ERROR_SECURITY_CFG_NOT_MEET)
        # 创建session并返回创建的token
        SessionManager.USER_IP = real_ip
        user_token = SessionManager.create_redfish_session(username, password, real_ip)
        # 查询创建的session
        one_user = UserManager.find_user_by_username(username)
        one_session = SessionManager.find_session_by_user_id(one_user.id)
        if one_session:
            self.result = {
                'Id': one_session.dialog_id,
                'UserName': username,
                'Token': user_token,
                'UserId': one_user.id,
                'AccountInsecurePrompt': one_user.account_insecure_prompt,
                'message': "[" + str(one_user.id) + ":" + one_user.username_db + "]",
            }

    def del_session(self, param_dict):
        if not CommonCheck.check_sub_list(["dialog_id"], list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        dialog_id = param_dict.get("dialog_id", None)
        if not DeleteServiceChecker().check({"index": dialog_id}):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_WRONG, "dialog_id")

        SessionManager.delete_dialog_id(dialog_id)

    def modify_session_service(self, session_timeout: int, user_id, password: AnyStr):
        """
        :param session_timeout: 会话超时时间
        :param user_id: 用户id
        :param password: 用户密码
        :return:
        """
        run_log.info("Start to modify session service.")
        # 判断会话超时入参类型
        if session_timeout is None:
            raise Exceptions.biz_exception(
                error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, "session_timeout")
        if not isinstance(session_timeout, int):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_TYPE, "session_timeout")
        if not NumberUtils.in_range(session_timeout, UserManagerConstants.MIN_SESSION_TIMEOUT,
                                    UserManagerConstants.MAX_SESSION_TIMEOUT, "SessionTimeout"):
            run_log.error("Parameter session_timeout range error.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE, "SessionTimeout")
        # 校验密码
        if not user_id or not password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        if not UserManager.check_user_password(user_id, password):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)

        self.update_session_timeout(session_timeout)

    def modify_session_timeout(self, param_dict):
        # 修改session超时时间
        if not CommonCheck.check_sub_list(["SessionTimeout", "Password", "user_id"],
                                          list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        session_timeout = param_dict.get("SessionTimeout", None)
        user_id = param_dict.get("user_id", None)
        password = param_dict.get("Password")
        self.modify_session_service(session_timeout, user_id, password)

    def update_session_timeout(self, session_timeout):
        """
        :param session_timeout: 会话超时时间
        :return:
        """
        # 更新数据库
        EdgeConfigManage.update_token_timeouts(session_timeout * 60)
        # check配置
        option_value = EdgeConfigManage.find_edge_config().token_timeout
        if not isinstance(option_value, int):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_FORMAT, "token_timeout")
        self.SessionTimeout = math.ceil(option_value / 60)

    def fd_modify_session_timeout(self, param_dict):
        """
        :param param_dict: 字典，key:SessionTimeout
        :return:
        """
        # 修改session超时时间
        if not CommonCheck.check_sub_list(["SessionTimeout"], list(param_dict.keys())):
            raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_NUMBER_WRONG)
        session_timeout = param_dict.get("SessionTimeout", None)
        ret = IntegerChecker(
            "SessionTimeout",
            min_value=UserManagerConstants.MIN_SESSION_TIMEOUT,
            max_value=UserManagerConstants.MAX_SESSION_TIMEOUT).check({"SessionTimeout": session_timeout})
        if not ret.success:
            run_log.error("Parameter session_timeout range error.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PARAM_RANGE, "SessionTimeout")
        # 防止更新数据库异常，要强转一下 session_timeout
        self.update_session_timeout(int(session_timeout))

    def patch_request(self, param_dict: Dict[AnyStr, Any]) -> Any:
        try:
            if SessionManager.libConfigMutex.locked():
                run_log.warning("SessionManager modify is busy")
                raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ROOT_OPERATE_LOCKED)
            with SessionManager.libConfigMutex:
                oper_type = param_dict.get("oper_type")
                if not oper_type:
                    raise Exceptions.biz_exception(
                        error_codes.CommonErrorCodes.ERROR_ARGUMENT_VALUE_NOT_EXIST, "oper_type")
                del param_dict['oper_type']
                if UserManagerConstants.OPER_TYPE_MODIFY_SESSION_TIMEOUT == oper_type:
                    self.modify_session_timeout(param_dict)
                elif UserManagerConstants.OPER_TYPE_FD_MODIFY_SESSION_TIMEOUT == oper_type:
                    self.fd_modify_session_timeout(param_dict)
                else:
                    raise Exceptions.biz_exception(error_codes.CommonErrorCodes.ERROR_ARGUMENT_FORMAT, "oper_type")
        except Exception as ex:
            run_log.error("Session modify failed.")
            return {"status": AppCommonMethod.ERROR, "message": ExceptionUtils.exception_process(ex)}
        return {"status": AppCommonMethod.OK, "message": AppCommonMethod.get_json_info(self)}


class HisPwdManage(UserManager):
    @staticmethod
    def judge_password_duplicate(user_id, new_password):
        with session_maker() as session:
            pwd_hash_list = session.query(HisPwd.history_pword_hash).filter_by(
                user_id=user_id).order_by(HisPwd.id.desc()).limit(5).all()
            if not pwd_hash_list:
                run_log.info("history password not found!")
                return
            password_hash_list = deepcopy(pwd_hash_list)

        for one_hash in password_hash_list:
            if UserUtils.check_hash_password(one_hash[0], new_password):
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_DUPLICATE)

    @staticmethod
    def save_his_pwd(user_id: int, pword_hash: AnyStr, modify_time: AnyStr, session=None):
        with simple_session_maker(session) as session:
            # 获取当前用户所有的历史密码记录
            pwd_hash_list = session.query(HisPwd.id).filter_by(user_id=user_id).order_by(HisPwd.id.desc()).all()
            if pwd_hash_list:
                delete_id_list_info = [his_pwd_id[0] for his_pwd_id in pwd_hash_list[4:]]
                # 只保留4条数据
                if delete_id_list_info:
                    session.query(HisPwd).filter(HisPwd.id.in_(delete_id_list_info)).delete(synchronize_session=False)
            # 插入历史表中
            his_pwd_info = HisPwd()
            his_pwd_info.user_id = user_id
            his_pwd_info.history_pword_hash = pword_hash
            his_pwd_info.pword_modify_time = modify_time
            session.add(his_pwd_info)

    @staticmethod
    def judge_is_default_password(new_password):
        file_path = os.path.join("{}/user_manager/config/".format(AppCommonMethod.get_project_absolute_path()),
                                 "paw.ini")
        file_context = FileUtils.read_file(file_path, "r")
        if UserUtils.check_hash_password(file_context[0], new_password):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_IS_DEFAULT)


class UserUtils:
    @staticmethod
    def check_username_pattern(username: str, min_length: int = 4, max_length: int = 30) -> bool:
        """
        用户名规则校验。1、必须是数字、大小写字母。2、长度是1-16位。3、用户名不能是全数字
        :param username: 用户名
        :param min_length: 用户名最小长度
        :param max_length: 用户名最大长度
        :return: 是否检验通过
        """
        is_ok = False
        max_length = max_length
        if username:
            pattern = r'^[a-zA-Z0-9]{' + str(min_length) + ',' + str(max_length) + '}$'
            if re.fullmatch(pattern, username):
                is_ok = True
            if username.isdigit():
                is_ok = False
        return is_ok

    @staticmethod
    def get_security_random():
        with open("/dev/random", 'rb') as file:
            system_random = file.read(24)
        return system_random

    @staticmethod
    def get_token():
        # 获取随机secret key
        secret_key = UserUtils.get_security_random()
        # 获取随机的盐值
        salt_read = UserUtils.get_security_random()
        random_id = UserUtils.get_security_random()
        payload = {'user_id': base64.encodebytes(random_id).decode('utf-8')}
        try:
            return generate_token(payload, secret_key, salt_read)
        except Exception as e:
            run_log.error(f'{e}')
            raise Exception("get_token failed") from e

    @staticmethod
    def check_username(username: AnyStr, fd_modify_flag: bool = False):
        """
        检验传入用户名是否正确
        @param username: 用户名
        @param fd_modify_flag: fd修改标志位
        """
        # 1、校验用户名是否为字符串
        if not isinstance(username, (str, bytes)):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        # 2、用户名的规则是否为：长度为1~16个字符，可由数字和字母组成，不能为全数字
        if not UserUtils.check_username_pattern(
                username, UserManagerConstants.MIN_USERNAME_PASSWORD_LENGTH,
                UserManagerConstants.MAX_USERNAME_PASSWORD_LENGTH):
            if fd_modify_flag:
                # uds接口提示信息
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
            else:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)

    @staticmethod
    def check_password_valid(password: AnyStr):
        """
        检验传入用户密码是否正确
        @param password: 用户密码
        """
        if isinstance(password, str) and (len(password) > 7) and (len(password) < 21):
            return True
        return False

    @staticmethod
    def verify_pword_and_locked(user_id: int, password: AnyStr) -> NoReturn:
        """
        判断用户输入的原密码是否正确，如果正确则返回，如果不正确，将错误次数加1，超过阀值，则锁定用户
        @param user_id: 用户id
        @param password: 用户的密码
        @return:
        """
        column_map = {}
        one_user_info = UserManager.find_user_by_id(user_id)
        try:
            if UserUtils.check_password_valid(password) and \
                    UserUtils.check_hash_password(one_user_info.pword_hash, password):
                # 将用户登录错误次数清零
                column_map['pword_wrong_times'] = 0
                column_map['start_lock_time'] = 0
                return
            # 抛出异常
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        except Exception as e:
            try:
                UserUtils.modify_wrong_times(user_id)
            except Exception as ex:
                raise ex
            raise e
        finally:
            if column_map:
                UserManager.update_user_specify_column(user_id, column_map)

    @staticmethod
    def modify_wrong_times(user_id: int):
        column_map = {}
        one_user_info = UserManager.find_user_by_id(user_id)
        error_times = one_user_info.pword_wrong_times + 1
        column_map['pword_wrong_times'] = error_times  # 密码错误的次数
        column_map['last_login_failure_time'] = DateUtils.get_format_time(time.time())
        # 连续登录失败次数门限
        # 用户重试次数过多，抛出异常
        edge_config = EdgeConfigManage.find_edge_config()
        if error_times >= edge_config.default_lock_times:
            column_map['lock_state'] = True
            column_map['start_lock_time'] = int(time.perf_counter())  # 开始锁定的时间
            UserManager.update_user_specify_column(user_id, column_map)
            # 锁定之后删除会话
            SessionManager.delete_session_by_user_id(user_id)
            run_log.warning("The session has been deleted.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_LOCK_STATE,
                                           str(edge_config.lock_duration))
        UserManager.update_user_specify_column(user_id, column_map)

    @staticmethod
    def unlock_user_locked(user_id: int) -> NoReturn:
        """
        校验当前用户是否锁定，如果已过了锁定时间，进行解锁
        @param user_id: 用户的id
        """
        user_one = UserManager.find_user_by_id(user_id)
        if not user_one.lock_state:
            return

        edge_config = EdgeConfigManage.find_edge_config()
        # 判断系统是否重启过, 是的话清零锁定时间
        if int(time.perf_counter()) < user_one.start_lock_time:
            column_map = {'start_lock_time': 0}
            UserManager.update_user_specify_column(user_id, column_map)
        # 校验用户是否到达解锁状态
        if int(time.perf_counter()) - user_one.start_lock_time <= edge_config.lock_duration:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_LOCK_STATE,
                                           str(edge_config.lock_duration - int(time.perf_counter()) +
                                               user_one.start_lock_time))
        else:
            # 解锁用户的锁定状态
            column_map = {'lock_state': False, 'start_lock_time': 0, 'pword_wrong_times': 0}
            UserManager.update_user_specify_column(user_id, column_map)
            operate_log.info("[%s@%s] %s" % (user_one.username_db, "LOCAL", "The user is unlocked."))

    @staticmethod
    def check_password(password: AnyStr, prompt: AnyStr) -> NoReturn:
        """
        检验密码是否符合要求，同时用户名不能与密码相同
        :param password: 提示内容
        :param prompt: 密码字符串
        :return:
        """
        if not password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        if not isinstance(password, (str, bytes)):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        # 密码复杂度，密码要求包含至少三种字符
        check_ret = PasswordComplexityChecker(attr_name="Password", min_len=8, max_len=20).check({"Password": password})
        if not check_ret.success:
            run_log.error("checker pword complexity failed.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)

    @staticmethod
    def modify_password(user_id, new_password, account_prompt: bool = False):
        hash_pword = UserUtils.hash_pword(new_password)
        column_map = {'pword_hash': hash_pword}

        if account_prompt:
            column_map['account_insecure_prompt'] = True  # 默认密码时设置为True
            run_log.info("Set account_insecure_prompt to True.")
        else:
            column_map['account_insecure_prompt'] = False

        modify_time = datetime.utcnow().strftime(CommonConstants.STR_DATE_FORMAT)
        column_map['pword_modify_time'] = modify_time
        UserManager.update_user_specify_column(user_id, column_map)

        # 向历史表中添加数据
        HisPwdManage.save_his_pwd(user_id, hash_pword, modify_time)

    @staticmethod
    def hash_pword(password):
        """密码加密"""
        hash_pword = HashGenerator.generate_password_hash(password)
        return hash_pword

    @staticmethod
    def check_hash_password(pword_hash, password):
        """密码校验"""
        if not pword_hash or not password:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        return HashGenerator.check_password_hash(pword_hash, password)

    @staticmethod
    def judge_password(user_id: int, new_username: AnyStr,
                       old_username: AnyStr, new_password: AnyStr,
                       new_password_second: AnyStr) -> NoReturn:
        """
        判断新密码是否与原用户名与新用户名相同
        @param user_id: 用户的id
        @param new_username: 新用户名
        @param old_username: 老的用户名
        @param new_password: 新的密码
        @param new_password_second: 新密码的二次确认
        """
        # 口令不能和(新)账号名或(新)账号名的倒写一致
        if new_password != new_password_second:
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_USER_NOT_MATCH_PASSWORD)
        if new_password in (new_username, new_username[::-1]) or new_password in (old_username, old_username[::-1]):
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_COMPARED_USERNAME_REVERSAL)

        # 不能设置为默认密码
        HisPwdManage.judge_is_default_password(new_password)
        # 新密码不能与前5次一样；根据写入规则：首条必定为默认密码
        HisPwdManage.judge_password_duplicate(user_id, new_password)
        # 弱字典校验
        request_data_dict = {"oper_type": UserManagerConstants.OPER_TYPE_CHECK_WEAK_DICT, "Password": new_password}
        ret_dict = LibRESTfulAdapter.lib_restful_interface("UserOperate", "POST", request_data_dict, False)
        ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
        # 判断脚本是否正确执行
        if not ret_status_is_ok:
            # 内部通讯失败，返回内部错误 send_msg_err_list为send_msg的错误返回值
            send_msg_err_list = ["Socket path is not exist.", "Send message failed.", "Connect server failed."]
            if ret_dict.get("message") in send_msg_err_list or "client_kmc.psd invalid" in ret_dict.get("message"):
                run_log.error("Send msg failed")
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.INTERNAL_ERROR)
            run_log.error(f"The new password strength is too weak, error.")
            raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_PASSWORD_IS_WEAK)

    @staticmethod
    def get_password_valid_days(account_insecure_prompt: bool, modify_time: AnyStr) -> AnyStr:
        """
        获取密码有效期天数
        @param account_insecure_prompt: 是否为默认密码的标示
        @param modify_time: 最后修改密码的时间
        @return: 密码剩余的天数
        """
        edge_config = EdgeConfigManage.find_edge_config()
        default_expiration_days = edge_config.default_expiration_days
        if account_insecure_prompt or default_expiration_days == 0:
            password_valid_days = "--"
        else:
            pword_modify_start_time = DateUtils.get_time(modify_time)
            end_time = DateUtils.get_time(datetime.utcnow().strftime(CommonConstants.STR_DATE_FORMAT))
            if end_time < pword_modify_start_time:
                password_valid_days = "-1"
            else:
                password_valid_days = str(default_expiration_days - (end_time - pword_modify_start_time).days)
        return password_valid_days


class EdgeConfigManage(object):
    @staticmethod
    def find_edge_config() -> EdgeConfig:
        """
        根据用户id，获取当前的用户信息
        """
        with session_maker() as session:
            config_info = session.query(EdgeConfig).order_by(EdgeConfig.id).first()
            if not config_info:
                raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_EDGE_CONFIG_NOT_FOUND)
            session.expunge(config_info)
        return config_info

    @staticmethod
    def update_expiration_days(expiration_days):
        with session_maker() as session:
            session.query(EdgeConfig).update({"default_expiration_days": expiration_days})

    @staticmethod
    def update_token_timeouts(token_timeout):
        with session_maker() as session:
            session.query(EdgeConfig).update({"token_timeout": token_timeout})
