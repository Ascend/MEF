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

"""
功    能：安全服务资源 URL处理模块
"""
import configparser
import json
import os
import shutil
import threading

from flask import g
from flask import request
from flask import send_from_directory

from common.checkers.param_checker import CertAlarmTimeInfoChecker
from common.checkers.param_checker import ImportServerCertificateChecker
from common.checkers.param_checker import PunyDictDeleteChecker
from common.checkers.param_checker import PunyDictImportChecker
from common.checkers.param_checker import SecurityLoadChecker
from common.checkers.param_checker import SecurityLoadImportChecker
from common.constants.base_constants import CommonConstants
from common.constants.base_constants import UserManagerConstants
from common.constants.error_codes import FileErrorCodes
from common.constants.error_codes import SecurityServiceErrorCodes
from common.constants.upload_constants import UploadConstants
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from system_service.systems_common import make_error_dict
from system_service.systems_serializer import HttpsCertAlarmTimeResourceSerializer
from system_service.systems_serializer import HttpsCertResourceSerializer
from system_service.systems_serializer import SecurityLoadResourceSerializer
from system_service.systems_serializer import SecurityServiceResourceSerializer
from upload_mark_file import UploadMarkFile
from user_manager.user_manager import UserManager

PUNY_DICT_EXPORT_LOCK = threading.Lock()
PUNY_DICT_DELETE_LOCK = threading.Lock()
DOWNLOAD_CSR_LOCK = threading.Lock()
SECURITY_LOAD_LOCK = threading.Lock()
IMPORT_SECURITY_LOAD_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Query security service info")
def rf_security_service():
    """
    功能描述：查询安全服务顶层资源
    """
    input_err_info = "Query security service info failed."
    try:
        # 获取资源文件
        security_service_resource = json.loads(SecurityServiceResourceSerializer().service.get_resource())
        return {"status": 200}, security_service_resource
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict("ResourceDoesNotExists", CommonConstants.ERR_CODE_404)
        return ret_dict, CommonConstants.ERR_PARTICULAR_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query security service")
def rf_https_cert():
    """
    功能描述：查询SSL证书资源信息
    """
    input_err_info = "Query security service failed."
    try:
        # 获取资源模板
        https_cert_resource = json.loads(HttpsCertResourceSerializer().service.get_resource())
        https_cert = "HttpsCert"
        ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityService", "GET", None, False, https_cert)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, https_cert_resource
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


IMPORT_CERTIFICATE_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Import custom certificate")
def rf_import_custom_certificate():
    """
    功能描述：导入服务器证书
    """
    if IMPORT_CERTIFICATE_LOCK.locked():
        message = "Import custom certificate failed because SecurityServiceViews modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with IMPORT_CERTIFICATE_LOCK:
        message = "Import custom certificate failed."
        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0 or recv_conf_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(ImportServerCertificateChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                "user_id": g.user_id,
                "password": password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message} {CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 获取Username和requestIp放入字典request_data_Dict
            request_data_dict["_User"] = RedfishGlobals.USER_NAME
            request_data_dict["_Xip"] = request.headers["X-Real-Ip"]
            ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityService", "POST", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)

            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Import custom certificate successfully."
                ret_dict = {"status": RedfishGlobals.creation_successfully_code, "message": {"Message": message}}
                return ret_dict, success_message_resource

            # 导入证书成功但存在合法性问题
            if isinstance(ret_dict["message"], list) and \
                    SecurityServiceErrorCodes.ERROR_IMPORT_CERTIFICATE_LEGALITY_RISKY.code == ret_dict["message"][0]:
                partial_success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Import custom certificate successfully, but legality is risky."
                ret_dict = {"status": RedfishGlobals.partial_successfully_code, "message": {"Message": message}}
                return ret_dict, partial_success_message_resource

            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        finally:
            # 业务完成后清除文件上传标记（必须）
            run_log.warning(message)
            UploadMarkFile.clear_upload_mark_file(
                os.path.join(UploadConstants.MARK_FILE_DIR, UploadConstants.MARK_FILES.get("cert")))
            FileUtils.delete_dir_content(UploadConstants.CERT_UPLOAD_DIR)


PUNY_DICT_IMPORT_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Import puny dict")
def rf_import_system_puny_dict():
    """
    功能描述：导入弱字典相关配置
    参数：无
    返回值：http response
    异常描述：NA
    """
    if PUNY_DICT_IMPORT_LOCK.locked():
        message = "Import puny dict failed because SecurityServiceViews import is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with PUNY_DICT_IMPORT_LOCK:
        message = "Import puny dict failed."
        try:
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            # 对密码进行校验
            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(PunyDictImportChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            message = "Import puny dict failed."
            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                "user_id": g.user_id,
                "password": password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message} {CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 调用接口LibAdapter获取 Softwares 相关信息
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            upload_file_path = os.path.join(UploadConstants.CONF_UPLOAD_DIR, request_data_dict.get("FileName"))
            if not os.path.exists(upload_file_path):
                run_log.error("Check file %s doesn't exist", upload_file_path)
                ret_dict = make_error_dict("Check upload file doesn't exist", CommonConstants.ERR_CODE_400)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            constant_file_path = os.path.join(UploadConstants.CONF_UPLOAD_DIR, UploadConstants.WEAK_DICT_NAME)
            os.rename(upload_file_path, constant_file_path)
            req_data = {"OperationType": CommonConstants.TYPE_IMPORT}
            ret_dict = LibRESTfulAdapter.lib_restful_interface("PunyDict", "POST", req_data)
            if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Import puny dict successfully."
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message},
                }
                return ret_dict, success_message_resource

            run_log.error(message)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        finally:
            # 业务完成后清除文件上传标记（必须）
            UploadMarkFile.clear_upload_mark_file(os.path.join(UploadConstants.MARK_FILE_DIR,
                                                               UploadConstants.MARK_FILES.get("conf")))
            FileUtils.delete_dir_content(UploadConstants.CONF_UPLOAD_DIR)


