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
import os
from typing import Iterable

from sqlalchemy import func

from lib.Linux.systems.nfs.models import NfsCfg
from monitor_db.session import session_maker

MAX_WHITE_PATH_CNT = 64
NFS_MAX_CFG_NUM = 32


def query_nfs_configs(limit=NFS_MAX_CFG_NUM, expunge=False) -> Iterable[NfsCfg]:
    """获取NFS的所有配置"""
    with session_maker() as session:
        for nfs in session.query(NfsCfg).limit(limit).all():
            if expunge:
                session.expunge(nfs)
            yield nfs


def get_nfs_config_count() -> int:
    with session_maker() as session:
        return session.query(func.count(NfsCfg.local_dir)).scalar()


def del_nfs_config_by_mount_path(path: str) -> int:
    with session_maker() as session:
        return session.query(NfsCfg).filter_by(local_dir=os.path.normpath(path)).delete()


def mount_path_already_exists(path: str) -> bool:
    with session_maker() as session:
        return bool(session.query(NfsCfg).filter_by(local_dir=os.path.normpath(path)).first())


def save_nfs_config(nfs: NfsCfg):
    with session_maker() as session:
        session.add(nfs)
