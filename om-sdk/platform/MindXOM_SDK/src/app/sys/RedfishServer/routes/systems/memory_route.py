# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class MemoryRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加内存相关的URL
        from system_service.memory_views import rf_system_memory_summary_collection

        self.blueprint.add_url_rule("/Memory", view_func=rf_system_memory_summary_collection, methods=["GET"])
