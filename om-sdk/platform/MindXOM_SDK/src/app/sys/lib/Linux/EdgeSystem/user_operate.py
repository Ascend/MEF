#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import ctypes
import os

from common.checkers import PasswordComplexityChecker
from common.constants.base_constants import CommonConstants, UserManagerConstants
from common.init_cmd import cmd_constants
from common.log.logger import run_log
from common.utils.common_check import CommonCheck
from common.utils.system_utils import SystemUtils
from devm.device_mgr import DEVM
from common.checkers import RegexStringChecker, IPChecker
from lib.Linux.systems.security_service.security_load import SecurityLoad
from common.common_methods import CommonMethods


class UserOperate:

    @staticmethod
    def check_security_config(param_dict) -> list:
        if not CommonCheck.check_sub_list(["real_ip"], list(param_dict.keys())):
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "Parameter is invalid."]
        real_ip = param_dict.get("real_ip")
        ret = IPChecker("data").check({"data": real_ip})
        if not ret.success or not SecurityLoad.check_session_request_usr_data(real_ip):
            run_log.error("check login ip failed.")
            return [CommonMethods.ERROR, "real_ip"]
        return [CommonMethods.OK, "check success"]

    @staticmethod
    def check_weak_dict(param_dict):
        # 检查系统是否存在默认弱口令字典配置，不存在则不需要检查弱口令
        for file in ("pw_dict.hwm", "pw_dict.pwd", "pw_dict.pwi"):
            file_path = os.path.join(CommonConstants.OS_CRACKLIB_DIR, file)
            if not os.path.exists(file_path):
                run_log.warning("Check OS cracklib %s file is invalid, skip weak dictionary check.", file)
                return [CommonMethods.OK, "OK"]

        if not CommonCheck.check_sub_list(["Password"], list(param_dict.keys())):
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "Parameter is invalid."]
        password = param_dict.get("Password")
        # 密码复杂度，密码要求包含至少三种字符
        check_ret = PasswordComplexityChecker(attr_name="Password", min_len=8, max_len=20).check({"Password": password})
        if not check_ret.success:
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "checker password complexity failed."]

        if cmd_constants.OS_NAME in ("EulerOS", "openEuler"):
            cracklib_so_path = "/usr/lib64/libcrack.so.2"
        else:
            cracklib_so_path = "/usr/lib/aarch64-linux-gnu/libcrack.so.2"

        try:
            pam_cracklib_so = ctypes.CDLL(cracklib_so_path)
            fascist_check = pam_cracklib_so.FascistCheck
            fascist_check.argtypes = (ctypes.c_char_p, ctypes.c_char_p)
            fascist_check.restype = ctypes.c_char_p
            pw_dict = CommonConstants.OS_PW_DICT_DIR
            ret = fascist_check(password.encode(encoding='utf-8'), pw_dict.encode(encoding='utf-8'))
        except Exception as err:
            run_log.error("load weak dict so failed: %s", err)
            return [CommonMethods.ERROR, "check weak dict failed."]

        if ret:
            err_msg = ret.decode(encoding="utf-8")
            run_log.error(f"The new password strength is too weak, error: {err_msg}")
            return [CommonMethods.ERROR, "The new password strength is too weak."]

        return [CommonMethods.OK, "OK"]

    @staticmethod
    def set_os_pwd(param_dict):
        if not CommonCheck.check_sub_list(["Username", "Password"], list(param_dict.keys())):
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "Parameter is invalid."]
        username = param_dict.get("Username", None)
        password = param_dict.get("Password")
        # 密码复杂度，密码要求包含至少三种字符
        check_ret = PasswordComplexityChecker(attr_name="Password", min_len=8, max_len=20).check({"Password": password})
        if not check_ret.success:
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "checker password complexity failed."]
        username_partern = r'^[a-zA-Z0-9]*[a-zA-Z]+[a-zA-Z0-9]*$'
        check_ret = RegexStringChecker(attr_name="Username", match_str=username_partern,
                                       min_len=1, max_len=16).check({"Username": username})
        if not check_ret.success:
            run_log.error("Parameter is invalid.")
            return [CommonMethods.ERROR, "checker username failed."]

        try:
            DEVM.get_device("system0").set_attribute("user_info", {"user_name": username, "passwd_info": password})
        except Exception:
            run_log.error("reset user passwd failed.")
            return [CommonMethods.ERROR, "reset user passwd failed."]

        run_log.info("modify os user info successfully")
        return [CommonMethods.OK, "OK"]

    def post_request(self, param_dict):
        oper_type = param_dict.get("oper_type")
        if not oper_type:
            return [CommonMethods.ERROR, "Parameter is invalid."]
        del param_dict['oper_type']
        if UserManagerConstants.OPER_TYPE_CHECK_WEAK_DICT == oper_type:
            return self.check_weak_dict(param_dict)
        elif UserManagerConstants.OPER_TYPE_CHECK_SECURITY_CONFIG == oper_type:
            return self.check_security_config(param_dict)
        elif UserManagerConstants.OPER_TYPE_RESET_OS_PASSWORD == oper_type:
            # 如果非A500则不重置OS密码
            return self.set_os_pwd(param_dict) if SystemUtils().is_a500 else [CommonMethods.OK, "OK"]
        else:
            run_log.error("operate type is invalid")
            return [CommonMethods.ERROR, "operate type is invalid."]
