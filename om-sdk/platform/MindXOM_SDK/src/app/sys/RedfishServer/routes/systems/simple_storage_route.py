# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from flask import Blueprint

from routes.route import Route


class SimpleStorageRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加简单存储相关的URL
        from system_service.simple_storage_views import rf_get_system_simple_storages_collection
        from system_service.simple_storage_views import rf_get_system_storage_info

        self.blueprint.add_url_rule("/SimpleStorages", view_func=rf_get_system_simple_storages_collection,
                                    methods=["GET"])
        self.blueprint.add_url_rule("/SimpleStorages/<storage_id>", view_func=rf_get_system_storage_info,
                                    methods=["GET"])
