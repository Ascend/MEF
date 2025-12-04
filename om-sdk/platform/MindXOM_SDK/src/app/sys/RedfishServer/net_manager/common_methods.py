# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import functools
import json
from copy import deepcopy
from typing import AnyStr
from typing import List
from typing import Union
from urllib import parse

from flask import Response, make_response

from common.ResourceDefV1.resource import RfResource
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import CommonErrorCodes
from common.constants.product_constants import SERVICE_ROOT
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.exec_cmd import ExecCmd
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.constants import ApiOperationMapper
from net_manager.exception import NetManagerException, ValidateParamsError, LockedError, DbOperateException


class ApiHelper:
    @staticmethod
    def catch_api_error(operation_message):
        """
        功能描述：将异常原因统一记录到运行日志中
        """
        def wrap(fun):
            @functools.wraps(fun)
            def func(*args, **kwargs):
                err_msg = f"{operation_message} failed."
                try:
                    ret = fun(*args, **kwargs)
                except ValidateParamsError as err:
                    run_log.error("%s Because of %s", err_msg, err)
                    return Serializer().make_400_response(err.err_msg, err.err_code)
                except LockedError as err:
                    run_log.error("%s Because of %s", err_msg, err)
                    return Serializer().make_400_response(err.err_msg, err.err_code)
                except DbOperateException as err:
                    run_log.error("%s Because of %s", err_msg, err)
                    return Serializer().make_400_response(err_msg, err.err_code)
                except NetManagerException as err:
                    run_log.error("%s Because of %s", err_msg, err)
                    return Serializer().make_400_response(err_msg, err.err_code)
                except Exception as err:
                    run_log.error("%s Because of %s", err_msg, err)
                    return Serializer().make_500_response()

                if isinstance(ret, list) and len(ret) > 1 and ret[0] != 0:
                    if ret[0] == 206:
                        return Serializer().make_206_response(ret)
                    return Serializer().make_400_response(ret)

                return ret

            return func
        return wrap

    @staticmethod
    def get_api_operation(request_data):
        url = request_data.url.split("/v1/")[1]
        method_url = f"{request_data.method.upper()} {url}"
        return ApiOperationMapper.OPERATION_MAPPER[method_url]

    @staticmethod
    def validate_params(checker, request_data: AnyStr) -> dict:
        """
        功能描述：外部参数校验触发函数
        """
        err_msg = "Parameter is invalid."
        params = json.loads(parse.unquote(request_data.decode()))

        try:
            check_ret = checker.check(params)
        except Exception as err:
            raise ValidateParamsError(err_msg) from err

        if not check_ret.success:
            raise ValidateParamsError(check_ret.reason)

        return params

    @staticmethod
    def check_server_name_forbidden(server_name: str) -> bool:
        """
        检查FD的域名配置参数是否在黑名单中
        :param server_name: 待配置的FD域名
        :return: True表示server_name在黑名单中，False表示可以配置
        """
        if server_name in AppCommonMethod.FORBIDDEN_DOMAINS:
            return True

        ret = AppCommonMethod.get_etc_host_content_lines()
        if not ret:
            # 获取etc hosts内容失败，默认校验失败，不允许配置域名，采用默认的域名
            run_log.error("get etc host content lines failed: %s", ret.error)
            return True

        local_ips = ("127.0.0.1", "::1")
        local_domains = set()
        for line in ret.data:
            line = line.strip()
            if not line:
                continue

            address = line.split()
            if address[0] in local_ips:
                local_domains.update(address[1:])

        return server_name in local_domains

    @staticmethod
    def send_operational_log(request, msg):
        from ibma_redfish_urls import RedfishURIs
        RedfishURIs.send_operational_log(request, msg)


class Serializer:
    root = SERVICE_ROOT
    service: RfResource

    resp_collection_json = {
        "Description": "Indicates that a general error has occurred.",
        "Message": "A general error has occurred. See ExtendedInfo for more information.",
        "Severity": "Critical",
        "ErrCode": 400,
    }

    def make_200_response(self, data: dict, status: int) -> Response:
        """
        功能描述：正常流程返回逻辑
        """
        resp_json = json.loads(self.service.get_resource())
        AppCommonMethod.replace_kv_list(resp_json, data)
        return make_response(json.dumps(resp_json), status)

    def make_206_response(self, err_msg) -> Response:
        """
        功能描述：部分成功返回逻辑
        """
        ret = {
            "status": CommonConstants.ERR_CODE_206,
            "message": err_msg,
        }
        return self._update_general_error(ret)

    def make_400_response(self, err_msg: Union[str, List], err_code=None) -> Response:
        """
        功能描述：捕获已定义过的异常，返回请求异常
        """
        if isinstance(err_msg, str) and err_code:
            err_msg = [err_code, err_msg]

        ret = {
            "status": CommonConstants.ERR_CODE_400,
            "message": err_msg,
        }
        return self._update_general_error(ret)

    def make_500_response(self) -> Response:
        """
        功能描述：捕获到非预期的 Exception 时，返回内部错误
        """
        ret = {
            "status": CommonConstants.ERR_CODE_500,
            "message": [CommonErrorCodes.ERROR_INTERNAL_SERVER.code, CommonErrorCodes.ERROR_INTERNAL_SERVER.messageKey]
        }
        return self._update_general_error(ret)

    def _update_general_error(self, ret_dict):
        resp_json = deepcopy(RedfishGlobals.resp_json)
        resp_collection_json = deepcopy(self.resp_collection_json)
        if isinstance(ret_dict["message"], list):
            resp_json["error"]["@Message.ExtendedInfo"][0]["Oem"]["status"] = ret_dict["message"][0]
            resp_collection_json["Message"] = ret_dict["message"][1]
        else:
            resp_collection_json["Message"] = ret_dict["message"]

        resp_json["error"]["code"] = resp_json["error"]["code"].replace("errorID", "GeneralError")
        resp_json["error"]["message"] = resp_json["error"]["message"].replace("oData", "GeneralError")
        AppCommonMethod.replace_kv_list(resp_json, resp_collection_json)

        return make_response(json.dumps(resp_json), ret_dict["status"])


def docker_root_is_mounted() -> bool:
    """检查/var/lib/docker是否已挂载，用于限制FD切换"""
    try:
        is_a500 = LibRESTfulAdapter.lib_restful_interface(
            "System", "GET", None, False
        ).get("message").get("SupportModel") == CommonConstants.ATLAS_500_A2
    except Exception as err:
        run_log.error("Get system info failed, catch %s", err.__class__.__name__)
        # 获取信息失败时，应当考虑成内部存在问题，返回False限制切换
        return False

    # 不是A500切换FD不受docker根目录是否存在影响
    if not is_a500:
        return True

    docker_root = "/var/lib/docker"
    return ExecCmd.exec_cmd_use_pipe_symbol(f"df {docker_root} | grep {docker_root}")[0] == 0
