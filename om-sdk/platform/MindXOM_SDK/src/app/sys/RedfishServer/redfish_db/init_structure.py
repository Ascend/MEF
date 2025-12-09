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
"""
首建模型字段持久化(需要平滑的表才记录)
    当首次新增了一个sqlite数据库ORM模型，必须将模型对应的字段名写入到此文件中的INIT_COLUMNS中；
    当版本升级修改了sqlite数据库ORM模型，不能更改INIT_COLUMNS对应表的初始字段；
"""
from typing import Dict, Iterable

from common.log.logger import run_log

INIT_COLUMNS: Dict[str, Iterable[str]] = {
    "users": (
        "id", "username_db", "pword_hash", "create_time", "pword_modify_time", "pword_wrong_times", "start_lock_time",
        "last_login_success_time", "last_login_failure_time", "account_insecure_prompt", "config_navigator_prompt",
        "log_in_time", "lock_state", "enabled", "role_id"
    ),
    "history_password": ("id", "user_id", "history_pword_hash", "pword_modify_time"),
    "edge_config": (
        "id", "default_lock_times", "lock_duration", "min_password_length", "max_password_length", "token_timeout",
        "default_expiration_days"
    ),
    "cert_manager": ("id", "name", "source", "update_time", "cert_contents", "crl_contents"),
    "net_manager": ("id", "net_mgmt_type", "node_id", "server_name", "ip", "port", "cloud_user", "cloud_pwd", "status"),
    "last_login_info": ("id", "ip"),
    "cert_info": (
        "id", "name", "subject", "issuer", "serial_num", "signature_algorithm", "signature_len", "cert_version",
        "pubkey_type", "fingerprint", "is_ca", "chain_num", "start_time", "end_time", "last_update", "next_update",
        "in_use", "key_cert_sign",
    ),
    "fd_previous_cert": ("id", "name", "source", "update_time", "cert_contents", "crl_contents"),
    "pre_cert_info": (
        "id", "name", "subject", "issuer", "serial_num", "signature_algorithm", "signature_len", "cert_version",
        "pubkey_type", "fingerprint", "is_ca", "chain_num", "start_time", "end_time", "last_update", "next_update",
        "in_use", "key_cert_sign",
    ),
}

# 尝试导入扩展，其他仓如果没有在对应的文件中实现扩展则会导入失败
try:
    from extend_interfaces import EXTEND_INIT_COLUMNS
    INIT_COLUMNS.update(**EXTEND_INIT_COLUMNS)
except ImportError as err:
    run_log.warning("Failed to import extension, ignore. %s", err)
except Exception as err:
    run_log.error("extend init columns invalid, catch %s", err.__class__.__name__)
