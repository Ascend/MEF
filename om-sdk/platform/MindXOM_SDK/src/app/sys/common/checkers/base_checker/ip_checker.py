# -*-coding:utf-8-*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import re
import socket
from abc import ABC

from common.checkers.base_checker.basic_attr_checker import ExistsChecker, IntegerChecker
from common.checkers.base_checker.abc_checker import CheckResult
from common.checkers.base_checker.list_checker import OrChecker


class _IPChecker(ExistsChecker, ABC):
    def __init__(self, attr_name: str, ip_type: int, required: bool = True):
        super().__init__(attr_name, required)
        self.ip_type = ip_type

    @staticmethod
    def domain_check(domain):
        if not isinstance(domain, str):
            return False

        pattern = r"^[A-Za-z0-9\-\.]{1,253}$"
        return True if re.match(pattern, domain) else False

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if not isinstance(value, str):
            msg_format = "IP checker: invalid value type '{}' of {}"
            return CheckResult.make_failed(msg_format.format(type(value), self.name()))

        if value == "0.0.0.0":
            msg_format = "IP checker: invalid value of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))

        return self._check_ip(value)

    def _check_ip(self, value) -> CheckResult:
        try:
            socket.inet_pton(self.ip_type, value)
            return CheckResult.make_success()
        except socket.error:
            return CheckResult.make_failed("IP checker: invalid ip address")


class IPV4Checker(_IPChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, socket.AF_INET, required)


class IPV6Checker(_IPChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, socket.AF_INET6, required)


class IPChecker(OrChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(IPV4Checker(attr_name, required), IPV6Checker(attr_name, required))


class PortChecker(IntegerChecker):
    MIN_PORT_NUM = 1
    MAX_PORT_NUM = 65535

    def __init__(self,
                 attr_name: str = None,
                 min_value: int = MIN_PORT_NUM,
                 max_value: int = MAX_PORT_NUM,
                 required: bool = True,
                 restrict: bool = False):
        super().__init__(attr_name, min_value, max_value, required, restrict)
