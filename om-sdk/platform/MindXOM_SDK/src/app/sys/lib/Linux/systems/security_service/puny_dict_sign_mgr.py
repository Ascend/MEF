# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
