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
import os
from typing import Iterable

from sqlalchemy import func

from lib.Linux.systems.disk.models import MountWhitelistPath
from monitor_db.session import session_maker

MAX_WHITE_PATH_CNT = 64


def get_whitelist_path_count():
    """获取挂载白名单路径数量"""
    with session_maker() as session:
        return session.query(func.count(MountWhitelistPath.path)).scalar()


def query_whitelist_path(limit=MAX_WHITE_PATH_CNT) -> Iterable[MountWhitelistPath]:
    """获取挂载白名单列表"""
    with session_maker() as session:
        for white in session.query(MountWhitelistPath).limit(limit).all():
            yield white


def path_in_whitelist(path: str) -> bool:
    with session_maker() as session:
        return bool(session.query(MountWhitelistPath).filter_by(path=os.path.normpath(path)).first())


def add_whitelist_paths(*paths: str):
    """保存白名单路径"""
    existed = {p.path for p in query_whitelist_path()}
    white_paths = (MountWhitelistPath(p) for p in paths)
    with session_maker() as session:
        session.bulk_save_objects((p for p in white_paths if p.path not in existed))


def add_whitelist_path(path: str):
    """保存单个挂载白名单"""
    with session_maker() as session:
        session.add(MountWhitelistPath(path=path))


def delete_mount_white_path(path: str) -> int:
    """删除白名单路径，入参校验由调用处保证"""
    with session_maker() as session:
        return session.query(MountWhitelistPath).filter_by(path=os.path.normpath(path)).delete()