@RedfishGlobals.redfish_operate_adapter(request, "Export puny dict")
def rf_export_system_puny_dict():
    """
    功能描述：导出弱字典相关配置
    参数：无
    返回值：http response
    异常描述：NA
    """
    if PUNY_DICT_EXPORT_LOCK.locked():
        run_log.error("Export puny dict is busy.")
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "Export puny dict is busy."}
        return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
    with PUNY_DICT_EXPORT_LOCK:
        try:
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            FileUtils.delete_full_dir(UploadConstants.CONF_UPLOAD_DIR)
            FileCreate.create_dir(UploadConstants.CONF_UPLOAD_DIR, mode=0o1700)
            req_data = {"OperationType": CommonConstants.TYPE_EXPORT}
            ret_dict = LibRESTfulAdapter.lib_restful_interface("PunyDict", "POST", req_data)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                run_log.error("Export puny dict failed, %s", ret_dict["message"])
                if "ERR.008" in ret_dict["message"]:
                    err_code = FileErrorCodes.ERROR_FILE_NOT_EXIST_ERROR
                    ret_dict["message"] = [err_code.code, err_code.messageKey]
                return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            puny_dict = ret_dict["message"]["export_file"]
            res = FileCheck.check_path_is_exist_and_valid(puny_dict)
            if not res:
                run_log.error("%s path invalid : %s", puny_dict, res.error)
                ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "puny_dict.conf is invalid"}
                return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            # 将文件拷贝到MindXOM用户目录下，提供下载，并通知monitor进程删除临时日志文件
            shutil.copyfile(puny_dict, os.path.join(UploadConstants.CONF_UPLOAD_DIR, os.path.basename(puny_dict)))
            LibRESTfulAdapter.lib_restful_interface("PunyDict", "DELETE", None)
            ret = send_from_directory(UploadConstants.CONF_UPLOAD_DIR, os.path.basename(puny_dict), as_attachment=True)
            run_log.info("send_from_directory %s", ret)
            return ret
        except Exception as err:
            run_log.error("export puny dict failed, reason is: %s", err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "export puny dict failed"}
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
        finally:
            FileUtils.delete_full_dir(UploadConstants.CONF_UPLOAD_DIR)


