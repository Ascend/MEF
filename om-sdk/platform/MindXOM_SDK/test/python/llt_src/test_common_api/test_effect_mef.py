# -*- coding: UTF-8 -*-
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

from pytest_mock import MockerFixture

from common.file_utils import FileUtils
from common.utils.exec_cmd import ExecCmd
from common.utils.system_utils import SystemUtils
from common.utils.result_base import Result
from effect_mef import effect_mef

EffectMefCase = namedtuple(
    "EffectMefCase", "expected, get_model_by_npu_smi, init, is_allow, check_script_file_valid, exec_cmd")


class TestEffectMef:
    use_cases = {
        "test_effect_mef": {
            "normal": (0, "Atlas A500 A2", range(10), (1, 1), Result(result=True), 0),
            "get_model_by_npu_smi_not_in_a50": (1, "/", 1, True, Result(result=True), 0),
            "ability_policy.init_is_err": (1, "Atlas A500 A2", Exception(), True, Result(result=True), 0),
            "ability_policy.is_allow_is_false": (1, "Atlas A500 A2", range(10), (0, 0), Result(result=True), 0),
            "check_script_file_valid_is_false": (1, "Atlas A500 A2", range(10), (1, 1), Result(result=False), 0),
            "exec_cmd_is_not_0": (1, "Atlas A500 A2", range(10), (1, 1), Result(result=True), 1),
        },
    }

    def test_effect_mef(self, mocker: MockerFixture, model: EffectMefCase):
        mocker.patch.object(SystemUtils, "get_model_by_npu_smi", return_value=model.get_model_by_npu_smi)
        mocker.patch("common.utils.ability_policy.init", side_effect=model.init)
        mocker.patch("common.utils.ability_policy.is_allow", side_effect=model.is_allow)
        mocker.patch.object(FileUtils, "check_script_file_valid", return_value=model.check_script_file_valid)
        mocker.patch.object(ExecCmd, "exec_cmd", return_value=model.exec_cmd)
        ret = effect_mef()
        assert model.expected == ret
