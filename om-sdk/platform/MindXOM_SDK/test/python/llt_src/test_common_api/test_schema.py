# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
