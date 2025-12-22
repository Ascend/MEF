#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from typing import Callable, Type, Iterable, Dict

from common.db.database import DataBase

# 扩展主仓的Monitor表注册函数, 此处写个lambda，以便register模块扩展时日志正常打印
register_extend_models: Callable[[Type[DataBase]], None] = lambda database: None

# 扩展主仓的init_structure.py内容
EXTEND_INIT_COLUMNS: Dict[str, Iterable[str]] = {}

# 扩展主仓的Classmap
EXTEND_CLASS_MAP: str = "OMClassMap.json"
