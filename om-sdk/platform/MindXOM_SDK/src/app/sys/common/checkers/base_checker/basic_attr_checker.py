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
import re
import sys
from abc import ABC
from typing import Tuple, Union

from common.checkers.base_checker.abc_checker import AttrCheckerBase, AttrCheckerInterface
from common.checkers.base_checker.abc_checker import CheckResult


def str_is_float(data: str) -> bool:
    if data.isdigit():
        return True
    try:
        float(data)
        return True
    except ValueError:
        return False


class ExistsChecker(AttrCheckerBase, ABC):
    def __init__(self, attr_name=None, required: bool = True):
        super().__init__(attr_name)
        self._required = required

    def required(self) -> bool:
        return self._required

    def check_dict(self, data: dict) -> CheckResult:
        if not self.required():
            return CheckResult.make_success()
        if data is None:
            return CheckResult.make_failed("Exists checker: input is null while check {}".format(self.name()))
        value = self.raw_value(data)
        if value is not None:
            return CheckResult.make_success()
        return CheckResult.make_failed("Exists checker: {} not exists".format(self.name()))


class ExistsCheckerWithMinMax(ExistsChecker, ABC):
    def __init__(self, attr_name: str, min_value, max_value, required: bool = True):
        super().__init__(attr_name, required)
        self.min_value = min_value
        self.max_value = max_value

    def check_min_max(self, value):
        msg_format = "Exists checker with min max: the value of {} is out of range [{}, {}]"
        if value > self.max_value or value < self.min_value:
            return CheckResult.make_failed(msg_format.format(self.name(), self.min_value, self.max_value))
        return CheckResult.make_success()


class IntegerChecker(ExistsCheckerWithMinMax):
    def __init__(self,
                 attr_name: str = None,
                 min_value: int = -sys.maxsize - 1,
                 max_value: int = sys.maxsize,
                 required: bool = True,
                 restrict: bool = False,
                 max_value_len: int = 128):
        super().__init__(attr_name, min_value, max_value, required)
        self.restrict = restrict
        self.max_value_len = max_value_len

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if self.restrict and not isinstance(value, int):
            return CheckResult.make_failed("Integer checker: {} is not a integer digit".format(self.name()))
        if len("%s" % value) > self.max_value_len:
            return CheckResult.make_failed("Integer checker: The length of {} exceeds the up limit({})".format(
                self.name(), self.max_value_len))
        if not ("%s" % value).isdigit():
            return CheckResult.make_failed("Integer checker: {} is not a digit".format(self.name()))
        return self.check_min_max(int(value))


class BoolChecker(ExistsChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, required)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()

        if not isinstance(value, bool):
            return CheckResult.make_failed("Bool checker: {} is not bool type".format(self.name()))

        return CheckResult.make_success()


class BoolEqualChecker(ExistsChecker):
    def __init__(self, attr_name: str = None, equal_value: bool = True, required: bool = True):
        super().__init__(attr_name, required)
        self.equal_value = equal_value

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if value != self.equal_value:
            return CheckResult.make_failed(f"Bool checker: {self.name()} is not equal to {self.equal_value}")

        return CheckResult.make_success()


class FloatChecker(ExistsCheckerWithMinMax, ABC):
    def __init__(
        self,
        attr_name: str = None,
        min_value: float = sys.float_info.min,
        max_value: float = sys.float_info.max,
        required: bool = True,
    ):
        super().__init__(attr_name, min_value, max_value, required)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if not str_is_float(f"{value}"):
            return CheckResult.make_failed("Float checker: {} is not a float".format(self.name()))
        return self.check_min_max(float(value))


class StringChoicesChecker(ExistsChecker, ABC):
    def __init__(self, attr_name: str, choices: Tuple[Union[str, bool], ...] = (), required: bool = True):
        super().__init__(attr_name, required)
        self.choices = choices

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if value not in self.choices:
            msg_format = "String choices checker: invalid parameter value of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))

        return CheckResult.make_success()


class StringRejectChoicesChecker(ExistsChecker, ABC):
    def __init__(self, attr_name: str, reject_choices: Tuple[Union[str, bool], ...] = (), required: bool = True):
        super().__init__(attr_name, required)
        self.reject_choices = reject_choices

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()

        if value in self.reject_choices:
            msg_format = "String reject choices checker: invalid parameter value of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))

        return CheckResult.make_success()


class StringEmptyChecker(StringChoicesChecker, ABC):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, ("",), required)


class StringLengthChecker(ExistsChecker, ABC):
    def __init__(
        self,
        attr_name=None,
        min_len: int = 0,
        max_len: int = 32,
        required: bool = True,
    ):
        super().__init__(attr_name, required)
        self.min_len: int = min_len
        self.max_len: int = max_len

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if not isinstance(value, str):
            msg_format = "String length checker: invalid value type '{}' of {}"
            return CheckResult.make_failed(msg_format.format(type(value), self.name()))
        if len(value) < self.min_len or len(value) > self.max_len:
            msg_format = "String length checker: invalid length of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))
        return CheckResult.make_success()


