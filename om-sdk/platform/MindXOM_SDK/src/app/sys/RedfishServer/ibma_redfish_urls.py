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

"""
功    能：Redfish Server URL处理模块
修改记录：2016-10-18 创建
"""
import functools
import json
import os
import ssl
import threading

from flask import Flask
from flask import Response
from flask import make_response
from flask import request

import lib_restful_adapter as LibRESTfulAdapter
from common.checkers.param_checker import FirmwareInventoryChecker
from common.constants import base_constants
from common.constants import error_codes
from common.constants.error_codes import CommonErrorCodes
from common.exception.biz_exception import BizException
from common.exception.biz_exception import Exceptions
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.exception_utils import ExceptionUtils
from common.kmc_lib.kmc import Kmc
from common.kmc_lib.tlsconfig import TlsConfig
from ibma_redfish_globals import RedfishGlobals
from register_blueprint import register_blueprint
from restful_extend_interfaces import EXTEND_RESTFUL_REGISTER_FUNCTIONS_PATH
from token_auth import get_privilege_auth
from upload_mark_file import UploadMarkFile

CERT_FILE = "/home/data/config/redfish/server_kmc.cert"
KEY_FILE = "/home/data/config/redfish/server_kmc.priv"
PSD_FILE = "/home/data/config/redfish/server_kmc.psd"
CERT_PRIMARY_KSF = "/home/data/config/redfish/om_cert.keystore"
CERT_STANDBY_KSF = "/home/data/config/redfish/om_cert_backup.keystore"
ALG_CONFIG_FILE = "/home/data/config/redfish/om_alg.json"


