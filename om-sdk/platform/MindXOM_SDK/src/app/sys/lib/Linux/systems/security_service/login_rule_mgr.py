#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from typing import NoReturn, List, Set, Dict

from monitor_db.session import session_maker
from lib.Linux.systems.security_service.models import LoginRules


class LoginRuleManager:
    MAX_LOGIN_RULE_NUM = 30
    model = LoginRules

    def get_all(self) -> List[Dict]:
        """
        获取所有数据对象
        :return: List["model"]: 数据对象列表，[]: 未找到数据对象
        """
        with session_maker() as session:
            return [obj.to_dict() for obj in session.query(self.model).limit(self.MAX_LOGIN_RULE_NUM + 1).all()]

    def over_write_database(self, objs: Set[LoginRules]) -> NoReturn:
        with session_maker() as session:
            session.query(self.model).delete()
            session.bulk_save_objects(objs)
