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


class LteRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加LTE相关的路由
        from system_service.lte_views import rf_get_system_lte
        from system_service.lte_views import rf_patch_system_lte_status_info
        from system_service.lte_views import rf_get_system_lte_config_info
        from system_service.lte_views import rf_patch_system_lte_config_info
        from system_service.lte_views import rf_get_system_lte_status_info

        self.blueprint.add_url_rule("/LTE", view_func=rf_get_system_lte, methods=["GET"])
        self.blueprint.add_url_rule("/LTE/StatusInfo", view_func=rf_get_system_lte_status_info, methods=["GET"])
        self.blueprint.add_url_rule("/LTE/StatusInfo", view_func=rf_patch_system_lte_status_info, methods=["PATCH"])
        self.blueprint.add_url_rule("/LTE/ConfigInfo", view_func=rf_get_system_lte_config_info, methods=["GET"])
        self.blueprint.add_url_rule("/LTE/ConfigInfo", view_func=rf_patch_system_lte_config_info, methods=["PATCH"])