class RedfishURIs:
    """
    功能描述：HTTP服务的URI路由定义类
    接口：NA
    修改记录：2016-10-31 创建
    """
    APP = Flask(__name__, static_folder=None)

    METHOD_NOT_ALLOWED_ERROR_CODE = 405

    libConfigMutex = threading.Lock()

    @staticmethod
    def user_operational_log(request_data, operation_message):
        def wrap(fun):
            @functools.wraps(fun)
            def func(*args, **kwargs):
                try:
                    if request_data.method != "GET":
                        RedfishURIs.send_operational_log(request_data, f"{operation_message} executing.")
                    return fun(*args, **kwargs)
                except Exception as err:
                    run_log.error(f"{operation_message} failed, {err}")
                    return make_response(json.dumps({"status": 500, "message": "operation failed."}), 500)

            return func

        return wrap

    @staticmethod
    def send_operational_log(req, msg, print_log=True, user_name=None):
        if not print_log:
            return
        if not user_name:
            user_name = RedfishGlobals.USER_NAME
        if req.method in ('PATCH', 'POST', 'DELETE',):
            RedfishGlobals.set_operational_log(req, user_name, msg)

    @staticmethod
    def file_request_error(error_msg):
        if str(error_msg) == '[Errno 28] No space left on device' or \
                '[Errno 2] No usable temporary directory found in' in str(error_msg):
            run_log.error("request imgfile failed: %s." % error_msg)
            input_info = "The size of uploaded file is larger than the remaining capacity."
            run_log.error(f'{input_info}')
            return input_info
        else:
            input_info = "request imgfile failed:{}".format(error_msg)
            run_log.error(f'{input_info}')
            return input_info

    @staticmethod
    def rf_api_ibma_server(root, versions, host, port, silence=True):
        """
        功能描述：启动flask, 对URL进行处理
        接口：NA
        修改记录：2016-10-18 创建
        修改记录：2016-12-06 创建Response类的子类,使用自定义的默认值
        修改记录：2017-01-14 删除对远端地址的限制
        """
        app = RedfishURIs.APP
        app.response_class = ResMimeTypeOfJson
        app.config['MAX_CONTENT_LENGTH'] = 512 * 1024 * 1024
        # init ssl context
        ssl_ctx = None
        # 注册蓝图
        register_blueprint(app, EXTEND_RESTFUL_REGISTER_FUNCTIONS_PATH)

        # 加入数据库配置
        privilege_auth = get_privilege_auth()
        try:
            # 获取ssl context
            ssl_ctx = RedfishURIs._get_ssl_context()
        except Exception as err:
            run_log.error("Init app error, %s", err)
            if not silence:
                raise Exception("get ssl context failed") from err

        @app.errorhandler(404)
        def handle_not_found(e):
            return RedfishGlobals.make404_error_resp()

        @app.errorhandler(Exception)
        def handle_common_exception(ex):
            if ex.code == RedfishURIs.METHOD_NOT_ALLOWED_ERROR_CODE:
                return ex

            ret = {
                "status": CommonErrorCodes.ERROR_INTERNAL_SERVER.code,
                "message": CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey
            }
            return make_response(json.dumps(ret), base_constants.CommonConstants.ERR_CODE_500)

        @app.route("/redfish/v1/UpdateService/FirmwareInventory", methods=['POST'])
        @privilege_auth.token_required
        @RedfishURIs.user_operational_log(request, "Upload file")
        def rf_update_file():
            # 公共上传文件的接口
            # 由于提供的升级包的大小在几百兆左右，所以限制文件大小待定
            # 1.指定文件存储的目标路径
            oper_log = "Upload file failed."
            try:

                if RedfishURIs.libConfigMutex.locked():
                    return RedfishGlobals.make_response("Make file dir failed.",
                                                        base_constants.CommonConstants.ERR_CODE_400)

                with RedfishURIs.libConfigMutex:

                    # 创建需要的目录
                    if not RedfishGlobals.create_all_required_upload_dir():
                        return RedfishGlobals.make_response("Make file dir failed.",
                                                            base_constants.CommonConstants.ERR_CODE_400)

                    # 获取文件
                    try:
                        file = request.files['imgfile']
                    except Exception as error:
                        run_log.error(f"Upload file failed, error: {error}")
                        return RedfishGlobals.make_response(RedfishURIs.file_request_error(error),
                                                            base_constants.CommonConstants.ERR_CODE_400)

                    # 校验文件名字
                    filename = file.filename
                    check_ret = RedfishGlobals.check_external_parameter(FirmwareInventoryChecker, {"imgfile": filename})
                    if check_ret is not None:
                        return RedfishGlobals.make_error_response(check_ret.get("message"), check_ret.get("status"))
                    file_type = os.path.splitext(filename)[1].lstrip(".").lower()
                    run_log.info("The uploaded file is %s" % filename)
                    oper_log = f"Upload {file_type} file failed."

                    ret_code, ret_info = RedfishGlobals.check_flask_file_size_is_zero(file)
                    if ret_code != 0:
                        return ret_info

                    # 检查文件大小
                    if not RedfishGlobals.check_upload_file_size(filename, request.form.get('size')):
                        raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_FILE_SIZE_INVALID_ERROR)

                    # 校验文件上传标记是否有效
                    if not UploadMarkFile.check_upload_mark_file_is_invalid_by_filename(filename):
                        msg = f"Forbidden to upload {file_type} new file, please wait for the completion of " \
                              "processing or 10 minutes later"
                        run_log.error(f'{msg}')
                        raise Exceptions.biz_exception(error_codes.FileErrorCodes.ERROR_UPLOAD_FLAG)

                    # 保存文件
                    if not RedfishGlobals.save_file(file):
                        msg = f"Upload {file_type} file failed"
                        run_log.error(msg)
                        return RedfishGlobals.make_response(msg, base_constants.CommonConstants.ERR_CODE_400)

                    # 创建文件上传标记
                    if not UploadMarkFile.create_upload_mark_file_by_filename(filename):
                        msg = f"Create upload mark file ({file_type} type) failed."
                        run_log.error(msg)
                        return RedfishGlobals.make_response(msg, base_constants.CommonConstants.ERR_CODE_400)

                    # 返回成功消息
                    input_status = RedfishGlobals.creation_successfully_code
                    msg = "Upload [%s] file successfully." % filename
                    run_log.info(msg)
                    oper_log = f"Upload {file_type} file successfully."
                    return RedfishGlobals.make_update_file_response(msg, input_status)

            except Exception as error:
                run_log.error(f"Upload file failed, error: {error}")
                if isinstance(error, BizException):
                    return RedfishGlobals.return_error_info_message(
                        {"status": base_constants.CommonConstants.ERR_CODE_400,
                         "message": ExceptionUtils.exception_process(error)},
                        base_constants.CommonConstants.ERR_GENERAL_INFO)
                return RedfishGlobals.make_response("Upload file failed.", base_constants.CommonConstants.ERR_CODE_400)
            finally:
                RedfishURIs.send_operational_log(request, oper_log)

        RedfishURIs._start_flask(app, ssl_ctx)

    @staticmethod
    def _get_ssl_context():
        for file in (CERT_PRIMARY_KSF, CERT_STANDBY_KSF, ALG_CONFIG_FILE, PSD_FILE):
            res = FileCheck.check_path_is_exist_and_valid(file)
            if not res:
                raise ValueError(f"{file} path invalid : {res.error}")

        kmc = Kmc(CERT_PRIMARY_KSF, CERT_STANDBY_KSF, ALG_CONFIG_FILE)
        with open(PSD_FILE, "r") as psd:
            encrypted = psd.read()
        decrypted = kmc.decrypt(encrypted)
        is_normal_value, ctx = TlsConfig.get_ssl_context(None, CERT_FILE, KEY_FILE, pwd=decrypted)
        if not is_normal_value:
            err_info = "get ssl context error."
            run_log.error(f'{err_info}')
            raise Exception(err_info)
        decrypted = None
        return ctx

    @staticmethod
    def _start_flask(app, ssl_ctx):
        rf_port = RedfishGlobals.get_http_port()
        ipv4 = '127.0.0.1'
        try:
            app.run(host=ipv4, port=rf_port, ssl_context=ssl_ctx)
        except Exception as e:
            run_log.error("start redfish-server fail: %s" % e)


class ResMimeTypeOfJson(Response):
    """
    功能描述：创建Response类的子类,
    定义自己的默认值(Content-Type默认类型改为application/json)
    接口：NA
    修改记录：2016-10-31 创建
    """

    default_mimetype = 'application/json'
