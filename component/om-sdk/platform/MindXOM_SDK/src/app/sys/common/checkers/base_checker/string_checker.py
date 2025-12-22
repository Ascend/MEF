#!/usr/bin/env python
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
from abc import ABC

from common.checkers.base_checker.basic_attr_checker import ExistsChecker
from common.checkers.base_checker.basic_attr_checker import RegexStringChecker
from common.checkers.base_checker.abc_checker import CheckResult


class UUID4Checker(RegexStringChecker, ABC):
    def __init__(self, attr_name: str, required: bool = True):
        super().__init__(attr_name,
                         match_str="^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
                         required=required)


class StringExcludeWordsChecker(ExistsChecker, ABC):
    def __init__(self, attr_name, words: tuple, required: bool = True):
        super().__init__(attr_name, required)
        self.words = words

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        value = self.raw_value(data)

        if value is None:
            return CheckResult.make_success()

        if not isinstance(value, str):
            msg_format = "String exclude words checker: invalid value type '{}' of {}"
            return CheckResult.make_failed(msg_format.format(type(value), self.name()))

        contain_word = any(word in value for word in self.words)
        if not contain_word:
            return CheckResult.make_success()
        msg_format = "String exclude words checker: contain special word {}"
        return CheckResult.make_failed(msg_format.format(self.name()))
