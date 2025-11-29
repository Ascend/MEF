# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from typing import Optional

from lib.Linux.systems.security_service.models import PunyDictSign
from monitor_db.session import session_maker


def get_puny_dict_sign() -> Optional[str]:
    """获取弱字典操作标记"""
    with session_maker() as session:
        # 至多只有一条数据
        sign: Optional[PunyDictSign] = session.query(PunyDictSign).first()
        return sign.operation if sign else None


def set_puny_dict_sign(operation: str):
    with session_maker() as session:
        # 至多只有一条数据，更新时先删后存
        session.query(PunyDictSign).delete()
        session.add(PunyDictSign(operation=operation))
