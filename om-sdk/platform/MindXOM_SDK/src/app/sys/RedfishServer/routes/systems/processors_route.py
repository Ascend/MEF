# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class ProcessorsRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加Processors相关的路由
        from system_service.processors_views import rf_system_processor_ai
        from system_service.processors_views import rf_system_processor_collection
        from system_service.processors_views import rf_system_processor_cpu

        self.blueprint.add_url_rule("/Processors", view_func=rf_system_processor_collection, methods=['GET'])
        self.blueprint.add_url_rule("/Processors/CPU", view_func=rf_system_processor_cpu, methods=['GET'])
        self.blueprint.add_url_rule("/Processors/AiProcessor", view_func=rf_system_processor_ai, methods=['GET'])
