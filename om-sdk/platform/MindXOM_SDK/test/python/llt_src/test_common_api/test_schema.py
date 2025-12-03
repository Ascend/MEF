# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import pytest

from common.schema import BaseModel, field


def my_function():
    return "Hello, World!"


class TestErrorCode:
    def test_to_dict(self):
        with pytest.raises(TypeError):
            BaseModel.to_dict(BaseModel())

    def test_from_dict_first_raise(self):
        with pytest.raises(TypeError):
            BaseModel.from_dict({"166": 1})

    def test_field(self):
        assert field(default=my_function)
