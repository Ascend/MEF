# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from common.checkers import PasswordComplexityChecker
from common.checkers.param_checker import FdIpChecker
from common.checkers import BoolChecker
from common.checkers import IntegerChecker
from common.checkers import ModelChecker
from common.checkers import OrChecker
from common.checkers import RegexStringChecker
from common.checkers import StringBinaryChecker
from common.checkers import StringChoicesChecker
from common.checkers import StringEmptyChecker
from common.checkers import UUID4Checker


def judge_pwd_by_account(password: str, account: str):
    if password in (account, account[:: -1]):
        return False
    return True


class FdConfigNetManagerInfoChecker(ModelChecker):
    """一机一密参数校验espmanager/netmanager"""

    class Meta:
        fields = (
            RegexStringChecker("account", "^[a-zA-Z0-9-_]{1,256}$"),
            PasswordComplexityChecker("password", min_len=8, max_len=32),
            StringBinaryChecker("password", "account", compare_fun=judge_pwd_by_account),
            OrChecker(
                StringEmptyChecker("address"),
                FdIpChecker("address")
            ),
            BoolChecker("test")
        )


class FdMsgHeaderChecker(ModelChecker):
    """Fd下发消息header内容校验"""

    class Meta:
        fields = (
            UUID4Checker("msg_id"),
            OrChecker(
                StringEmptyChecker("parent_msg_id"),
                UUID4Checker("parent_msg_id"),
            ),
            BoolChecker("sync"),
            IntegerChecker("timestamp", min_value=0)
        )


class FdMsgRouteChecker(ModelChecker):
    """Fd下发消息route内容校验"""

    class Meta:
        fields = (
            StringChoicesChecker("group", choices=("resource", "twin", "hardware", "function", "user")),
            StringChoicesChecker("operation", choices=("insert", "update", "delete", "query", "restart")),
            RegexStringChecker("resource", "^[a-zA-Z0-9-_/]{1,256}$"),
            RegexStringChecker("source", "^[a-zA-Z0-9]{1,256}$"),
        )


class FdMsgChecker(ModelChecker):
    """FD下发消息格式校验"""

    class Meta:
        fields = (
            FdMsgHeaderChecker("header"),
            FdMsgRouteChecker("route"),
        )
