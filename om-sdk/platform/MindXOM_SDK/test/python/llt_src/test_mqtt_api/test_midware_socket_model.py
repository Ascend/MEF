# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from pytest_mock import MockerFixture

from common.utils.app_common_method import AppCommonMethod
from fd_msg_process.midware_socket_model import MidWareSocketModel


def some_function():
    return "some result"


class TestGetJsonInfoObj:
    def test_add_url_rule(self):
        assert not MidWareSocketModel.check_socket_model({"header": "", "route": ""})

    def test_add_url_rule_second(self):
        assert MidWareSocketModel.check_socket_model({"header": "1", "route": "1"})

    def test_get_socket_model(self):
        assert MidWareSocketModel.get_socket_model({"header": "1", "route": "1", "content": ""})

    def test_get_socket_info(self, mocker: MockerFixture):
        mocker.patch.object(AppCommonMethod, "get_json_info", side_effect={"header": some_function(), "route": ""})
        assert MidWareSocketModel.get_socket_info(MidWareSocketModel)
