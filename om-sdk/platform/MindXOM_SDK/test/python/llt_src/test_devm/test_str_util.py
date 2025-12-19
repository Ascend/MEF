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
from collections import namedtuple

import pytest

from devm.device_attrs import StrUtil
from devm.exception import DeviceManagerError


class TestStrUtil:
    StrToBytesCase = namedtuple("StrToBytesCase", ["expect", "data"])
    use_cases = {
        "test_str_to_bytes_normal": {
            "input_str": (b"\x61\x62\x63", "abc"),
            "input_bytes": (b"\x61\x62\x63", b"\x61\x62\x63")
        },
        "test_str_to_bytes_exception": {
            "input_int": (DeviceManagerError, 1),
            "input_float": (DeviceManagerError, 1.1),
            "input_bool": (DeviceManagerError, True),
            "input_dict": (DeviceManagerError, {"abc": 123})
        },
        "test_load_from_bytes": {
            "byte_data_end_with_null_char": ("abc", b"\x61\x62\x63\x00\x00"),
            "byte_data_end_without_null_char": ("", b"\x61\x62\x63")
        }
    }

    def test_str_to_bytes_normal(self, model: StrToBytesCase):
        assert StrUtil.str_to_bytes(model.data) == model.expect

    def test_str_to_bytes_exception(self, model: StrToBytesCase):
        with pytest.raises(model.expect):
            StrUtil.str_to_bytes(model.data)

    def test_load_from_bytes(self, model: StrToBytesCase):
        assert StrUtil.load_from_bytes(model.data) == model.expect
