# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from collections import namedtuple

import pytest

from common.utils.timer import RepeatingTimer, DynamicRepeatingTimer


def foo(a_b, b_c):
    return a_b + b_c


class TestUtils1:
    @staticmethod
    def is_set():
        params = {'a': 0, 'b': 0}
        return foo(**params)

    @staticmethod
    def wait(a_b):
        raise Exception("Stop the loop")

    @staticmethod
    def is_set2():
        return {'a': 1, 'b': 2}

    @staticmethod
    def is_set3():
        return {'c': 0, 'd': 0}

    @staticmethod
    def is_set1(a_b, b_b, c_c, d_d):
        return foo(a_b, b_b)


class TestGetJsonInfoObj:
    def test_run(self):
        CopyFile1 = namedtuple("CopyFile1", ["finished", "function", "args", "kwargs", "interval"])
        backup_restore_base_two = CopyFile1(*(TestUtils1, TestUtils1.is_set1, TestUtils1.is_set2(),
                                              TestUtils1.is_set3(), TestUtils1.is_set3(), ))
        with pytest.raises(Exception):
            RepeatingTimer.run(backup_restore_base_two)

    def test_run1_except(self):
        CopyFile1 = namedtuple("CopyFile1", ["finished", "name", ])
        backup_restore_base_two = CopyFile1(*(TestUtils1, ""))
        with pytest.raises(Exception):
            DynamicRepeatingTimer.run(backup_restore_base_two)

    def test_run1(self):
        CopyFile1 = namedtuple("CopyFile1", ["finished", "name", "function", "args", "kwargs", "retry_cnt"])
        backup_restore_base_two = CopyFile1(*(TestUtils1, "", TestUtils1.is_set1, TestUtils1.is_set2(),
                                              TestUtils1.is_set3(), 5))
        with pytest.raises(Exception):
            DynamicRepeatingTimer.run(backup_restore_base_two)
