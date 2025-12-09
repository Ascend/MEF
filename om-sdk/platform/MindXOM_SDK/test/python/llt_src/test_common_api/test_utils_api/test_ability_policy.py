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
from unittest import mock
from unittest.mock import mock_open

import pytest
from pytest_mock import MockerFixture

from common.utils.ability_policy import init, HighRiskOpPolicyDto


class TestGetJsonInfoObj:
    def test_init_first(self):
        with pytest.raises(Exception):
            init("/home/data/config/ability_policy.json")

    @mock.patch("json.load", mock.Mock(return_value={'esp_enable': True}))
    def test_init_second(self, mocker: MockerFixture):
        mocker.patch("builtins.open", mock_open(read_data="TEST"))
        mocker.patch.object(HighRiskOpPolicyDto, "load_from_json").side_effect = Exception()
        with pytest.raises(Exception):
            init("/home/data/config/ability_policy.json")
