# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint
from flask import request

from common.log.logger import run_log
from routes.route_config import SUPPORT_COMPONENTS
from token_auth import get_privilege_auth

privilege_auth = get_privilege_auth()

account_service_bp = Blueprint("account_service", __name__, url_prefix="/redfish/v1/AccountService")

for route_obj in SUPPORT_COMPONENTS.get("account"):
    route_obj(account_service_bp).add_route()


@account_service_bp.before_request
@privilege_auth.token_required
def set_endpoint_executing_log():
    """account_service进入前先记录一条运行日志"""
    if request.method in ("PATCH", "POST", "DELETE",):
        run_log.info("account service access.")