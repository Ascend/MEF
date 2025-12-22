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
import re
from typing import AnyStr
from typing import List

from common.constants import error_codes
from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.utils.app_common_method import AppCommonMethod
from common.checkers import IPChecker
from common.utils.result_base import Result


class CommonCheck:
    @staticmethod
    def check_file_stat(file_path):
        """
        功能描述：校验文件的路径、有效性、属组是否为 MindXOM、文件大小是否为 1MB
        """
        if not FileCheck.check_path_is_exist_and_valid(file_path):
            return Result(result=False, err_msg=f"Check {file_path} is invalid")

        if not FileCheck.check_path_mode_owner_group(file_path, user="MindXOM", group="MindXOM"):
            return Result(result=False, err_msg=f"Check {file_path} is invalid, because of invalid owner.")

        if os.path.getsize(file_path) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES or os.path.getsize(file_path) <= 0:
            return Result(result=False, err_msg=f"Check {file_path} is invalid, because file size is invalid")

        return Result(result=True)

    @staticmethod
    def check_all_param_not_empty(parm_list) -> bool:
        """
          检查所有的参数是否为空.
          :param parm_list: list集合
          :return: True:所有参数不为空。False：有部分参数为空。
        """
        if isinstance(parm_list, list) and not all(parm_list):
            return False
        if isinstance(parm_list, dict):
            for _, value in parm_list.items():
                if not value:
                    return False
        return True

    @staticmethod
    def check_sub_list(default_list: List[AnyStr], pars_list: List[AnyStr]) -> bool:
        """
        判断子集list是否在父集list中
        """
        return set(pars_list).issubset(set(default_list))

    @staticmethod
    def check_check_code(parm):
        "校验sha256校验码格式是否正确"
        reg_str = "^[a-f0-9]{64}$"
        pattern_str = re.compile(reg_str)
        if not pattern_str.fullmatch(parm):
            return False
        return True

    @staticmethod
    def check_operator(operator_name, operator_ip) -> Result:
        """
        公共的操作对象校验函数，规则:
                1、操作用户名，长度为1~16个字符，支持数字和英文字母，不能全为数字
                2、操作对象IP地址满足IPV4或者IPV6格式要求
                3、如果字段为空，则不进行校验
        :param:
                operator_name: 操作用户名
                operator_ip: 操作对象IP地址
        :return:
                True: 满足要求
                False: 不满足要求
        """
        if operator_name:
            pattern = r"^[a-zA-Z0-9]*[a-zA-Z]+[a-zA-Z0-9]*$"
            if len(operator_name) > 16 or not re.fullmatch(pattern, operator_name):
                return Result(result=False, err_msg="operator name invalid")

        if operator_ip:
            check_result = IPChecker("data").check({"data": operator_ip})
            if not check_result.success:
                return Result(result=False, err_msg="operator ip invalid")

        return Result(result=True)


def get_error_resource(index_file_path):
    res_data = json.dumps(TokenCheck.account_error_code)
    res = FileCheck.check_path_is_exist_and_valid(index_file_path)
    if not res:
        return res_data, f"{index_file_path} path invalid : {res.error}"

    try:
        with open(index_file_path, "r") as res_file:
            res_raw_data = res_file.read()
            res_data = json.loads(res_raw_data)
    except Exception as ex:
        return json.dumps(TokenCheck.account_error_code), f'{ex}'
    return res_data, ""


