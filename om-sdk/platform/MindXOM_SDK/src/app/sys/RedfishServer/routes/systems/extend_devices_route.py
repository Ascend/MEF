# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class ExtendDevicesRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加外部设备相关的URL
        from system_service.extended_devices_views import rf_system_extended_device_info
        from system_service.extended_devices_views import rf_system_extended_devices_collection

        self.blueprint.add_url_rule("/ExtendedDevices", view_func=rf_system_extended_devices_collection,
                                    methods=["GET"])
        self.blueprint.add_url_rule("/ExtendedDevices/<extend_id>", view_func=rf_system_extended_device_info,
                                    methods=["GET"])
