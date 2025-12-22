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

from itertools import groupby
from operator import itemgetter
from typing import Iterable


def group_by(objs: Iterable[dict], group_word) -> dict:
    # groupby是连续比较group_word，所以先排序
    opt_objs = sorted(objs, key=itemgetter(group_word))
    group_res = (groupby(opt_objs, key=itemgetter(group_word)))
    return {k: list(v) for k, v in group_res}