class TokenCheck:
    account_error_code = {"err_message": "Illegal request"}
    # 获取工程目录
    projectDir = AppCommonMethod.get_project_absolute_path()
    profile_mock_up_path = os.path.join(projectDir, "common", "MockupData", "iBMAServerV1")
    index_file_path = os.path.join(profile_mock_up_path, "redfish", "v1", "ErrorCollection", "index.json")
    index_file_path_resource = os.path.join(profile_mock_up_path, "redfish", "v1", "ErrorCollection", "1", "index.json")

    @staticmethod
    def token_check(message: list, request_url: str = None):
        # 获取error模板资源
        resp_json, _ = get_error_resource(TokenCheck.index_file_path)
        resp_collection_json, _ = get_error_resource(TokenCheck.index_file_path_resource)
        if not isinstance(message, list) or not isinstance(message[0], int):
            ret_dict = {
                'status': CommonConstants.ERR_CODE_400,
                'message': "PropertyMissing",
                "errorKey": 'X-Auth-Token'
            }
            return TokenCheck.update_json_of_error_info(
                resp_json, resp_collection_json, ret_dict, "PropertyMissing")

        error_key = int(message[0])
        ret_dict = {'status': CommonConstants.ERR_CODE_400, 'message': ''}
        if error_codes.UserManageErrorCodes.ERROR_SESSION_NOT_FOUND.code == error_key:
            # 数据库中未找到当前的token信息
            ret_dict['message'] = 'AccountForSessionNoLongerExists'
            input_err_info = "AccountForSessionNoLongerExists"
        elif error_codes.UserManageErrorCodes.ERROR_REQUEST_IP_ADDR.code == error_key:
            # 请求的ip地址错误
            ret_dict['message'] = 'AccountForSessionNoLongerExists'
            input_err_info = "AccountForSessionNoLongerExists"
        elif error_codes.UserManageErrorCodes.ERROR_SESSION_TIME_LIMIT.code == error_key:
            # session超期
            ret_dict['message'] = 'SessionTimeLimitExceeded'
            input_err_info = "SessionTimeLimitExceeded"
        elif error_codes.UserManageErrorCodes.ERROR_PASSWORD_VALID_DAY.code == error_key:
            # 密码失效
            ret_dict['message'] = 'PasswordValidDaysError'
            input_err_info = "PasswordValidDaysError"
        else:
            # token参数未传递
            ret_dict['message'] = 'PropertyMissing'
            ret_dict['errorKey'] = 'X-Auth-Token'
            input_err_info = "PropertyMissing"

        ret_dict['errorKey'] = ret_dict.get('errorKey') or error_key
        return TokenCheck.update_json_of_error_info(
            resp_json, resp_collection_json, ret_dict, input_err_info)

    @staticmethod
    def update_json_of_error_info(resp_json, resp_collection_json, ret_dict, input_err_info):
        """
        功能描述：调用LibAdapter中的接口获取相关信息时
             返回error信息，错误的集中处理
        :param resp_json: 为对应的资源模板
        :param resp_collection_json: error模板集合
        :param ret_dict:
        :param input_err_info: retDict字典中正常情况下包含key(status)和 key(message)
        :return:
        """

        if ret_dict is None or not isinstance(ret_dict, dict):
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            input_err_info = "GeneralError"
            ret_dict = {'status': CommonConstants.ERR_CODE_500, 'message': "Internal server error"}

        # retDict字典中是否包含key(status)
        has_status_key = "status" in ret_dict

        # retDict字典中是否包含key(message)
        has_message_key = "message" in ret_dict

        if not has_status_key or not has_message_key:
            # retDict字典中未同时包含key(status)和key(message)的异常处理(生成对应的http response)
            input_err_info = "GeneralError"
            ret_dict = {'status': CommonConstants.ERR_CODE_500, 'message': "Internal server error"}

        # retDict字典中同时包含key(status)和key(message)

        # 检测retDict字典的key(status)是否为HTTP OK
        if ret_dict["status"] == 200:
            input_err_info = "GeneralError"
            ret_dict = {'status': CommonConstants.ERR_CODE_500, 'message': "Internal server error"}

        message_id = ret_dict["message"]
        if input_err_info == "GeneralError":
            message_id = "GeneralError"
            resp_collection_json["Messages"][message_id]['Message'] = ret_dict["message"]
        errormessage = resp_collection_json["Messages"][message_id]
        resp_json['error']['code'] = resp_json['error']['code'].replace('errorID', message_id)
        resp_json['error']['message'] = resp_json['error']['message'].replace('oData', message_id)
        # 更新resp_json的对应数据
        AppCommonMethod.replace_kv_list(resp_json, errormessage)
        respsonecode = errormessage["ErrCode"]
        if ret_dict["status"] != 400:
            respsonecode = ret_dict["status"]
        error_key = ret_dict.get("errorKey")
        error_value = ret_dict.get("errorValue")
        param_types = ret_dict.get("ParamTypes")
        number_of_args = ret_dict.get("NumberOfArgs")
        if error_key:
            resp_json['error']["@Message.ExtendedInfo"][0]['Message'] = \
                resp_json['error']["@Message.ExtendedInfo"][0]['Message'].replace('oData', str(error_key))
            resp_json['error']["@Message.ExtendedInfo"][0]['Oem']['status'] = error_key
        if error_value:
            resp_json['error']["@Message.ExtendedInfo"][0]['Message'] = \
                resp_json['error']["@Message.ExtendedInfo"][0]['Message'].replace('OValue', error_value)
        if param_types:
            resp_json['error']["@Message.ExtendedInfo"][0]['ParamTypes'] = param_types
        if number_of_args:
            resp_json['error']["@Message.ExtendedInfo"][0]['NumberOfArgs'] = number_of_args
        # 生成对应的http response
        return [respsonecode, json.dumps(resp_json)]


