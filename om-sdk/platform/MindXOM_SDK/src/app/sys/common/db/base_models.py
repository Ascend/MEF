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
from typing import Iterable, NoReturn, Type

from sqlalchemy import Column, String
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import scoped_session

from common.utils.date_utils import DateUtils

Base = declarative_base()


class Structure(Base):
    """记录当前所有表的使用信息，升级过程用于数据平滑"""
    __tablename__ = "structure"

    table_name = Column(String, primary_key=True, unique=True, comment="数据表名")
    columns = Column(String, comment="当前版本使用的字段")
    create_time = Column(String, default=DateUtils.default_time, comment="创建时间")
    update_time = Column(String, default=DateUtils.default_time, onupdate=DateUtils.default_time, comment="更新时间")
    # 平滑维护表显式扩展三个预留字段
    reserved1 = Column(String, default=None, nullable=True, comment="预留字段1")
    reserved2 = Column(String, default=None, nullable=True, comment="预留字段2")
    reserved3 = Column(String, default=None, nullable=True, comment="预留字段3")


class SaveDefaultsMixin:

    @classmethod
    def default_instance_generator(cls) -> Iterable[Base]:
        """生成默认的Base子类对应的实例"""
        yield cls()

    @classmethod
    def save_defaults(cls, session: scoped_session) -> NoReturn:
        session.bulk_save_objects(cls.default_instance_generator())


class RebuildMixin:
    """需要在migrate前先drop后重建的表，此类表无需考虑平滑，也无需冗余字段"""
    pass


class ClearOnStartMixin:
    """启动时，需要清理历史数据的类"""

    @classmethod
    def clear_table(cls: Type[Base], session: scoped_session) -> int:
        """清理数据表"""
        return session.query(cls).delete()
