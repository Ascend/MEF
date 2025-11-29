# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from pytest_mock import MockerFixture

from fd_msg_process.fd_configs import get_msg_handling_mapping


class TestUtils:
    a = [1, ]


class TestGetJsonInfoObj:
    def test_publish_ws_msg(self):
        get_msg_handling_mapping("")

    def test_publish_ws_msg_exception(self):
        get_msg_handling_mapping("a")

    def test_publish_ws_msg1(self, mocker: MockerFixture):
        mocker.patch("importlib.import_module", return_value=TestUtils)
        get_msg_handling_mapping("a")
