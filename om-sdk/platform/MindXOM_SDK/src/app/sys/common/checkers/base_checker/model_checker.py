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
from abc import ABC
from abc import ABCMeta
from typing import Tuple, Sequence

from common.checkers.base_checker.basic_attr_checker import _ListChecker
from common.checkers.base_checker.basic_attr_checker import is_in
from common.checkers.base_checker.abc_checker import AttrCheckerBase, AttrCheckerInterface
from common.checkers.base_checker.abc_checker import CheckResult
from common.checkers.base_checker.abc_checker import DefaultAttrCheck


class _BaseModelChecker(AttrCheckerBase, ABC):
    def __init__(self, attr_name: str, required: bool = True):
        super().__init__(attr_name)
        self._required = required
        self.fields: Sequence[AttrCheckerInterface] = []

    @staticmethod
    def _check(data, field):
        ret = CheckResult.make_success()
        if isinstance(field, BaseModelChecker):
            if field.required() or data.get(field.name()) is not None:
                if data.get(field.name()) is None:
                    ret = CheckResult.make_failed("Base model checker: {} not exists".format(field.name()))
                else:
                    ret = field.check(data.get(field.name()))
        else:
            ret = field.check(data)
        return ret

    def required(self) -> bool:
        return self._required

    def get_field(self, field_name: str):
        for field in self.fields:
            if field.name() == field_name:
                return field
        return self._get_field_by_is_in_list_checker(field_name)

    def check_dict(self, data: dict) -> CheckResult:
        if not self.required() and not data:
            return CheckResult.make_success()
        if data is None:
            data = {}
        for field in self.fields:
            ret = self._check(data, field)
            if not ret.success:
                return ret
        return CheckResult.make_success()

    def _get_field_by_is_in_list_checker(self, field_name):
        for field in self.fields:
            if isinstance(field, _ListChecker):
                if is_in(field.checkers, field_name):
                    return field

        return DefaultAttrCheck()


class BaseModelChecker(_BaseModelChecker):
    def __init__(self, attr_name: str = "", required: bool = True):
        super().__init__(attr_name, required)
        # base_fields 在这个类外定义，fields为Tuple, 用于提醒使用者：不要修改
        # noinspection PyUnresolvedReferences
        self.fields: Tuple[AttrCheckerInterface] = tuple(self.base_fields)


class ModelCheckerMetaclass(ABCMeta):
    def __new__(mcs, name, bases, attrs):
        new_class = super(ModelCheckerMetaclass, mcs).__new__(mcs, name, bases, attrs)
        if new_class == (BaseModelChecker,):
            return new_class

        if name == "ModelChecker":
            return new_class

        opts = new_class._meta = getattr(new_class, "Meta", None)
        fields = getattr(opts, "fields")
        if isinstance(fields, AttrCheckerInterface):
            fields = (fields,)

        if not isinstance(fields, tuple):
            raise Exception("Creating CheckerModel 'fields' must be tuple")

        new_class.base_fields = fields
        return new_class


class ModelChecker(BaseModelChecker, metaclass=ModelCheckerMetaclass):
    pass
