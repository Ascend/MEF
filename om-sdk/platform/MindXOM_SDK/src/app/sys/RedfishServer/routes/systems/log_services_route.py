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


class LogServicesRoute(Route):
    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加日志服务相关的路由
        from system_service.log_services_views import rf_system_log_download, rf_system_log_collect_progress
        from system_service.log_services_views import rf_system_logservices

        self.blueprint.add_url_rule("/LogServices", view_func=rf_system_logservices, methods=["GET"])
        self.blueprint.add_url_rule("/LogServices/progress", view_func=rf_system_log_collect_progress, methods=["GET"])
        self.blueprint.add_url_rule("/LogServices/Actions/download", view_func=rf_system_log_download, methods=["POST"])
