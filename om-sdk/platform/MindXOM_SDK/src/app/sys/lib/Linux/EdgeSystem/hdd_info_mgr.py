# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Optional

from sqlalchemy import func

from lib.Linux.EdgeSystem.models import HddInfo
from monitor_db.session import session_maker


def get_hdd_serial_number() -> str:
    """获取存在表中的序列号，至多只有一条数据"""
    with session_maker() as session:
        hdd: Optional[HddInfo] = session.query(HddInfo).first()
        return hdd.serial_number if hdd else ""


def get_hdd_info_count() -> int:
    """获取hdd条数"""
    with session_maker() as session:
        return session.query(func.count(HddInfo.serial_number)).scalar()


def clear_or_save_hdd_info(serial_number: Optional[str] = None):
    with session_maker() as session:
        session.query(HddInfo).delete()
        if serial_number:
            session.add(HddInfo(serial_number=serial_number))
