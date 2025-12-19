# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from routes.account.account_route import AccountRoute
from routes.session.session_route import SessionRoute
from routes.systems.actions_route import ActionRoute
from routes.systems.alarm_route import AlarmRoute
from routes.systems.extend_devices_route import ExtendDevicesRoute
from routes.systems.log_services_route import LogServicesRoute
from routes.systems.lte_route import LteRoute
from routes.systems.memory_route import MemoryRoute
from routes.systems.modules_route import ModulesRoute
from routes.systems.nfs_route import NfsRoute
from routes.systems.nic_route import NicRoute
from routes.systems.ntp_route import NtpRoute
from routes.systems.partitions_route import PartitionsRoute
from routes.systems.processors_route import ProcessorsRoute
from routes.systems.security_service_route import SecurityServiceRoute
from routes.systems.simple_storage_route import SimpleStorageRoute
from routes.systems.system_route import SystemRoute
from routes.upgrade.upgrade_route import UpgradeRoute

# 目前总的组件包括
# account_manager(用户管理), session_manager(会话管理),
# system(系统资源), actions(复位、恢复出厂),
# alarm(告警),modules(devm),
# extend_devices(外设), processors(处理器),
# ntp, memory(内存),
# log_services(日志), lte, nfs,
# simple_storage(存储), partitions(分区),
# nic, security_service(安全服务), upgrade(升级)
SUPPORT_COMPONENTS = {
    "systems": [
        ModulesRoute, SystemRoute, ActionRoute, AlarmRoute, ExtendDevicesRoute,
        ProcessorsRoute, NtpRoute, MemoryRoute, LogServicesRoute, LteRoute, NfsRoute,
        SimpleStorageRoute, PartitionsRoute, NicRoute, SecurityServiceRoute
    ],
    "session": [SessionRoute, ],
    "account": [AccountRoute, ],
    "upgrade": [UpgradeRoute, ]
}
