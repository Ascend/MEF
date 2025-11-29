# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint, request

from common.log.logger import run_log
from routes.route_config import SUPPORT_COMPONENTS
from token_auth import get_privilege_auth

https_upgrade_service_bp = Blueprint("UpdateService_view", __name__, url_prefix="/redfish/v1/UpdateService")

for route_obj in SUPPORT_COMPONENTS.get("upgrade"):
    route_obj(https_upgrade_service_bp).add_route()

privilege_auth = get_privilege_auth()


@https_upgrade_service_bp.before_request
@privilege_auth.token_required
def set_endpoint_executing_log():
    """请求进入前先记录一条日志"""
    if request.method != "GET":
        run_log.info("Start upgrade task executing.")
