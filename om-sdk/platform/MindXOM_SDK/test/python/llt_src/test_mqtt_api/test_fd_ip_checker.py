# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple

from common.checkers import LocalIpChecker


class TestUtils:
    @staticmethod
    def raw_value():
        return True


class TestGetJsonInfoObj:
    CheckLocalIpCase = namedtuple("CheckLocalIpCase", "tested_class, param_key, input_param, excepted_success")
    CheckDictCase = namedtuple("CheckDictCase", "tested_class, param_key, input_param, excepted_success")
    use_cases = {
        "test_check_local_ip": {
            "not_exists_checker_is_existed": (LocalIpChecker, "1", '51.38.69.401', True),
            "not_exists_checker_not_existed": (LocalIpChecker, "2", "127.0.0.1", False),
        },
        "test_check_dict": {
            "not_exists_checker_is_existed": (LocalIpChecker, "1", {"1": "1"}, True),
            "not_exists_checker_not_existed": (LocalIpChecker, "2", {"1": "1"}, False),
            "not net_manager": (LocalIpChecker, None, {}, False),
            "not value str": (LocalIpChecker, None, {None: {"ManagerType": None}}, False),
        },
    }

    def test_get_ip_address(self):
        assert LocalIpChecker.get_ip_address(LocalIpChecker, b"eth0")

    def test_check_local_ip(self, model: CheckLocalIpCase):
        ret = model.tested_class(model.param_key).check_local_ip(model.input_param)
        assert model.excepted_success == ret.success

    def test_check_dict(self, model: CheckDictCase):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
