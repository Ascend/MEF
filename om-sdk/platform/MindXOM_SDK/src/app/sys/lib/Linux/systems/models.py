# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from sqlalchemy import Column, String, Integer

from common.db.base_models import Base


class TimeZoneConfig(Base):
    __tablename__ = "time_zone_config"

    id = Column(Integer, primary_key=True, comment="至多只有一条配置")
    offset = Column(String, comment="系统时间时区")
