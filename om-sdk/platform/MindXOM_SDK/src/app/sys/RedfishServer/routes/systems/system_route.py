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


class SystemRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        from system_service.systems_views import rf_eth_ip_list
        from system_service.systems_views import rf_get_system_collection
        from system_service.systems_views import rf_get_system_time
        from system_service.systems_views import rf_patch_system_collection

        self.blueprint.add_url_rule("", view_func=rf_get_system_collection, methods=["GET"])
        self.blueprint.add_url_rule("", view_func=rf_patch_system_collection, methods=["PATCH"])
        # 添加系统时间以及中间件查询的路由
        self.blueprint.add_url_rule("/SystemTime", view_func=rf_get_system_time, methods=['GET'])
        # 添加以太网口的路由
        self.blueprint.add_url_rule("/EthIpList", view_func=rf_eth_ip_list, methods=["GET"])
