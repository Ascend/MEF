# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
