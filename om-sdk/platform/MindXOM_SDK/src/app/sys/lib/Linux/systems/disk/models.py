# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os.path

from sqlalchemy import Column, String, Integer

from common.db.base_models import Base


class MountWhitelistPath(Base):
    __tablename__ = "mount_white_path"

    path = Column(String(256), unique=True, comment="路径")
    id = Column(Integer, primary_key=True, comment="id用做主键")

    def __init__(self, path: str):
        self.path = os.path.normpath(path)
