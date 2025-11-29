# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from sqlalchemy import Column, String, Integer

from common.db.base_models import Base


class PunyDictSign(Base):
    __tablename__ = "puny_dict_sign"

    id = Column(Integer, primary_key=True, comment="id，至多只有一条记录")
    operation = Column(String, comment="操作类型")


class LoginRules(Base):
    __tablename__ = "login_rules"

    id = Column(Integer, primary_key=True)
    enable = Column(String(32), default=None)
    start_time = Column(String(32), default=None)
    end_time = Column(String(32), default=None)
    ip_addr = Column(String(32), default=None)
    mac_addr = Column(String(32), default=None)

    def __hash__(self):
        return hash(f"{self.start_time}:{self.end_time}:{self.ip_addr}:{self.mac_addr}")

    def __eq__(self, other):
        if not isinstance(other, self.__class__):
            return False

        return self.start_time == other.start_time and self.end_time == other.end_time and \
               self.ip_addr == other.ip_addr and self.mac_addr == other.ip_addr and \
               self.enable == other.enable

    @classmethod
    def to_obj(cls, data: dict):
        """将data转成LoginRules对象"""
        return cls(
            enable=data.get("enable"),
            start_time=data.get("start_time"),
            end_time=data.get("end_time"),
            ip_addr=data.get("ip_addr"),
            mac_addr=data.get("mac_addr"),
        )

    def to_dict(self):
        return {
            "enable": self.enable,
            "start_time": self.start_time,
            "end_time": self.end_time,
            "ip_addr": self.ip_addr,
            "mac_addr": self.mac_addr,
        }

    def is_part_same(self, other):
        if not isinstance(other, self.__class__):
            return False

        return self.start_time == other.start_time or self.end_time == other.end_time or \
               self.ip_addr == other.ip_addr or self.mac_addr == other.mac_addr
