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
