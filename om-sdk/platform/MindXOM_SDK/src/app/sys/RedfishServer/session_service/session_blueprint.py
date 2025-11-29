# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint
from flask import request

from common.log.logger import run_log
from routes.route_config import SUPPORT_COMPONENTS
from token_auth import get_privilege_auth

privilege_auth = get_privilege_auth()
session_service_bp = Blueprint("session_service", __name__, url_prefix="/redfish/v1/SessionService")

for route_obj in SUPPORT_COMPONENTS.get("session"):
    route_obj(session_service_bp).add_route()


@session_service_bp.before_request
@privilege_auth.token_required
def set_endpoint_executing_log():
    """session_service进入前先记录一条运行日志"""
    if request.method in ("PATCH", "POST", "DELETE",):
        run_log.info("session service access.")