@RedfishGlobals.redfish_operate_adapter(request, "Delete puny dict")
def rf_delete_system_puny_dict():
    """
    功能描述：删除弱字典相关配置
    参数：无
    返回值：http response
    异常描述：NA
    """
    if PUNY_DICT_DELETE_LOCK.locked():
        message = "Delete puny dict failed because SecurityServiceViews Delete is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with PUNY_DICT_DELETE_LOCK:
        message = "Delete puny dict failed."
        try:
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            # 对密码进行校验
            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(PunyDictDeleteChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            message = "Delete puny dict failed."
            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                "user_id": g.user_id,
                "password": password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message} {CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 调用接口LibAdapter获取 Softwares 相关信息
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            req_data = {"OperationType": CommonConstants.TYPE_DELETE}
            ret_dict = LibRESTfulAdapter.lib_restful_interface("PunyDict", "POST", req_data)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Delete puny dict successfully."
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message},
                }
                return ret_dict, success_message_resource

            run_log.error(message)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Query HttpsCertAlarmTime service")
def rf_get_system_https_cert_alarm_time():
    """
    功能描述：查询证书过期提醒时间
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query HttpsCertAlarmTime service failed."
    try:
        # 获取资源模板
        httpscert_alarm_time_resource = json.loads(HttpsCertAlarmTimeResourceSerializer().service.get_resource())
        message = "Query HttpsCertAlarmTime service failed."
        ret_dict = LibRESTfulAdapter.lib_restful_interface("CertAlarmTime", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, httpscert_alarm_time_resource

        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


HTTPSCERT_ALARM_TIME_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Set CertAlarmTime")
def rf_modify_system_https_cert_alarm_time():
    """
    功能描述：修改证书过期提醒时间
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if HTTPSCERT_ALARM_TIME_LOCK.locked():
        message = "Set CertAlarmTime failed because SecurityServiceViews modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
    with HTTPSCERT_ALARM_TIME_LOCK:
        message = "Set CertAlarmTime failed."
        try:
            # 获取资源模板
            httpscert_alarm_time_resource = json.loads(HttpsCertAlarmTimeResourceSerializer().service.get_resource())
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(CertAlarmTimeInfoChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                "user_id": g.user_id,
                "password": password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message} {CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            ret_dict = LibRESTfulAdapter.lib_restful_interface(
                "CertAlarmTime", "PATCH", request_data_dict, False)

            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                return ret_dict, httpscert_alarm_time_resource

            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Download CSR file")
def rf_download_csr_file():
    """
    功能描述：下载CSR文件
    参数：无
    返回值：http response
    异常描述：NA
    """
    if DOWNLOAD_CSR_LOCK.locked():
        run_log.error("Download CSR failed because SecurityServiceViews export is busy.")
        ret_dict = make_error_dict("SecurityServiceViews export is busy.", CommonConstants.ERR_CODE_400)
        return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
    with DOWNLOAD_CSR_LOCK:
        if not os.path.exists(UploadConstants.CSR_UPLOAD_DIR) and \
                not FileCreate.create_dir(UploadConstants.CSR_UPLOAD_DIR, 0o1700):
            run_log.error("Create path run_web_cert failed.")
            ret_dict = make_error_dict("Create path run_web_cert failed.", CommonConstants.ERR_CODE_400)
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
        file_name = "x509Req.pem"
        file_path = os.path.join(UploadConstants.CSR_UPLOAD_DIR, file_name)
        try:
            ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityServiceCSR", "POST", None, False)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

            # 获取证书内容写入到文件后，提供给用户下载
            with os.fdopen(os.open(file_path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600), "w") as file:
                file.write(ret_dict["message"]["x509rep_content"])
            ret = send_from_directory(UploadConstants.CSR_UPLOAD_DIR, file_name, as_attachment=True)
            run_log.info("File sent successfully.")
            return ret
        except Exception as err:
            run_log.error("export csr failed reason is: %s", err)
            ret_dict = make_error_dict("export csr failed", CommonConstants.ERR_CODE_400)
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
        finally:
            FileUtils.delete_file_or_link(file_path)


@RedfishGlobals.redfish_operate_adapter(request, "Query securityLoad info")
def rf_get_security_load_config_info():
    """
    功能描述：获取登录规则信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query SecurityLoad info failed."
    try:
        # 获取资源模板
        security_load_resource = json.loads(SecurityLoadResourceSerializer().service.get_resource())
        # 调用接口SecurityLoad获取
        ret = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "GET", None, False)
        if LibRESTfulAdapter.check_status_is_ok(ret):
            return ret, security_load_resource
        return ret, CommonConstants.ERR_GENERAL_INFO
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Config security load")
def rf_modify_security_load_config_info():
    """
    功能描述：配置登录规则信息
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    if SECURITY_LOAD_LOCK.locked():
        message = "Config security load failed because SecurityServiceViews modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with SECURITY_LOAD_LOCK:
        message = "Config security load failed."
        try:
            req_param_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(req_param_dict)
            if ret[0] != 0 or req_param_dict is None:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(SecurityLoadChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                'oper_type': UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                'user_id': g.user_id,
                'password': password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message}{CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 调用接口LibAdapter获取 Softwares 相关信息
            # 根据调用接口LibAdapter返回的retDict字典，生产对应的json,返回给web
            ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "PATCH", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Config security load successfully."
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message}
                }
                return ret_dict, success_message_resource
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Export configuration of security load")
def rf_export_security_load():
    """登录配置导出接口"""
    def export_rules_to_file(sec_cfg, target_file):
        cfg_parser = configparser.ConfigParser()
        for idx, cfg_item in enumerate(sec_cfg):
            section = f"cfg_item{idx}"
            cfg_parser.add_section(section)
            for option in ("enable", "start_time", "end_time", "ip_addr", "mac_addr"):
                value = cfg_item.get(option) or "None"
                if option == "mac_addr":
                    value = cfg_item.get(option).upper() if cfg_item.get(option) else "None"

                cfg_parser.set(section, option, value)
        # 登录规则写入文件
        FileUtils.delete_full_dir(UploadConstants.INI_UPLOAD_DIR)
        FileCreate.create_dir(UploadConstants.INI_UPLOAD_DIR, mode=0o1700)
        with os.fdopen(os.open(target_file, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w") as file:
            cfg_parser.write(file)

    session_sec_cfg_ini = "session_sec_cfg.ini"
    try:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "POST", {"action": "export"}, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            ret_dict = make_error_dict("Export configuration of security load failed", CommonConstants.ERR_CODE_400)
            return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

        # 将规则写入文件，提供下载
        export_rules_to_file(ret_dict["message"]["load_cfg"],
                             os.path.join(UploadConstants.INI_UPLOAD_DIR, session_sec_cfg_ini))
        ret = send_from_directory(UploadConstants.INI_UPLOAD_DIR, session_sec_cfg_ini, as_attachment=True)
        return ret

    except Exception as err:
        run_log.error("export login rules failed, %s", err)
        ret_dict = make_error_dict("export login rules failed", CommonConstants.ERR_CODE_400)
        return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)
    finally:
        FileUtils.delete_file_or_link(os.path.join(UploadConstants.INI_UPLOAD_DIR, session_sec_cfg_ini))


@RedfishGlobals.redfish_operate_adapter(request, "Import configuration of security load")
def rf_import_security_load():
    """登录配置导入接口"""
    if IMPORT_SECURITY_LOAD_LOCK.locked():
        message = "Import configuration of security load failed because SecurityServiceViews modify is busy."
        run_log.error(message)
        ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with IMPORT_SECURITY_LOAD_LOCK:
        message = "Import configuration of security load failed."
        try:
            recv_conf_dict = request.get_data()
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if recv_conf_dict is None or ret[0] != 0:
                ret_dict = ret[1]
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(SecurityLoadImportChecker, request_data_dict)
            if check_ret is not None:
                ret_dict = make_error_dict(check_ret.get("message"), check_ret.get("status"))
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            # 二次身份认证,调用插件层对用户进行密码校验
            password = request_data_dict.get("Password")
            recv_conf_dict = {
                'oper_type': UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                'user_id': g.user_id,
                'password': password
            }
            ret_dict = UserManager().get_all_info(recv_conf_dict)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if not ret_status_is_ok:
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"{message}{CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO
            # 导入文件
            request_data_dict['action'] = 'import'
            ret_dict = LibRESTfulAdapter.lib_restful_interface("SecurityLoad", "POST", request_data_dict, False)
            ret_status_is_ok = LibRESTfulAdapter.check_status_is_ok(ret_dict)
            if ret_status_is_ok:
                success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
                message = "Import configuration of security load successfully."
                ret_dict = {
                    "status": RedfishGlobals.creation_successfully_code,
                    "message": {"Message": message},
                }
                return ret_dict, success_message_resource
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
        finally:
            # 业务完成后清除文件上传标记（必须）
            UploadMarkFile.clear_upload_mark_file(os.path.join(UploadConstants.MARK_FILE_DIR,
                                                               UploadConstants.MARK_FILES.get("ini")))
            FileUtils.delete_dir_content(UploadConstants.INI_UPLOAD_DIR)
