# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from flask import Blueprint

from routes.route import Route


class SessionRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        from session_service.session_views import rf_create_new_session
        from session_service.session_views import rf_get_session_service_collection
        from session_service.session_views import rf_patch_session_service_collection
        from session_service.session_views import rf_session_id
        self.blueprint.add_url_rule("", view_func=rf_get_session_service_collection, methods=["GET"])
        self.blueprint.add_url_rule("", view_func=rf_patch_session_service_collection, methods=["PATCH"])
        self.blueprint.add_url_rule("/Sessions", view_func=rf_create_new_session, methods=["POST"])
        self.blueprint.add_url_rule("/Sessions/<member_id>", view_func=rf_session_id, methods=["DELETE"])
