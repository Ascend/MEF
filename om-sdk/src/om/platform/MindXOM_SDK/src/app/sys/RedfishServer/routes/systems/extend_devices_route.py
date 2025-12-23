# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
