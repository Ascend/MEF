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

from common.utils.ability_policy import HighRiskOpPolicyDto
from om_ability_policy import cli_set_config, main


class TestOmAbilityPolicy:
    MainCase = namedtuple("MainCase", "expected, cli_set_config")

    use_cases = {
        "test_main": {
            "normal": (0, 0),
        },
        "test_main_help_exception": {
            "Exception": (1, 1),
        },
    }

    def test_cli_set_config(self, mocker: MockerFixture):
        mocker.patch("om_ability_policy.check_param", return_value=None)
        mocker.patch("om_ability_policy.OmCli.parse_args", return_value=HighRiskOpPolicyDto)
        mocker.patch.object(HighRiskOpPolicyDto, "dump_to_json", return_value=None)
        mocker.patch("json.dumps", return_value=None)
        mocker.patch("os.fdopen").return_value.__enter__.return_value.read.side_effect = "abc"
        ret = cli_set_config("/home/data/config/ability_policy.json")
        assert ret == 0

    def test_main(self, mocker: MockerFixture, model: MainCase):
        mocker.patch("om_ability_policy.cli_set_config", return_value=model.cli_set_config)
        ret = main()
        assert model.expected == ret

    def test_main_help_exception(self, model: MainCase):
        ret = main()
        assert model.expected == ret
