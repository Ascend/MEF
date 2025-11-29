#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from itertools import groupby
from operator import itemgetter
from typing import Iterable


def group_by(objs: Iterable[dict], group_word) -> dict:
    # groupby是连续比较group_word，所以先排序
    opt_objs = sorted(objs, key=itemgetter(group_word))
    group_res = (groupby(opt_objs, key=itemgetter(group_word)))
    return {k: list(v) for k, v in group_res}
