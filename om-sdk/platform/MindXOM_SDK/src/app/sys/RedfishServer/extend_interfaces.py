# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Callable, Type, Dict, Iterable

from common.db.database import DataBase
from common.kmc_lib.kmc_updater import MultiKmcUpdater
from common.utils.result_base import Result

# 新增组件信息查询函数，用于静态信息上报FD
EXTEND_COMPONENTS_INFO_FUNC_PATHS = []

# 扩展主仓的Redfish表注册函数，此处写个lambda，以便register模块扩展时日志正常打印
register_extend_models: Callable[[Type[DataBase]], None] = lambda database: None

# 扩展主仓的init_structure.py内容
EXTEND_INIT_COLUMNS: Dict[str, Iterable[str]] = {}

# 启动Redfish扩展功能
register_extend_func: Callable[[], None] = lambda: None

# 扩展主仓的schema文件
EXTEND_REDFISH_SCHEMA = []

# 扩展主仓的Kmc更新注册函数
extend_updater_and_adapters: Callable[[Type[MultiKmcUpdater]], None] = lambda updater: None

# 扩展证书管理功能
extend_certs_manager_funcs: Callable[[str], Result] = \
    lambda action: Result(False, err_msg="Extend manage func not exists.")
