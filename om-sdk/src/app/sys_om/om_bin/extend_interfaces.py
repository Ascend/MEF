#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from typing import Callable, Type, Iterable, Dict

from common.db.database import DataBase

# 扩展主仓的Monitor表注册函数, 此处写个lambda，以便register模块扩展时日志正常打印
register_extend_models: Callable[[Type[DataBase]], None] = lambda database: None

# 扩展主仓的init_structure.py内容
EXTEND_INIT_COLUMNS: Dict[str, Iterable[str]] = {}

# 扩展主仓的Classmap
EXTEND_CLASS_MAP: str = "OMClassMap.json"
