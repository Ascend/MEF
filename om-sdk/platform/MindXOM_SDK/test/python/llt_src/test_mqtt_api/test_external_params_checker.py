# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple

from net_manager.checkers.external_params_checker import judge_pwd_by_account, NetManageConfigChecker


class TestUtils:
    @staticmethod
    def raw_value():
        return True


class TestGetJsonInfoObj:
    CheckJudgePwdByAccountCase = namedtuple("CheckJudgePwdByAccountCase", "expected, account")
    CheckConfigParamsCase = namedtuple("CheckConfigParamsCase", "expected, net_manager_params")
    CheckDictCase = namedtuple("CheckDictCase", "tested_class, param_key, input_param, excepted_success")
    use_cases = {
        "test_judge_pwd_by_account": {
            "False": (False, ""),
            "True": (True, "1"),
        },
        "test_check_config_params": {
            "True": (True, {"ManagerType": "Web"}),
            "False": (False, {"ManagerType": "FusionDirector"}),
            "False1": (False, {"ManagerType": ""}),
        },
        "test_check_dict": {
            "not_exists_checker_is_existed": (NetManageConfigChecker, "1", {"1": {"ManagerType": "FusionDirector"}},
                                              False),
            "not net_manager": (NetManageConfigChecker, None, {None: {"ManagerType": None}}, False),
            "not_exists_checker_not_existed": (NetManageConfigChecker, "2", {"1": {"ManagerType": "FusionDirector"}},
                                               False),
        },
    }

    def test_judge_pwd_by_account(self, model: CheckJudgePwdByAccountCase):
        assert judge_pwd_by_account("", model.account) == model.expected

    def test_check_config_params(self, model: CheckConfigParamsCase):
        assert bool(NetManageConfigChecker.check_config_params(model.net_manager_params)) == model.expected

    def test_check_dict(self, model: CheckDictCase):
        ret = model.tested_class(model.param_key).check_dict(model.input_param)
        assert model.excepted_success == ret.success
