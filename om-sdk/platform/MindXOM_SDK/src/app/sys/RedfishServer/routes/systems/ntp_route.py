# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class NtpRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加NTP相关的URL
        from system_service.ntp_views import rf_get_ntp_service_collection
        from system_service.ntp_views import rf_patch_ntp_service_collection

        self.blueprint.add_url_rule("/NTPService", view_func=rf_get_ntp_service_collection, methods=["GET"])
        self.blueprint.add_url_rule("/NTPService", view_func=rf_patch_ntp_service_collection, methods=["PATCH"])
