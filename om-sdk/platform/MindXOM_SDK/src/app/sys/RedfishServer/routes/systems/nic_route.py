# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class NicRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加Ethernet相关的路由
        from system_service.nic_views import rf_get_system_ethernet_collection
        from system_service.nic_views import rf_modify_system_ethernet_eth_x
        from system_service.nic_views import rf_get_system_ethernet_eth_x

        self.blueprint.add_url_rule("/EthernetInterfaces", view_func=rf_get_system_ethernet_collection, methods=['GET'])
        self.blueprint.add_url_rule("/EthernetInterfaces/<eth_id>", view_func=rf_get_system_ethernet_eth_x,
                                    methods=['GET'])
        self.blueprint.add_url_rule("/EthernetInterfaces/<eth_id>",
                                    view_func=rf_modify_system_ethernet_eth_x, methods=["PATCH"])
