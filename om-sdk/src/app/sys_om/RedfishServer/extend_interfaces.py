#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from typing import Callable, Type, Iterable, Dict

from common.db.database import DataBase
from common.kmc_lib.kmc_updater import MultiKmcUpdater
from om_kmc_update.om_kmc_updater import extend_om_updater_and_adapters
from om_redfish_db.init_structure import INIT_COLUMNS
from om_redfish_db.register import register_extend_om_models
from om_redfish_service.om_redfish_public import ExtendSchemaCollection
from redfish_extend_funcs import start_extend_funcs


# 新增组件信息查询函数，用于静态信息上报FD
EXTEND_COMPONENTS_INFO_FUNC_PATHS = ["om_fd_msg_process.midware_tasks.get_digital_warranty"]

# 扩展主仓的Redfish表注册函数
register_extend_models: Callable[[Type[DataBase]], None] = register_extend_om_models

# 扩展主仓的init_structure.py内容
EXTEND_INIT_COLUMNS: Dict[str, Iterable[str]] = INIT_COLUMNS

# 启动Redfish扩展功能
register_extend_func: Callable[[], None] = start_extend_funcs

# 扩展主仓的schema文件
EXTEND_REDFISH_SCHEMA = ExtendSchemaCollection.get_extend_schema_list()

# 扩展主仓的Kmc更新注册函数
extend_updater_and_adapters: Callable[[Type[MultiKmcUpdater]], None] = extend_om_updater_and_adapters
