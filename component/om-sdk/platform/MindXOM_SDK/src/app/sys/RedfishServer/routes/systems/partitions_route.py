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


class PartitionsRoute(Route):

    def __init__(self, blueprint: Blueprint):
        super().__init__(blueprint)

    def add_route(self):
        # 添加磁盘分区相关的URL
        from system_service.partitions_views import rf_create_system_partitions
        from system_service.partitions_views import rf_get_system_partition_info
        from system_service.partitions_views import rf_get_system_partitions
        from system_service.partitions_views import rf_delete_system_partition
        from system_service.partitions_views import rf_mount_system_partition
        from system_service.partitions_views import rf_unmount_system_partition

        self.blueprint.add_url_rule("/Partitions", view_func=rf_get_system_partitions, methods=["GET"])
        self.blueprint.add_url_rule("/Partitions", view_func=rf_create_system_partitions, methods=["POST"])
        self.blueprint.add_url_rule("/Partitions/<partition_id>", view_func=rf_get_system_partition_info,
                                    methods=["GET"])
        self.blueprint.add_url_rule("/Partitions/<partition_id>", view_func=rf_delete_system_partition,
                                    methods=["DELETE"])
        self.blueprint.add_url_rule("/Partitions/Mount", view_func=rf_mount_system_partition, methods=["PATCH"])
        self.blueprint.add_url_rule("/Partitions/Unmount", view_func=rf_unmount_system_partition, methods=["PATCH"])
