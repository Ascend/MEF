# -*-coding:utf-8-*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import abc
from abc import ABC
from typing import Optional, Union, Any


class CheckResult:
    def __init__(self, success: bool, reason: str, checker_description: Any = None, err_code: int = None):
        self.success: bool = success
        self.reason: str = reason
        self.check_description: Any = checker_description
        self.err_code: int = err_code

    def __str__(self):
        return "%s" % self.__dict__

    def __repr__(self):
        return self.__str__()

    def __bool__(self):
        return self.success

    @classmethod
    def make_success(cls):
        return cls(True, "")

    @classmethod
    def make_failed(cls, msg: str, checker_description=None, err_code=None):
        return cls(False, msg, checker_description, err_code)


class AttrCheckerInterface:
    @abc.abstractmethod
    def name(self) -> str:
        pass

    @abc.abstractmethod
    def required(self) -> bool:
        pass

    @abc.abstractmethod
    def set_name(self, attr_name):
        pass

    @abc.abstractmethod
    def check(self, data: Optional[Union[object, dict, int, str, float, bool, bytes, list, tuple]]) -> CheckResult:
        pass

    @abc.abstractmethod
    def raw_value(self, data):
        pass

    @abc.abstractmethod
    def check_dict(self, data: Optional[Union[dict, int, str, float, bool, bytes, list, tuple]]) -> CheckResult:
        pass


class AttrCheckerBase(AttrCheckerInterface):
    def __init__(self, attr_name: str):
        self.attr_name = attr_name

    def name(self) -> str:
        return self.attr_name

    @abc.abstractmethod
    def required(self) -> bool:
        pass

    def set_name(self, attr_name):
        self.attr_name = attr_name

    def check(self, data: Optional[Union[object, dict, int, str, float, bool, bytes, list, tuple]]) -> CheckResult:
        if isinstance(data, dict):
            return self.check_dict(data)
        elif isinstance(data, (int, str, float, bool, bytes, list, tuple)) or data is None:
            return self.check_dict(data)
        else:
            return self.check_dict(data.__dict__)

    def raw_value(self, data):
        if self.name() is None:
            return data
        return data.get(self.name())

    @abc.abstractmethod
    def check_dict(self, data: Optional[Union[dict, int, str, float, bool, bytes, list, tuple]]) -> CheckResult:
        pass


class DefaultAttrCheck(AttrCheckerBase, ABC):
    def __init__(self, attr_name: str = None):
        super().__init__(attr_name)

    def required(self) -> bool:
        return False

    def check_dict(self, data: Optional[Union[dict, int, str, float, bool, bytes, list, tuple]]) -> CheckResult:
        return CheckResult.make_success()
