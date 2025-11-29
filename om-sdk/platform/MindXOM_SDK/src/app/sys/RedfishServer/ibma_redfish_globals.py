#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server公共变量及接口模块
修改记录：2016-10-18 创建
"""
import configparser
import functools
import json
import os
import re
import shutil
import threading
import time
from copy import deepcopy
from functools import partial
from typing import Any, AnyStr, Dict

from flask import jsonify
from flask import make_response

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.checkers import IPChecker

from common.constants import base_constants, error_codes
from common.constants.error_codes import CommonErrorCodes, FileErrorCodes
from common.constants.upload_constants import UploadConstants
from common.exception.biz_exception import Exceptions
from common.file_utils import FileUtils
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.common_check import get_error_resource, TokenCheck
from common.utils.number_utils import NumberUtils
from common.constants.product_constants import DOWNLOAD_FUNC

USER_LOCK_STATE = "The user is locked."


class RedfishGlobals(object):
    USER_NAME = ''

    Ip_errorcode = {
        "223": "Request data is invaild.",
        "224": "Parameter ip address is incorrect.",
        "225": "Parameter subnetMask is incorrect.",
        "226": "Parameter gateway is incorrect.",
        "227": "The ip address and gateway are not on the same network segment.",
        "228": "Write config file failed"
    }
    Reboot_error_code = {
        "223": "System is rebooting,can not reboot",
        "224": "Exec reboot cmd failed",
        "225": "Invalid request"
    }

    rfPort = 36665

    rfUser = "MindXOM"

    rfIPAddr = "127.0.0.1"

    rfNicName = "eth1"

    # 订阅服务的配置是否有过改变

    # 内部服务器错误

    internalErrInfo = "Internal server error"

    # 内部服务器错误对应的HTTP错误码
    internalErrCode = AppCommonMethod.INTERNAL_ERROR

    # 创建资源成功对应的响应码
    createdSuccessfullyCode = 201
    SuccessfullyCode = AppCommonMethod.OK
    creation_successfully_code = 202
    partial_successfully_code = 206
    Rese_codet = 204
    inputParamErrCode = AppCommonMethod.ERROR

    resp_json, _ = get_error_resource(TokenCheck.index_file_path)
    resp_collection_json, _ = get_error_resource(TokenCheck.index_file_path_resource)

    # 涉及重启等高风险操作的全局互斥锁
    high_risk_exclusive_lock = threading.Lock()

    @staticmethod
    def redfish_operate_adapter(request_data, operation_message):
        """
        功能描述： 记录操作日志 生成 http response
        参数：请求 操作日志信息
        返回值：http response
        异常描述：NA
        修改记录：2022-11-18创建
        """

        def wrap(fun):
            @functools.wraps(fun)
            def func(*args, **kwargs):
                operation_log = ""
                try:
                    if request_data.method in ("PATCH", "POST", "DELETE",):
                        RedfishGlobals.set_operational_log(request_data, RedfishGlobals.USER_NAME,
                                                           f"{operation_message} executing.")

                    if request_data.endpoint in DOWNLOAD_FUNC and request_data.method == "POST":
                        ret = fun(*args, **kwargs)
                        if ret.status_code in CommonConstants.STATUS_OK:
                            operation_log = f"{operation_message} success."
                        else:
                            operation_log = f"{operation_message} failed."
                        return ret

                    ret_dict, info = fun(*args, **kwargs)
                    if ret_dict.get("status") in CommonConstants.STATUS_OK:
                        operation_log = f"{operation_message} success."
                        return RedfishGlobals.update_json_of_list(info, ret_dict)
                    else:
                        operation_log = f"{operation_message} failed."
                        return RedfishGlobals.return_error_info_message(ret_dict, info)
                except Exception as err:
                    operation_log = f"{operation_message} failed."
                    run_log.error(f"{operation_message} failed, {err}")
                    ret_dict = {
                        "status": RedfishGlobals.internalErrCode,
                        "message": "operation failed."
                    }
                    return RedfishGlobals.return_error_info_message(ret_dict,
                                                                    base_constants.CommonConstants.ERR_GENERAL_INFO)
                finally:
                    if request_data.method in ("PATCH", "POST", "DELETE",) and operation_log:
                        RedfishGlobals.set_operational_log(request_data, RedfishGlobals.USER_NAME, operation_log)

            return func

        return wrap

    @staticmethod
    def init_http_server_param():
        """
        功能描述：初始化http server运行参数
        参数：无
        返回值：无
        异常描述：NA
        修改记录：2016-10-19 创建
        """
        config_parser = configparser.ConfigParser()
        dir_name = AppCommonMethod.get_project_absolute_path()
        config_full_path = os.path.join(dir_name, "config", "iBMA.ini")

        res = FileCheck.check_path_is_exist_and_valid(config_full_path)
        if not res:
            err_msg = f"Check {config_full_path} is invalid, {res.error}"
            run_log.error(err_msg)
            raise Exception(err_msg)

        if os.path.getsize(config_full_path) > base_constants.CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
            err_msg = f"Check {config_full_path} path invalid: file size greater than " \
                      f"{base_constants.CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES}"
            run_log.error(err_msg)
            raise Exception(err_msg)

        try:
            config_parser.read(config_full_path, encoding="utf-8")
        except Exception as err:
            run_log.error("Config parser file %s", err)
            raise Exception from err

        RedfishGlobals.rfPort = 36665
        get_option = partial(config_parser.get, section="iBMA_System")
        try:
            options = [get_option(option=opt) for opt in ['iBMA_user', 'iBMA_http_server_ip', 'iBMA_nic']]
        except Exception as err:
            run_log.error("fail to parse iBMA ini file, error message: %s", err)
            raise Exception from err
        RedfishGlobals.rfUser, RedfishGlobals.rfIPAddr, RedfishGlobals.rfNicName = options

    @staticmethod
    def check_external_parameter(checker_class, para_data):
        request_info = "Check external parameter failed, Because of "
        input_param_err_info = [
            CommonErrorCodes.ERROR_PARAMETER_INVALID.code,
            CommonErrorCodes.ERROR_PARAMETER_INVALID.messageKey
        ]
        error_response = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_param_err_info
        }
        try:
            check_ret = checker_class().check(para_data)
        except Exception as err:
            run_log.error("%s %s, %s", request_info, input_param_err_info, err)
            return error_response
        if not check_ret.success:
            run_log.error("%s %s", request_info, check_ret.reason)
        return error_response if not check_ret.success else None

    @staticmethod
    def create_all_required_upload_dir():
        required_dir_list = (
            UploadConstants.CERT_UPLOAD_DIR,
            UploadConstants.CONF_UPLOAD_DIR,
            UploadConstants.INI_UPLOAD_DIR,
            UploadConstants.MARK_FILE_DIR,
            UploadConstants.ZIP_UPLOAD_DIR
        )
        for required_dir in required_dir_list:
            ret = FileCreate.create_dir(required_dir, 0o1700)
            if not ret:
                run_log.error("create %s failed, error: %s", required_dir, ret.error)
                return False

        run_log.info("create all required upload dir success")

        return True

    @staticmethod
    def check_upload_file_size(filename, file_size):
        if not file_size:
            return True

        if not file_size.isdigit():
            run_log.error("upload file failed: size is not numbers")
            return False

        if len(file_size) > NumberUtils.MAX_LEN:
            run_log.error("upload file failed: length of parameter size is too large")
            return False
        try:
            file_size = int(file_size)
        except Exception as err:
            run_log.error(f"upload file failed: {err}")
            return False

        ext = os.path.splitext(filename)[1].lstrip(".").lower()
        file_max_size = UploadConstants.FILE_MAX_SIZE.get(ext)
        if file_size > file_max_size:
            run_log.error("upload file failed: over max size")
            return False

        return True

    @staticmethod
    def save_file(flask_file):
        ext = os.path.splitext(flask_file.filename)[1].lstrip(".").lower()
        file_save_dir = UploadConstants.FILE_SAVE_DIR.get(ext)
        file_max_size = UploadConstants.FILE_MAX_SIZE.get(ext)
        file_save_path = os.path.join(file_save_dir, flask_file.filename)

        if not FileUtils.delete_dir_content(file_save_dir):
            return False

        ret = FileCreate.create_file(file_save_path, 0o640)
        if not ret:
            run_log.error(f"create file {file_save_path} failed, error: {ret.error}")
            return False

        try:
            RedfishGlobals.save_content(file_save_path, flask_file, file_max_size)
        except Exception as err:
            FileUtils.delete_file_or_link(file_save_path)
            run_log.error("save file %s failed, %s", file_save_path, err)
            raise err

        return True

    @staticmethod
    def save_content(file_save_path, flask_file, file_max_size):
        with os.fdopen(os.open(file_save_path, os.O_WRONLY | os.O_CREAT, 0o640), "ab") as fstream:
            total_read_size = 0
            file_contents = iter(partial(flask_file.read, 50 * 1024 * 1024), b'')  # 每次只读50M，防止撑爆内存

            for file_content in file_contents:
                curr_read_size = len(file_content)
                if curr_read_size == 0:
                    break
                total_read_size += curr_read_size
                if total_read_size > file_max_size:
                    run_log.error("the size of the uploaded file exceeds the maximum allowed size")
                    AppCommonMethod.force_remove_file(file_save_path)
                    raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_SIZE_INVALID_ERROR)

                curr_disk_free_size = shutil.disk_usage(UploadConstants.UPLOAD_BASE_DIR).free
                if curr_disk_free_size < curr_read_size:
                    run_log.error(
                        f"{UploadConstants.UPLOAD_BASE_DIR} space not enough, "
                        f"current require {curr_read_size}, current disk free size {curr_disk_free_size}")
                    AppCommonMethod.force_remove_file(file_save_path)
                    raise Exceptions.biz_exception(error_codes.UserManageErrorCodes.ERROR_TMP_DIR_SPACE)

                fstream.write(file_content)

    @staticmethod
    def get_http_port():
        """
        功能描述：获取http port
        参数：无
        返回值：rfPort
        异常描述：NA
        修改记录：2016-10-19 创建
        """
        return RedfishGlobals.rfPort

    @staticmethod
    def set_http_port(value):
        """
        功能描述：设置http port
        参数：value 设置值
        返回值：无
        异常描述：NA
        修改记录：2016-10-19 创建
        """
        RedfishGlobals.rfPort = value
        return

    @staticmethod
    def get_http_nic_name():
        """
        功能描述：获取http nic name
        参数：无
        返回值：rfNicName
        异常描述：NA
        修改记录：2017-01-12 创建
        """
        return RedfishGlobals.rfNicName

    @staticmethod
    def get_http_user():
        """
        功能描述：获取http Server User
        参数：无
        返回值：rfUser
        异常描述：NA
        修改记录：2017-1-2 创建
        """
        return RedfishGlobals.rfUser

    @staticmethod
    def replace_kv(adict, k, v):
        AppCommonMethod.replace_kv(adict, k, v)

    @staticmethod
    def replace_kv_list(adict, copy):
        AppCommonMethod.replace_kv_list(adict, copy)

    @staticmethod
    def add_itmes(adict, k, items):
        """
        功能描述：根据key值，增加键值对
        参数：adict JSON模板
             k Key值
             items 增加的键值对
        返回值：无
        异常描述：NA
        修改记录：2016-10-19 创建
        """
        if k in adict:
            adict[k].append(items)

    @staticmethod
    def host_name_check(hostname):
        pattern = r"^(?!-)[A-Za-z0-9\-]{1,64}(?<!-)$"
        return re.fullmatch(pattern, hostname)

    @staticmethod
    def replace_id(resp_json, data_id):
        """
        功能描述：替换模板中的ID
        参数：resp_json JSON 模板
             data_id 替换的ID
        返回值：True/False
        异常描述：无
        修改记录：2016-10-19 创建
        """
        resp_json['Id'] = data_id
        resp_json['@odata.id'] = resp_json['@odata.id'].replace('oDataID', data_id)

    @staticmethod
    def replaceinfo(resp_json, data_id, start_time, request_data_dict, task_percentage):
        """
        功能描述：替换模板中的ID
        参数：resp_json JSON 模板
             Id 替换的ID
             starttime 开始任务时间
             Message 修改的信息
             TaskPercenttage
        返回值：True/False
        异常描述：无
        修改记录：2016-10-19 创建
        """
        resp_json["Id"] = data_id
        resp_json['@odata.id'] = resp_json['@odata.id'].replace('oDataID', data_id)
        resp_json["IPv4Addresses"] = request_data_dict["IPv4Addresses"]
        resp_json["Oem"]["StartTime"] = start_time
        resp_json["Oem"]["TaskPercentage"] = task_percentage

    @staticmethod
    def return_error_info_message(error_message_dict: Dict[AnyStr, Any], input_err_info: str):
        """
        向前端返回错误信息，接收字符参数
        @param error_message_dict:接收的错误信息，里面包含status、message
        :return: 交 flask 框架进行处理
        """
        status: int = error_message_dict.get("status")
        # 没有获取消息返回内部错误
        ret_dict = {
            'status': base_constants.CommonConstants.ERR_CODE_500,
            'message': [CommonErrorCodes.ERROR_INTERNAL_SERVER.code, CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey]
        }
        if status != base_constants.CommonConstants.ERR_CODE_500:
            try:
                return RedfishGlobals.update_json_of_error_info(error_message_dict, input_err_info)
            except Exception as err_info:
                run_log.error(f'{err_info}')
                return RedfishGlobals.update_json_of_error_info(ret_dict, "GeneralError")

        return RedfishGlobals.update_json_of_error_info(ret_dict, "GeneralError")

    @staticmethod
    def update_json_of_list(resp_json, ret_dict, prefix_id=""):
        """
        功能描述：调用LibAdapter中的接口获取相关信息时
        返回的retDict字典中包含key(status)
        和key(message),若retDict数据正常则
        更新resp_json的对应数据
        参数：resp_json为对应的资源模板,
             retDict字典中正常情况下包含key(status)和 key(message)
        返回值：http response
        异常描述：NA
        修改记录：2016-11-22 创建
        修改记录：2016-12-06 Response类的Content-type默认值已经修改为application/json
        """
        if "message" in ret_dict:
            # 集合类资源更新resp_json的对应数据
            if isinstance(ret_dict["message"], list):
                RedfishGlobals.replace_mem_count_kv_list(resp_json, ret_dict["message"], prefix_id)
            # 非集合类资源更新resp_json的对应数据
            else:
                RedfishGlobals.replace_kv_list(resp_json, ret_dict["message"])
        return make_response(json.dumps(resp_json), ret_dict.get("status"))

    @staticmethod
    def replace_mem_count_kv_list(resp_json, ret_dict_message, prefix_id):
        if len(ret_dict_message) == 0:
            RedfishGlobals.replace_kv(resp_json, "Members@odata.count", 0)
            del resp_json["Members"][0]
            return resp_json
        else:
            RedfishGlobals.replace_kv(resp_json, "Members@odata.count", len(ret_dict_message))
            data_id = resp_json["Members"][0]["@odata.id"]
            if prefix_id != "":
                data_id = data_id.replace("oDataPrefixID", prefix_id)

            for i, _ in enumerate(ret_dict_message):
                if i == 0:
                    resp_json["Members"][0]["@odata.id"] = data_id. \
                        replace("entityID", ret_dict_message[i])
                else:
                    url = data_id.replace("entityID", ret_dict_message[i])
                    RedfishGlobals.add_itmes(resp_json, "Members",
                                             {"@odata.id": url})
            return resp_json

    @staticmethod
    def make_response(input_info, input_status):
        """
        功能描述： 生产对应Response类
        参数：inputInfo 输入参数 ,inputStatus HTTP 状态码
        返回值：http response
        异常描述：NA
        修改记录：2016-11-26 创建
        """
        ret = {}
        ret["status"] = input_status
        ret["message"] = input_info

        # 生成对应的http response
        res = make_response(json.dumps(ret), input_status)
        return res

    @staticmethod
    def make_error_response(input_info, input_status):
        """
        功能描述： 生产对应Response类
        参数：inputInfo 输入参数 ,inputStatus HTTP 状态码
        返回值：http error response
        异常描述：NA
        修改记录：2022-10-28 创建
        """
        ret = {}
        ret["status"] = input_status
        ret["message"] = input_info

        return RedfishGlobals.update_json_of_error_info(ret, "GeneralError")

    @staticmethod
    def make_update_file_response(input_info, input_status):
        """
        功能描述： 生产对应Response类
        参数：inputInfo 输入参数 ,inputStatus HTTP 状态码
        返回值：http response
        异常描述：NA
        修改记录：2016-11-26 创建
        """
        response = jsonify({"status": input_status, "message": input_info})
        response.status_code = input_status
        response.headers['X-Content-Type-Options'] = 'nosniff'
        response.headers['X-XSS-Protection'] = "1; mode=block"
        response.headers['Strict-Transport-Security'] = "max-age=31536000; includeSubDomains"
        response.headers['Pragma'] = 'no-cache'
        response.headers["X-Frame-Options"] = 'DENY'
        response.headers["Content-Security-Policy"] = "script-src 'self' 'unsafe-eval'; " \
                                                      "form-action 'self'; frame-ancestors 'self'; plugin-types 'none'"
        response.headers["Referrer-Policy"] = "same-origin"
        response.headers["Cache-control"] = "no-cache, no-store, must-revalidate"
        response.headers["Expires"] = 0
        # 生成对应的http response
        return response

    @staticmethod
    def make404_error_resp():
        """
        功能描述：返回404错误
        参数：无
        返回值：http response
        异常描述：NA
        修改记录：2017-1-7 创建
        """
        description = "The requested URL was not found on the server"
        return RedfishGlobals.make_error_response(description, 404)

    @staticmethod
    def replace_target_id(resp_json, data_id):
        """
        功能描述：替换模板中target的ID
        参数：resp_json JSON 模板
             Id 替换的ID
        返回值：True/False
        异常描述：无
        修改记录：2019-1-30 创建
        """
        resp_json['Id'] = data_id
        resp_json['@odata.id'] = resp_json['@odata.id'].replace('oDataID', data_id)

    @staticmethod
    def update_json_of_error_info(ret_dict, input_err_info):
        """
        功能描述：调用LibAdapter中的接口获取相关信息时
        返回error信息，错误的集中处理
        参数： retDict字典中正常情况下包含key(status)和 key(message)
             inputErrInfo为错误描述信息.
        返回值：http response
        异常描述：NA
        修改记录：2019-1-31 创建
        """
        resp_json = deepcopy(RedfishGlobals.resp_json)
        resp_collection_json = deepcopy(RedfishGlobals.resp_collection_json)
        if ret_dict is None or not isinstance(ret_dict, dict):
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            input_err_info = "GeneralError"
            ret_dict = {'status': RedfishGlobals.internalErrCode, 'message': RedfishGlobals.internalErrInfo}

        # retDict字典中是否包含key(status)
        has_status_key = "status" in ret_dict

        # retDict字典中是否包含key(message)
        has_message_key = "message" in ret_dict

        if not has_status_key or not has_message_key:
            # retDict字典中未同时包含key(status)和key(message)的异常处理(生成对应的http response)
            input_err_info = "GeneralError"
            ret_dict = {'status': RedfishGlobals.internalErrCode, 'message': RedfishGlobals.internalErrInfo}

        # retDict字典中同时包含key(status)和key(message)

        # 检测retDict字典的key(status)是否为HTTP OK
        if ret_dict["status"] == 200:
            input_err_info = "GeneralError"
            ret_dict = {'status': RedfishGlobals.internalErrCode, 'message': RedfishGlobals.internalErrInfo}

        message_id = ret_dict["message"]
        if input_err_info == "GeneralError":
            message_id = "GeneralError"
            resp_collection_json["Messages"][message_id]['Message'] = ret_dict["message"]
            if isinstance(ret_dict["message"], list):
                resp_json['error']["@Message.ExtendedInfo"][0]['Oem']['status'] = ret_dict["message"][0]
                resp_collection_json["Messages"][message_id]['Message'] = ret_dict["message"][1]
        error_message = resp_collection_json["Messages"][message_id]
        resp_json['error']['code'] = resp_json['error']['code'].replace(
            'errorID', message_id)
        resp_json['error']['message'] = resp_json['error']['message'].replace(
            'oData', message_id)
        # 更新resp_json的对应数据
        AppCommonMethod.replace_kv_list(resp_json, error_message)
        respsonecode = error_message["ErrCode"]
        if ret_dict["status"] != 400:
            respsonecode = ret_dict["status"]
        error_key = ret_dict.get("errorKey", None)
        error_value = ret_dict.get("errorValue", None)
        param_types = ret_dict.get("ParamTypes", None)
        number_of_args = ret_dict.get("NumberOfArgs", None)
        if error_key:
            resp_json['error']["@Message.ExtendedInfo"][0]['Message'] = \
                resp_json['error']["@Message.ExtendedInfo"][0]['Message'].replace(
                    'oData', error_key)
        if error_value:
            resp_json['error']["@Message.ExtendedInfo"][0]['Message'] = \
                resp_json['error']["@Message.ExtendedInfo"][0]['Message'].replace(
                    'OValue', error_value)
        if param_types:
            resp_json['error']["@Message.ExtendedInfo"][0]['ParamTypes'] = param_types
        if number_of_args:
            resp_json['error']["@Message.ExtendedInfo"][0]['NumberOfArgs'] = number_of_args
        # 生成对应的http response
        res = make_response(json.dumps(resp_json),
                            respsonecode)
        return res

    @staticmethod
    def check_json_request(request_data):
        """
        功能描述：校验requestbody是否是json
        参数：request_data  request body请求体里的字符串
        返回值：无
        异常描述：NA
        修改记录：2019-2-1 创建
        """

        err_code = RedfishGlobals.inputParamErrCode
        ret = {'status': err_code, 'message': "MalformedJSON"}
        try:
            # 解码并通过转换判断字符串是否为 json
            request_data_dict = json.loads(request_data.decode("utf-8"))
        except Exception:
            run_log.error("ERR.0%s,Request data is not json.", err_code)
            return [1, ret]
        if not isinstance(request_data_dict, dict) or request_data_dict == {}:
            return [1, ret]
        return [0, request_data_dict]

    @staticmethod
    def set_operational_log(request, username, message):
        request_ip = request.headers.get('X-Real-Ip')
        res = IPChecker("data").check({"data": request_ip})
        if not res.success:
            run_log.error("request ip invalid: %s", res.reason)
            request_ip = None
        operate_log.info("[%s@%s] %s" % (username, request_ip, message))
        return

    @staticmethod
    def check_input_parm(parm):
        return AppCommonMethod.check_input_parm(parm)

    @staticmethod
    def update_alarm_time_stamp(ret_dict):
        if ret_dict is None or not isinstance(ret_dict, dict):
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return

        # retDict字典中是否包含key(status)
        has_status_key = "status" in ret_dict

        # retDict字典中是否包含key(message)
        has_message_key = "message" in ret_dict

        # 检测retDict字典中是否同时包含key(status)和key(message)
        if not has_status_key or not has_message_key:
            # retDict字典中未同时包含key(status)
            # 和key(message)的异常处理(生成对应的http response)
            return

        # 检测retDict字典的key(status)是否为HTTP OK
        if ret_dict["status"] != 200:
            # retDict字典中key(status)为非200错误处理
            # (生成对应的http response)
            return

        # 更新retDict的对应数据
        alarm_list = ret_dict.get("message").get("AlarMessages")
        for tmp_dict in alarm_list:
            if "Timestamp" in tmp_dict:
                tmp_dict["Timestamp"] = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(int(tmp_dict["Timestamp"])))
        return

    @staticmethod
    def get_user_locked_state(ret_dict):
        """
        返回消息中"message"状态码为110206时记录锁定用户操作日志
        """
        ret_status = ret_dict.get("status", "")
        ret_message = ret_dict.get("message", "")
        return ret_status == AppCommonMethod.ERROR and ret_message[0] == 110206

    @staticmethod
    def check_flask_file_size_is_zero(file):
        if len(file.stream.read(1)) == 0:
            run_log.error("file size is 0")
            error_code = FileErrorCodes.ERROR_FILE_CONTENT_NULL_ERROR
            return error_code.code, RedfishGlobals.return_error_info_message(
                {"status": base_constants.CommonConstants.ERR_CODE_400,
                 "message": [error_code.code, "Flask file size is 0"]},
                base_constants.CommonConstants.ERR_GENERAL_INFO)

        file.stream.seek(0)
        return 0, "Flask file size is not 0"
