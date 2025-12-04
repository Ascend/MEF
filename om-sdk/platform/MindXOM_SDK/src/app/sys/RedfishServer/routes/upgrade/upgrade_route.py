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

from routes.route import Route


class UpgradeRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        from upgrade_service.upgrade_service_view import get_upgrade_service_resource
        from upgrade_service.upgrade_service_view import get_upgrade_service_actions
        from upgrade_service.upgrade_service_view import rf_upgrade_service_actions
        from upgrade_service.upgrade_service_view import rf_upgrade_reset_actions

        self.blueprint.add_url_rule("", view_func=get_upgrade_service_resource, methods=["GET"])
        self.blueprint.add_url_rule("/Actions/UpdateService.SimpleUpdate",
                                    view_func=get_upgrade_service_actions, methods=["GET"])
        self.blueprint.add_url_rule("/Actions/UpdateService.SimpleUpdate",
                                    view_func=rf_upgrade_service_actions, methods=["POST"])
        self.blueprint.add_url_rule("/Actions/UpdateService.Reset",
                                    view_func=rf_upgrade_reset_actions, methods=["POST"])
