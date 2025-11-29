# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class ActionRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # Action类的URL
        from system_service.action_views import rf_restore_defaults
        from system_service.action_views import rf_system_reset

        self.blueprint.add_url_rule("/Actions/ComputerSystem.Reset", view_func=rf_system_reset, methods=["POST"])
        self.blueprint.add_url_rule("/Actions/RestoreDefaults.Reset", view_func=rf_restore_defaults, methods=["POST"])
