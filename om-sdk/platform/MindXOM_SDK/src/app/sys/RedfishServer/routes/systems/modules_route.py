# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class ModulesRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        from system_service.module_views import rf_get_system_device_info
        from system_service.module_views import rf_get_system_module_collection
        from system_service.module_views import rf_get_system_module_info
        from system_service.module_views import rf_patch_system_device_info
        # 添加Devm相关的URL
        self.blueprint.add_url_rule("/Modules", view_func=rf_get_system_module_collection, methods=["GET"])
        self.blueprint.add_url_rule("/Modules/<module_id>", view_func=rf_get_system_module_info, methods=["GET"])
        self.blueprint.add_url_rule("/Modules/<module_id>/<device_id>", view_func=rf_get_system_device_info,
                                    methods=["GET"])
        self.blueprint.add_url_rule("/Modules/<module_id>/<device_id>",
                                    view_func=rf_patch_system_device_info, methods=["PATCH"])
