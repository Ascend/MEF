#!/usr/bin/env python
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
import sys
from abc import ABC

from common.checkers.base_checker.basic_attr_checker import ExistsCheckerWithMinMax
from common.checkers.base_checker.basic_attr_checker import _ListChecker
from common.checkers.base_checker.abc_checker import AttrCheckerInterface
from common.checkers.base_checker.abc_checker import CheckResult
from common.checkers.base_checker.model_checker import BaseModelChecker


class ListChecker(ExistsCheckerWithMinMax, ABC):
    def __init__(
        self,
        attr_name: str = None,
        elem_checker: AttrCheckerInterface = None,
        min_len=0,
        max_len=sys.maxsize,
        required: bool = True,
    ):
        super().__init__(attr_name, min_len, max_len, required)
        self.elem_checker = elem_checker

    def check_dict(self, data: dict) -> CheckResult:
        ret = super().check_dict(data)
        if not ret.success:
            return ret
        value_list = self.raw_value(data)
        if value_list is None:
            return CheckResult.make_success()
        if type(value_list) not in [tuple, list]:
            msg_format = "List checker: invalid list type of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))
        ret = self.check_min_max(len(value_list))
        if not ret.success:
            return ret
        return self._check_elem_in_list(value_list)

    def _check_elem_in_list(self, value_list):
        if self.elem_checker is None:
            return CheckResult.make_success()
        for value in value_list:
            ret = self.elem_checker.check_dict(value)
            if not ret.success:
                return ret
        return CheckResult.make_success()


class UniqueListChecker(ListChecker):
    def __init__(
        self,
        attr_name: str = None,
        elem_checker: AttrCheckerInterface = None,
        min_len=0,
        max_len=sys.maxsize,
        required: bool = True,
    ):
        super().__init__(attr_name, elem_checker, min_len, max_len, required)

    def _check_elem_in_list(self, value_list):
        ret = super()._check_elem_in_list(value_list)
        if not ret.success:
            return ret

        if len(set(value_list)) != len(value_list):
            return CheckResult.make_failed(f"UniqueListChecker: not unique in {self.name()}")

        return CheckResult.make_success()


class AndChecker(_ListChecker):
    def __init__(self, *checkers):
        super().__init__("AndChecker", *checkers, required=False)

    def check_dict(self, data: dict) -> CheckResult:
        for checker in self.checkers:
            if isinstance(checker, BaseModelChecker):
                ret = checker.check(data.get(checker.name()))
            else:
                ret = checker.check(data)
            if not ret.success:
                return ret
        return CheckResult.make_success()


class OrChecker(_ListChecker):
    def __init__(self, *checkers):
        super().__init__("OrChecker", *checkers, required=False)

    def check_dict(self, data: dict) -> CheckResult:
        failed_msg = []
        checker_description = None
        for checker in self.checkers:
            if isinstance(checker, BaseModelChecker):
                ret = checker.check(data.get(checker.name()))
            else:
                ret = checker.check(data)
            if ret.success:
                return ret
            if checker_description is None:
                checker_description = ret.check_description
            failed_msg.append(ret.reason)
        return CheckResult.make_failed("\n".join(failed_msg), checker_description)
