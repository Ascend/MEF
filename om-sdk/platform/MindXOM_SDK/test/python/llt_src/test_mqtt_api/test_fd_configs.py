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
