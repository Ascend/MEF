# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
"""
首建模型字段持久化(需要平滑的表才记录)
    当首次新增了一个sqlite数据库ORM模型，必须将模型对应的字段名写入到此文件中的INIT_COLUMNS中；
    当版本升级修改了sqlite数据库ORM模型，不能更改INIT_COLUMNS对应表的初始字段；
"""
from typing import Dict, Iterable

from common.log.logger import run_log

INIT_COLUMNS: Dict[str, Iterable[str]] = {
    "time_zone_config": ("id", "offset"),
    "puny_dict_sign": ("id", "operation"),
    "mount_white_path": ("path", "id"),
    "hdd_info": ("serial_number",),
    "nfs_cfg": ("id", "server_ip", "server_dir", "local_dir", "fs_type"),
    "login_rules": ("id", "enable", "start_time", "end_time", "ip_addr", "mac_addr"),
    "net_config": ("id", "name", "tag", "ipv4"),
}

# 尝试导入扩展，其他仓如果没有在对应的文件中实现扩展则会导入失败
try:
    from bin.extend_interfaces import EXTEND_INIT_COLUMNS
    INIT_COLUMNS.update(**EXTEND_INIT_COLUMNS)
except ImportError as err:
    run_log.warning("Failed to import extension, ignore. %s", err)
except Exception as err:
    run_log.error("extend init columns invalid, catch %s", err.__class__.__name__)
