# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
