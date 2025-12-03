# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import threading
from typing import Union

from flask import Blueprint, request, g

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from net_manager.checkers.external_params_checker import NetManageConfigChecker, CertUploadChecker, CrlUploadChecker
from net_manager.common_methods import ApiHelper
from net_manager.constants import NetManagerConstants
from net_manager.exception import LockedError, ValidateParamsError
from net_manager.manager.fd_cert_manager import FdCertManager
from net_manager.manager.import_manager import ImportCrl, ImportCert
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.manager.net_switch_manager import WebNetSwitchManager, FdNetSwitchManager, NetSwitchManagerBase
from net_manager.models import NetManager
from net_manager.serializers import (GetNetManageConfigSerializer, GetNodeIdSerializer,
                                     GetFdCertSerializer, ImportFdCrlSerializer, ImportFdCertSerializer)
from token_auth import get_privilege_auth

privilege_auth = get_privilege_auth()
net_manager_bp = Blueprint("net_manager_https_request", __name__, url_prefix="/redfish/v1/NetManager")
import_cert_lock = threading.Lock()
import_crl_lock = threading.Lock()
net_manager_lock = threading.Lock()


@net_manager_bp.before_request
@privilege_auth.token_required
def set_executing_log():
    """
    功能描述：在视图函数执行前执行，将操作的开始统一记录到操作日志中
    """
    ApiHelper.send_operational_log(request, f"{ApiHelper.get_api_operation(request)} executing.")
    # 说明token校验通过，需要记录操作日志
    g.hasErrorBeforeRequest = False


@net_manager_bp.after_request
def set_response_operational_log(response):
    """
    功能描述：在视图函数执行后执行，将操作的结果统一记录到日志中
    """
    # 防止token校验失败时记录操作日志，通过状态标志判断
    if hasattr(g, "hasErrorBeforeRequest"):
        if response.status_code in CommonConstants.STATUS_OK:
            run_log.info("%s success.", ApiHelper.get_api_operation(request))
            ApiHelper.send_operational_log(request, f"{ApiHelper.get_api_operation(request)} success.")
        else:
            # 由于要记录异常原因，失败的运行日志在 ApiHelper.catch_api_error 里通过捕获异常来记录
            ApiHelper.send_operational_log(request, f"{ApiHelper.get_api_operation(request)} failed.")
    return response


@net_manager_bp.route("", methods=["POST"])
@ApiHelper.catch_api_error("Configuring net manage")
def rf_config_net_manage():
    """
    功能描述：配置网管资源信息
    """
    if net_manager_lock.locked():
        raise LockedError("Net manager POST API is busy.")

    with net_manager_lock:
        request_data = ApiHelper.validate_params(NetManageConfigChecker(), request.get_data())
        params = {}
        manager_type = request_data.get("ManagerType")
        if manager_type == "FusionDirector":
            server_name = request_data.get("ServerName")
            if server_name and ApiHelper.check_server_name_forbidden(server_name):
                run_log.error("Net manager config failed: invalid parameter ServerName")
                raise ValidateParamsError("Net manager config checkers: invalid parameter ServerName")

            params = {
                "ManagerType": request_data.get("ManagerType"),
                "NetIP": request_data.get("NetIP"),
                "NetAccount": request_data.get("NetAccount"),
                "NetPassword": request_data.get("NetPassword"),
                "test": request_data.get("test"),
                "ServerName": server_name,
                "Port": request_data.get("Port"),
                "NodeId": request_data.get("NodeId"),
            }
        switch_manager: NetSwitchManagerBase = {
            NetManagerConstants.WEB: WebNetSwitchManager(params),
            NetManagerConstants.FUSION_DIRECTOR: FdNetSwitchManager(params),
        }.get(manager_type)
        ret = switch_manager.switch_deal()
        if ret and ret[0] != 0:
            return ret

        net_manager_info = NetCfgManager().get_net_cfg_info().to_dict_for_query()
        return GetNetManageConfigSerializer().make_200_response(net_manager_info, CommonConstants.ERR_CODE_200)


@net_manager_bp.route("", methods=["GET"])
@ApiHelper.catch_api_error("Querying net manage config")
def rf_get_net_manage_config():
    """
    功能描述：查询网管资源信息
    """
    net_manager_info = NetCfgManager().get_net_cfg_info().to_dict_for_query()
    return GetNetManageConfigSerializer().make_200_response(net_manager_info, CommonConstants.ERR_CODE_200)


@net_manager_bp.route("/NodeID", methods=["GET"])
@ApiHelper.catch_api_error("Querying node id")
def rf_get_node_id():
    """
    功能描述：查询网管节点ID
    """
    node_id: str = NetCfgManager().get_net_cfg_info().node_id
    return GetNodeIdSerializer().make_200_response({"NodeConnectID": node_id}, CommonConstants.ERR_CODE_200)


@net_manager_bp.route("/QueryFdCert", methods=["GET"])
@ApiHelper.catch_api_error("Querying fd cert")
def rf_get_fd_cert():
    """
    功能描述：查询 FD 根证书内容
    """
    net_manager: NetManager = NetCfgManager().get_net_cfg_info()
    cert_info = FdCertManager(net_manager.ip, net_manager.port).get_cert_info(net_manager.net_mgmt_type,
                                                                              net_manager.status)
    return GetFdCertSerializer().make_200_response(cert_info, CommonConstants.ERR_CODE_200)


@net_manager_bp.route("/ImportFdCert", methods=["POST"])
@ApiHelper.catch_api_error("Importing FD Cert")
def rf_import_fd_cert():
    """
    功能描述：导入 FD 根证书
    """
    if import_cert_lock.locked():
        raise LockedError("Importing FD cert is busy.")

    with import_cert_lock:
        info = ImportCert().import_deal(import_contents(CommonConstants.MAX_CERT_LIMIT, CertUploadChecker()))
        return ImportFdCertSerializer().make_200_response(info, CommonConstants.ERR_CODE_200)


@net_manager_bp.route("/ImportFdCrl", methods=["POST"])
@ApiHelper.catch_api_error("Importing FD Crl")
def rf_import_fd_crl():
    """
    功能描述：导入 FD 吊销列表
    """
    if import_crl_lock.locked():
        raise LockedError("Importing FD crl is busy.")

    with import_crl_lock:
        info = ImportCrl().import_deal(import_contents(CommonConstants.MAX_CRL_LIMIT, CrlUploadChecker()))
        return ImportFdCrlSerializer().make_200_response(info, CommonConstants.ERR_CODE_200)


def import_contents(max_content_length: int, checker: Union[CrlUploadChecker, CertUploadChecker]) -> str:
    try:
        img_file = request.files["imgfile"]
        check_ret = checker.check({"imgfile": img_file.filename})
    except Exception as err:
        raise ValidateParamsError("Parameter is invalid.") from err
    if not check_ret.success:
        raise ValidateParamsError(check_ret.reason)
    content_size = img_file.seek(0, 2)
    img_file.seek(0, 0)
    if not 0 < content_size <= max_content_length:
        raise ValidateParamsError("upload file failed: invalid file size.")
    # size为冗余字段，前端未传入，但需要保留该字段用作兼容，判断逻辑与之前版本一致
    f_size = request.form.get("size")
    if f_size is not None and not (f_size.isdigit() and 0 < int(f_size) <= max_content_length):
        raise ValidateParamsError("upload file failed: invalid file size.")
    try:
        return img_file.stream.read().decode()
    except Exception as err:
        raise ValidateParamsError("upload file failed: invalid contents.") from err
