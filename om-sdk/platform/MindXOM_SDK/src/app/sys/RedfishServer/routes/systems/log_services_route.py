# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