class BinaryChecker(AttrCheckerBase):
    def __init__(self, lhs_name: str, rhs_name: str, compare_fun):
        super().__init__("{}<>{}".format(lhs_name, rhs_name))
        self._lhs_name = lhs_name
        self._rhs_name = rhs_name
        self._compare_fun = compare_fun

    def required(self) -> bool:
        return True

    def check_dict(self, data: dict) -> CheckResult:
        lhs_value = data.get(self._lhs_name)
        rhs_value = data.get(self._rhs_name)
        if not self._compare_fun(lhs_value, rhs_value):
            return CheckResult.make_failed("Binary checker {} return False".format(self.name()))
        return CheckResult.make_success()


class IntegerBinaryChecker(BinaryChecker):
    def __init__(self, lhs_name: str, rhs_name: str, compare_fun):
        super().__init__(lhs_name, rhs_name, compare_fun)

    def check_dict(self, data: dict) -> CheckResult:
        lhs_value = data.get(self._lhs_name)
        rhs_value = data.get(self._rhs_name)
        if not ("%s" % lhs_value).isdigit():
            return CheckResult.make_failed("Integer binary checker: lhs: {} is not a digit".format(self._lhs_name))
        if not ("%s" % rhs_value).isdigit():
            return CheckResult.make_failed("Integer binary checker: rhs: {} is not a digit".format(self._rhs_name))
        if not self._compare_fun(int(lhs_value), int(rhs_value)):
            return CheckResult.make_failed("Integer binary checker {} return False".format(self.name()))
        return CheckResult.make_success()


class FloatBinaryChecker(BinaryChecker):
    def __init__(self, lhs_name: str, rhs_name: str, compare_fun):
        super().__init__(lhs_name, rhs_name, compare_fun)

    def check_dict(self, data: dict) -> CheckResult:
        lhs_value = data.get(self._lhs_name)
        rhs_value = data.get(self._rhs_name)
        if not str_is_float("%s" % lhs_value):
            return CheckResult.make_failed("Float binary checker: lhs: {} is not a float".format(self._lhs_name))
        if not str_is_float("%s" % rhs_value):
            return CheckResult.make_failed("Float binary checker: rhs: {} is not a float".format(self._rhs_name))
        if not self._compare_fun(float(lhs_value), float(rhs_value)):
            return CheckResult.make_failed("Float binary checker {} return False".format(self.name()))
        return CheckResult.make_success()


class StringBinaryChecker(BinaryChecker):
    def __init__(self, lhs_name: str, rhs_name: str, compare_fun):
        super().__init__(lhs_name, rhs_name, compare_fun)

    def check_dict(self, data: dict) -> CheckResult:
        lhs_value = data.get(self._lhs_name)
        rhs_value = data.get(self._rhs_name)
        if not isinstance(lhs_value, str):
            return CheckResult.make_failed("String binary checker: lhs: {} is not a float".format(self._lhs_name))
        if not isinstance(rhs_value, str):
            return CheckResult.make_failed("String binary checker: rhs: {} is not a float".format(self._rhs_name))
        if not self._compare_fun(lhs_value, rhs_value):
            return CheckResult.make_failed("String binary checker {} return False".format(self.name()))
        return CheckResult.make_success()


class RegexStringChecker(StringLengthChecker, ABC):
    def __init__(
        self,
        attr_name: str = None,
        match_str: str = "",
        min_len: int = 0,
        max_len: int = sys.maxsize,
        required: bool = True,
    ):
        super().__init__(attr_name, min_len, max_len, required)
        self.match_str = match_str

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if not isinstance(value, str):
            msg_format = "Regex string checker: invalid value type '{}' of {}"
            return CheckResult.make_failed(msg_format.format(type(value), self.name()))
        find_iter = re.fullmatch(self.match_str, value)
        if find_iter is None or find_iter.group(0) != value:
            msg_format = "Regex string checker: invalid format of {}"
            return CheckResult.make_failed(msg_format.format(self.name()))
        return CheckResult.make_success()


class _ListChecker(ExistsChecker, ABC):
    def __init__(self, attr_name, *checkers: AttrCheckerInterface, required: bool = True):
        super().__init__(attr_name, required)
        self.checkers: Tuple[AttrCheckerInterface] = checkers


def is_in(checkers: Tuple[AttrCheckerInterface], field_name: str) -> bool:
    for checker in checkers:
        if isinstance(checker, _ListChecker):
            if is_in(checker.checkers, field_name):
                return True
        else:
            if checker.name() == field_name:
                return True

    return False
