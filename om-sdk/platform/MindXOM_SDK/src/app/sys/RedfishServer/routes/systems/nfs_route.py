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


class NfsRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加Nfs相关的路由
        from system_service.nfs_views import rf_get_system_nfs_manage
        from system_service.nfs_views import rf_mount_system_nfs_manage
        from system_service.nfs_views import rf_unmount_system_nfs_manage

        self.blueprint.add_url_rule("/NfsManage", view_func=rf_get_system_nfs_manage, methods=["GET"])
        self.blueprint.add_url_rule("/NfsManage/Actions/NfsManage.Mount", view_func=rf_mount_system_nfs_manage,
                                    methods=["POST"])
        self.blueprint.add_url_rule("/NfsManage/Actions/NfsManage.Unmount",
                                    view_func=rf_unmount_system_nfs_manage, methods=["POST"])
