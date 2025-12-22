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
from pytest_mock import MockerFixture

from devm.device_driver import DeviceDriver
from devm.exception import DriverError
from ut_utils.mock_utils import mock_cdll


class TestDeviceDriver:
    DeviceLoadCase = namedtuple("DeviceLoadCase", ["expect", "so_file_name"])
    DeviceOpenCase = namedtuple("DeviceOpenCase", ["expect", "so_file_name", "device_name", "fd"])
    use_cases = {
        "test_load": {
            "ok": (None, "/path/to/libabc.so")
        },
        "test_unload": {
            "ok": (None, "/path/to/libabc.so")
        },
        "test_open_ok": {
            "ok": (0, "/path/to/libabc.so", "device_abc", 1),
        },
        "test_open_err": {
            "err": (DriverError, "/path/to/libabc.so", "device_abc", 1),
        },
        "test_close_ok": {
            "ok": (0, "/path/to/libabc.so", "device_abc", 1)
        },
        "test_close_err": {
            "err": (DriverError, "/path/to/libabc.so", "device_abc", 1)
        },
    }

    def test_load(self, model: DeviceLoadCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            assert driver_inst.load() == model.expect

    def test_unload(self, model: DeviceLoadCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            driver_inst.load()
            assert driver_inst.unload() == model.expect

    def test_open_ok(self, model: DeviceOpenCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            driver_inst.load()
            res = driver_inst.open(model.device_name)
            driver_inst.close(res)
            assert res == model.expect

    def test_open_err(self, mocker: MockerFixture, model: DeviceOpenCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            driver_inst.load()
            mocker.patch.object(DeviceDriver, "get_func").return_value.return_value = 1
            with pytest.raises(model.expect):
                res = driver_inst.open(model.device_name)
                driver_inst.close(res)

    def test_close_ok(self, model: DeviceOpenCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            driver_inst.load()
            res = driver_inst.open(model.device_name)
            assert driver_inst.close(res) == model.expect

    def test_close_err(self, mocker: MockerFixture, model: DeviceOpenCase):
        with mock_cdll():
            driver_inst = DeviceDriver(model.so_file_name)
            driver_inst.load()
            res = driver_inst.open(model.device_name)
            mocker.patch.object(DeviceDriver, "get_func").return_value.return_value = 1
            with pytest.raises(model.expect):
                driver_inst.close(res)
