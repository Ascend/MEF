# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os.path

from sqlalchemy import Column, String, Integer

from common.db.base_models import Base


class MountWhitelistPath(Base):
    __tablename__ = "mount_white_path"

    path = Column(String(256), unique=True, comment="路径")
    id = Column(Integer, primary_key=True, comment="id用做主键")

    def __init__(self, path: str):
        self.path = os.path.normpath(path)
