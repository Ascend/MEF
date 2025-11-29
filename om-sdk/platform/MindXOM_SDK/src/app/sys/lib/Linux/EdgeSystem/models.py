# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from sqlalchemy import Column, String

from common.db.base_models import Base


class HddInfo(Base):
    __tablename__ = "hdd_info"

    serial_number = Column(String, primary_key=True, unique=True, comment="hdd序列号")
