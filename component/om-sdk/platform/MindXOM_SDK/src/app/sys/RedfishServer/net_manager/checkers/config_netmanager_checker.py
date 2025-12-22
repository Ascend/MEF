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

import sys

from common.checkers import PasswordComplexityChecker
from common.checkers.param_checker import FdIpChecker
from common.checkers import AndChecker
from common.checkers import BoolChecker
from common.checkers import IntegerChecker
from common.checkers import ModelChecker
from common.checkers import OrChecker
from common.checkers import PortChecker
from common.checkers import RegexStringChecker
from common.checkers import StringChoicesChecker
from common.checkers import StringEmptyChecker
from common.checkers import StringExcludeWordsChecker
from common.checkers import StringLengthChecker
from common.checkers import UUID4Checker


class ConfigNetManagerChecker(ModelChecker):
    class Meta:
        fields = (OrChecker(
            StringChoicesChecker("manager_type", ("Web",)),
            AndChecker(
                StringChoicesChecker("manager_type", ("FusionDirector",)),
                FdIpChecker("net_ip"),
                RegexStringChecker("net_account", "^[a-zA-Z0-9-_]{1,256}$"),
                StringLengthChecker("net_password", min_len=1, max_len=256),
                BoolChecker("test"),
                RegexStringChecker("server_name", "^[A-Za-z0-9-.]{0,64}$"),
                PortChecker("port", required=False),
                UUID4Checker("node_id", required=False),
            )))


class FdConfigChecker(ModelChecker):
    class Meta:
        fields = (
            OrChecker(
                AndChecker(
                    StringChoicesChecker("dev_mgmt_type", ("Web",)),
                    RegexStringChecker("node_id", "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
                ),
                AndChecker(
                    StringChoicesChecker("dev_mgmt_type", ("AtlasEdge",)),
                    FdIpChecker("server_ip"),
                    PortChecker("server_port"),
                    RegexStringChecker("cloud_user", "^[a-zA-Z0-9-_]{1,256}$"),
                    RegexStringChecker("cloud_pwd", "^.{1,256}$"),
                    OrChecker(
                        StringEmptyChecker("crl_path"),
                        AndChecker(
                            RegexStringChecker("crl_path", "^[0-9a-zA-Z_./-]{1,256}$"),
                            StringExcludeWordsChecker("crl_path", words=("..",)),
                        )
                    ),
                    RegexStringChecker("cert_path", "^[0-9a-zA-Z_./-]{1,256}$"),
                    StringExcludeWordsChecker("cert_path", words=("..",)),
                    RegexStringChecker("node_id", "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
                    StringChoicesChecker("status", ("", "wait", "connecting", "ready")),
                ),
            ),
        )


class FdConfigNetManagerInfoChecker(ModelChecker):
    """一机一密参数校验espmanager/netmanager"""

    class Meta:
        fields = (
            RegexStringChecker("account", "^[a-zA-Z0-9-_]{1,256}$"),
            PasswordComplexityChecker("password", min_len=8, max_len=20),
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
            IntegerChecker("timestamp", min_value=0, max_value=sys.maxsize)
        )


class FdMsgRouteChecker(ModelChecker):
    """Fd下发消息route内容校验"""

    class Meta:
        fields = (
            StringChoicesChecker("group", choices=("resource", "twin", "hardware", "function", "user")),
            StringChoicesChecker("operation", choices=("insert", "update", "delete", "query")),
            RegexStringChecker("resource", "^[a-zA-Z0-9-_/]{1,256}$"),
            RegexStringChecker("source", "^[a-zA-Z0-9]{1,256}$"),
        )


class FdMsgChecker(ModelChecker):
    """FD下发消息格式校验"""

    class Meta:
        fields = (
            FdMsgHeaderChecker("header"),
            FdMsgRouteChecker("route", required=True),
        )
