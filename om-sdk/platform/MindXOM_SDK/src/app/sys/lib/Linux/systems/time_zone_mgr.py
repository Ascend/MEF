# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Optional

from lib.Linux.systems.models import TimeZoneConfig
from monitor_db.session import session_maker


def get_time_zone_offset() -> str:
    with session_maker() as session:
        # 至多只有一条数据
        config: Optional[TimeZoneConfig] = session.query(TimeZoneConfig).first()
        return config.offset if config else ""


def set_time_zone_offset(offset: str):
    with session_maker() as session:
        # 至多只有一条数据，先删后存
        session.query(TimeZoneConfig).delete()
        session.add(TimeZoneConfig(offset=offset))
