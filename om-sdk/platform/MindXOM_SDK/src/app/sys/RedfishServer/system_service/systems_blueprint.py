# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from flask import Blueprint
from flask import request

from common.log.logger import run_log
from routes.route_config import SUPPORT_COMPONENTS
from token_auth import get_privilege_auth

system_bp = Blueprint("systems", __name__, url_prefix="/redfish/v1/Systems")

for route_obj in SUPPORT_COMPONENTS.get("systems"):
    route_obj(system_bp).add_route()

privilege_auth = get_privilege_auth()


@system_bp.before_request
@privilege_auth.token_required
def set_endpoint_executing_log():
    """system_service进入前先记录一条运行日志"""
    if request.method in ("PATCH", "POST", "DELETE",):
        run_log.info("system service access.")
